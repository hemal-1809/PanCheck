import { useState } from 'react';
import { LinkInput } from '@/components/LinkInput';
import { ResultTable } from '@/components/ResultTable';
import { linkApi } from '@/api/linkApi';
import { toast } from 'sonner';
import type { CheckLinksResponse } from '@/types';

export function Home() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<CheckLinksResponse | null>(null);

  const handleCheck = async (links: string[], selectedPlatforms?: string[]) => {
    setLoading(true);
    setResult(null);

    try {
      const requestData: any = { links };
      if (selectedPlatforms && selectedPlatforms.length > 0) {
        requestData.selected_platforms = selectedPlatforms;
      }
      const response = await linkApi.checkLinks(requestData);
      setResult(response);
    } catch (error: any) {
      toast.error('检测失败: ' + (error.response?.data?.error || error.message));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto py-8 space-y-8">
        <div className="text-center space-y-2">
          <h1 className="text-4xl font-bold">网盘链接检查工具</h1>
          <p className="text-muted-foreground">
            支持夸克、UC、百度、天翼、123、115、阿里云、迅雷、移动云盘等9种网盘平台
          </p>
        </div>

        <LinkInput onCheck={handleCheck} loading={loading} />

        {result && (
          <ResultTable
            invalidLinks={result.invalid_links}
            pendingLinks={result.pending_links}
            validLinks={result.valid_links}
            totalDuration={result.total_duration}
            invalidFormatCount={result.invalid_format_count}
            duplicateCount={result.duplicate_count}
          />
        )}
      </div>
    </div>
  );
}

