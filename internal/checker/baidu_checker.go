package checker

import (
	"PanCheck/internal/model"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// BaiduChecker 百度网盘检测器
type BaiduChecker struct {
	*BaseChecker
}

// NewBaiduChecker 创建百度网盘检测器
func NewBaiduChecker(concurrencyLimit int, timeout time.Duration) *BaiduChecker {
	return &BaiduChecker{
		BaseChecker: NewBaseChecker(model.PlatformBaidu, concurrencyLimit, timeout),
	}
}

// normalizeBaiduURL 规范化百度网盘URL，提取有效部分并进行编码
func normalizeBaiduURL(link string) (string, error) {
	cleaned := strings.TrimSpace(link)

	// 找到 https://pan.baidu.com/s/ 的位置
	startIdx := strings.Index(cleaned, "https://pan.baidu.com/s/")
	if startIdx == -1 {
		startIdx = strings.Index(cleaned, "http://pan.baidu.com/s/")
	}
	if startIdx == -1 {
		return "", fmt.Errorf("未找到有效的百度网盘URL")
	}

	// 从URL起始位置开始，找到URL结束位置（第一个空格、换行或"提取码"等关键词）
	endIdx := startIdx
	for endIdx < len(cleaned) {
		char := cleaned[endIdx]
		// 遇到空白字符，停止
		if char == ' ' || char == '\n' || char == '\r' || char == '\t' {
			break
		}
		// 检查是否遇到"提取码"等关键词
		remaining := cleaned[endIdx:]
		if strings.HasPrefix(remaining, "提取码") || strings.HasPrefix(remaining, "密码") {
			break
		}
		endIdx++
	}

	// 提取URL部分（不包含后面的额外文本）
	urlStr := cleaned[startIdx:endIdx]
	urlStr = strings.TrimSpace(urlStr)

	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %v", err)
	}

	// 重新构建URL，确保查询参数正确编码
	// 这样即使原始URL包含未编码的字符，也会被正确编码
	normalized := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	if parsedURL.RawQuery != "" {
		normalized += "?" + parsedURL.Query().Encode()
	}
	if parsedURL.Fragment != "" {
		normalized += "#" + parsedURL.Fragment
	}

	return normalized, nil
}

// Check 检测链接是否有效
func (c *BaiduChecker) Check(link string) (*CheckResult, error) {
	// 规范化URL（提取有效部分并进行编码）
	normalizedLink, err := normalizeBaiduURL(link)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "URL规范化失败: " + err.Error(),
			Duration:      0,
		}, nil
	}

	// 应用频率限制
	c.ApplyRateLimit()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.GetTimeout())
	defer cancel()

	parsedURL, err := url.Parse(normalizedLink)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "解析URL失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	queryParams := parsedURL.Query()
	password := queryParams.Get("pwd")

	// 第一步：初始请求
	step1Result, err := step1Request(ctx, normalizedLink)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "第一步请求失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 过期 200
	if step1Result.StatusCode == 200 && step1Result.FullRedirectURL == "" {
		return &CheckResult{
			Valid:         false,
			FailureReason: "分享文件已过期",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 正常 302
	if step1Result.StatusCode != 302 || step1Result.FullRedirectURL == "" || step1Result.SURL == "" {
		return &CheckResult{
			Valid:         false,
			FailureReason: "第一步302失败",
			Duration:      time.Since(start).Milliseconds(),
			IsRateLimited: true, // 第一步302失败表示被平台限制
		}, nil
	}

	// 第二步：验证请求
	step2Result, err := step2Request(ctx, step1Result, password)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "第二步请求失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 如果响应包含JSON，根据errno判断错误类型
	if step2Result.JSONResponse != nil {
		errno := step2Result.JSONResponse.Errno
		errMsg := step2Result.JSONResponse.ErrMsg

		switch errno {
		case -12:
			// 缺少提取码
			return &CheckResult{
				Valid:         false,
				FailureReason: fmt.Sprintf("缺少提取码 (errno: %d, err_msg: %s)", errno, errMsg),
				Duration:      time.Since(start).Milliseconds(),
				IsRateLimited: false,
			}, nil
		case -9:
			// 提取码错误
			return &CheckResult{
				Valid:         false,
				FailureReason: fmt.Sprintf("提取码错误 (errno: %d, err_msg: %s)", errno, errMsg),
				Duration:      time.Since(start).Milliseconds(),
				IsRateLimited: false,
			}, nil
		case -62:
			// 请求接口受限
			return &CheckResult{
				Valid:         false,
				FailureReason: fmt.Sprintf("请求接口受限 (errno: %d, err_msg: %s)", errno, errMsg),
				Duration:      time.Since(start).Milliseconds(),
				IsRateLimited: true,
			}, nil
		case 0:
			// errno为0表示成功，继续检查BDCLND Cookie
		default:
			// 其他错误码
			return &CheckResult{
				Valid:         false,
				FailureReason: fmt.Sprintf("第二步验证失败 (errno: %d, err_msg: %s)", errno, errMsg),
				Duration:      time.Since(start).Milliseconds(),
				IsRateLimited: false,
			}, nil
		}
	}

	// 如果errno为0或未解析到JSON，检查BDCLND Cookie
	if step2Result.BDCLND == "" {
		failureReason := fmt.Sprintf("第二步响应未返回BDCLND Cookie (StatusCode: %d, Response: %s)", step2Result.StatusCode, step2Result.Body)
		return &CheckResult{
			Valid:         false,
			FailureReason: failureReason,
			Duration:      time.Since(start).Milliseconds(),
			IsRateLimited: true, // 第二步响应未返回BDCLND Cookie表示被平台限制
		}, nil
	}

	return &CheckResult{
		Valid:         true,
		FailureReason: "",
		Duration:      time.Since(start).Milliseconds(),
	}, nil
}

// Step1Response 第一步响应结构体
type Step1Response struct {
	StatusCode      int
	Location        string
	FullRedirectURL string
	SetCookies      []*http.Cookie
	SURL            string
}

// Step2Response 第二步响应结构体
type Step2Response struct {
	StatusCode   int
	SetCookies   []*http.Cookie
	BDCLND       string
	Body         string             // 响应体内容
	JSONResponse *Step2JSONResponse // 解析后的JSON响应
}

// Step2JSONResponse 第二步JSON响应结构体
type Step2JSONResponse struct {
	Errno     int    `json:"errno"`
	ErrMsg    string `json:"err_msg"`
	RequestID int64  `json:"request_id"`
}

// Step3Response 第三步响应结构体
type Step3Response struct {
	JSONResponse *ShareListResponse
}

// ShareListResponse 第三步JSON响应结构体
type ShareListResponse struct {
	Errno int    `json:"errno"`
	Title string `json:"title"`
}

// step1Request 第一步请求
func step1Request(ctx context.Context, targetURL string) (*Step1Response, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	setStep1Headers(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	return parseStep1Response(resp, targetURL)
}

// setStep1Headers 设置第一步请求头
func setStep1Headers(req *http.Request) {
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language":           "en",
		"Connection":                "keep-alive",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// parseStep1Response 解析第一步响应
func parseStep1Response(resp *http.Response, originalURL string) (*Step1Response, error) {
	result := &Step1Response{
		StatusCode: resp.StatusCode,
		Location:   resp.Header.Get("Location"),
		SetCookies: resp.Cookies(),
	}

	if result.Location != "" {
		fullURL, err := buildFullRedirectURL(originalURL, result.Location)
		if err != nil {
			return nil, fmt.Errorf("构建重定向URL失败: %v", err)
		}
		result.FullRedirectURL = fullURL

		if surl, err := extractSURLFromLocation(result.Location); err == nil {
			result.SURL = surl
		}
	}

	return result, nil
}

// step2Request 第二步请求
func step2Request(ctx context.Context, step1Result *Step1Response, password string) (*Step2Response, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建Cookie Jar失败: %v", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	baseURL := "https://pan.baidu.com/share/verify"
	params := url.Values{}
	params.Add("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Add("surl", step1Result.SURL)
	params.Add("channel", "chunlei")
	params.Add("web", "1")
	params.Add("app_id", "250528")
	params.Add("clienttype", "0")

	fullURL := baseURL + "?" + params.Encode()

	postData := url.Values{}
	postData.Add("pwd", password)
	postData.Add("vcode", "")
	postData.Add("vcode_str", "")

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBufferString(postData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建第二步请求失败: %v", err)
	}

	setStep2Headers(req, step1Result.FullRedirectURL)

	u, _ := url.Parse("https://pan.baidu.com")
	jar.SetCookies(u, step1Result.SetCookies)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("第二步请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取第二步响应体失败: %v", err)
	}

	return parseStep2Response(resp, string(body))
}

// setStep2Headers 设置第二步请求头
func setStep2Headers(req *http.Request, refererURL string) {
	headers := map[string]string{
		"Accept":           "application/json, text/javascript, */*; q=0.01",
		"Accept-Language":  "en",
		"Connection":       "keep-alive",
		"Content-Type":     "application/x-www-form-urlencoded; charset=UTF-8",
		"Origin":           "https://pan.baidu.com",
		"Referer":          refererURL,
		"Sec-Fetch-Dest":   "empty",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Site":   "same-origin",
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		"X-Requested-With": "XMLHttpRequest",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// parseStep2Response 解析第二步响应
func parseStep2Response(resp *http.Response, body string) (*Step2Response, error) {
	result := &Step2Response{
		StatusCode: resp.StatusCode,
		SetCookies: resp.Cookies(),
		Body:       body,
	}

	// 解析JSON响应
	var jsonResp Step2JSONResponse
	if err := json.Unmarshal([]byte(body), &jsonResp); err == nil {
		result.JSONResponse = &jsonResp
	}

	for _, cookie := range result.SetCookies {
		if cookie.Name == "BDCLND" {
			result.BDCLND = cookie.Value
			break
		}
	}

	return result, nil
}

// buildFullRedirectURL 构建完整的重定向URL
func buildFullRedirectURL(baseURL, location string) (string, error) {
	if location == "" {
		return "", fmt.Errorf("location为空")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	redirect, err := url.Parse(location)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(redirect).String(), nil
}

// extractSURLFromLocation 从Location中提取surl参数
func extractSURLFromLocation(location string) (string, error) {
	parsedURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}

	queryParams, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		return "", err
	}

	surl := queryParams.Get("surl")
	if surl == "" {
		return "", fmt.Errorf("未找到surl参数")
	}

	return surl, nil
}
