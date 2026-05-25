# Annual Statement Generator

A Go-based application for generating and syncing annual financial statements for IFA (Independent Financial Advisor) clients. The application retrieves investor statement data from a MySQL database and synchronizes PDF files between AWS S3 buckets.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Supported Clients](#supported-clients)
- [Project Structure](#project-structure)
- [Docker Deployment](#docker-deployment)
- [Development](#development)
- [License](#license)

## 🎯 Overview

This application automates the process of generating and distributing annual financial statements for investors managed by various IFA clients. It:

1. Retrieves statement records from a MySQL database for a specific IFA and financial year
2. Downloads PDF statements from a source S3 bucket (LOS)
3. Uploads the statements to a destination S3 bucket (SFTP) organized by client and financial year
4. Processes statements concurrently for improved performance

## ✨ Features

- **Multi-IFA Support**: Handles statements for multiple IFA clients
- **Hierarchical IFA Relationships**: Retrieves statements for an IFA and all child IFAs (up to 5 levels deep)
- **Concurrent Processing**: Processes up to 10 files simultaneously for faster execution
- **S3 Integration**: Seamless download and upload between AWS S3 buckets
- **Financial Year Calculation**: Automatically calculates financial year based on Indian fiscal calendar (April-March)
- **Database Integration**: Uses MySQL with the Upper.io ORM for efficient data retrieval
- **Error Tracking**: Reports success and failure counts for file operations

## 🏗 Architecture

### Components

1. **Main Package** (`main.go`): Entry point, argument parsing, and orchestration
2. **Common Package** (`common/`): Database connections, S3 operations, and file synchronization
3. **Data Package** (`data/`): Database models and query logic

### Data Flow

```
MySQL Database → Retrieve Statement Records → Download PDFs from LOS S3 
→ Upload to SFTP S3 (organized by client/FY) → Report Results
```

## 📦 Prerequisites

- **Go**: Version 1.21.4 or higher
- **MySQL**: Access to the investor database
- **AWS Account**: With S3 bucket access (LOS and SFTP buckets)
- **AWS Credentials**: IAM credentials with S3 read/write permissions
- **Docker** (optional): For containerized deployment

## 🚀 Installation

### Local Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/kurmidev/annualstatement.git
   cd annualstatement
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables** (see [Configuration](#configuration))

4. **Build the application**
   ```bash
   go build -o bin .
   ```

### Docker Installation

```bash
docker build -t annual-statement-generator .
docker run --env-file .env annual-statement-generator ./bin <IFA_ID> <START_DATE>
```

## ⚙️ Configuration

Create a `.env` file in the root directory with the following variables:

```env
# Database Configuration
DB_HOST=your-mysql-host
DB_NAME=your-database-name
DB_USER=your-database-user
DB_PASSWORD=your-database-password

# AWS S3 Configuration - LOS Bucket (Source)
AWS_REGION=ap-south-1
AWS_LOS_ID=your-los-access-key-id
AWS_LOS_SECRET=your-los-secret-access-key
LOS_BUCKET=your-los-bucket-name

# AWS S3 Configuration - SFTP Bucket (Destination)
AWS_SFTP_ID=your-sftp-access-key-id
AWS_SFTP_SECRET=your-sftp-secret-access-key
SFTP_BUCKET=your-sftp-bucket-name
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_HOST` | MySQL database host | `localhost:3306` |
| `DB_NAME` | Database name | `investor_db` |
| `DB_USER` | Database username | `admin` |
| `DB_PASSWORD` | Database password | `secret123` |
| `AWS_REGION` | AWS region for S3 | `ap-south-1` |
| `AWS_LOS_ID` | AWS Access Key for LOS bucket | `AKIAIOSFODNN7EXAMPLE` |
| `AWS_LOS_SECRET` | AWS Secret Key for LOS bucket | `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY` |
| `LOS_BUCKET` | Source S3 bucket name | `statements` |
| `AWS_SFTP_ID` | AWS Access Key for SFTP bucket | `AKIAIOSFODNN7EXAMPLE` |
| `AWS_SFTP_SECRET` | AWS Secret Key for SFTP bucket | `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY` |
| `SFTP_BUCKET` | Destination S3 bucket name | `sftp-statements` |

## 💻 Usage

### Command Line

```bash
./bin <IFA_ID> <FINANCIAL_YEAR_START_DATE>
```

### Parameters

- **IFA_ID**: Integer ID of the IFA client (see [Supported Clients](#supported-clients))
- **FINANCIAL_YEAR_START_DATE**: Start date of the financial year in `YYYY-MM-DD` format

### Examples

```bash
# Generate statements for Prudent (IFA ID: 1799) for FY 2023-24
./bin 1799 2023-04-01

# Generate statements for ET-Money-Supply (IFA ID: 3569) for FY 2024-25
./bin 3569 2024-04-01

# Generate statements for Uni (IFA ID: 2300) for FY 2022-23
./bin 2300 2022-04-01
```

### Output

The application will:
1. Print the total number of investor statements found
2. Download and upload each file
3. Display upload status for each file
4. Print final success and error counts

```
1799 2023-04-01 00:00:00 +0000 UTC
Total investor count  150
Searching file statement_001.pdf
Uploading file statement_001.pdf to new path /Prudent/Annual-statement/FY-23-24/statement_001.pdf
File uploaded successfully statement_001.pdf
...
current upload status success 148 and error is 2
```

## 📁 Project Structure

```
annualstatement/
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── go.sum                  # Dependency checksums
├── Dockerfile              # Docker configuration
├── .env                    # Environment variables (not in git)
├── .gitignore              # Git ignore rules
├── common/
│   └── common.go          # Database, S3, and sync logic
└── data/
    ├── model.go           # Database queries and IFA hierarchy
    └── statementlog.go    # Statement log data structure
```

### Package Descriptions

#### `main`
- Entry point of the application
- Parses command-line arguments
- Orchestrates the statement generation process
- Maps IFA IDs to client names

#### `common`
- **Database Connection**: Establishes MySQL connection using Upper.io ORM
- **S3 Operations**: Handles AWS S3 connections, downloads, and uploads
- **File Sync**: Concurrent file synchronization with semaphore-based limiting
- **Financial Year Calculation**: Converts dates to Indian FY format (e.g., "23-24")

#### `data`
- **IFA Hierarchy**: Retrieves child IFAs up to 5 levels deep
- **Statement Queries**: Fetches active annual statements for specified date range
- **Data Models**: Defines the `InvInvestorStatementLog` structure

## 🐳 Docker Deployment

### Build the Image

```bash
docker build -t annual-statement-generator .
```

### Run the Container

```bash
docker run --env-file .env annual-statement-generator ./bin 1799 2023-04-01
```

### Docker Compose (Optional)

Create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  annual-statement:
    build: .
    env_file:
      - .env
    command: ["./bin", "1799", "2023-04-01"]
```

Run with:
```bash
docker-compose up
```

## 🛠 Development

### Running Locally

```bash
go run main.go <IFA_ID> <START_DATE>
```

### Running Tests

```bash
go test ./...
```

### Code Formatting

```bash
go fmt ./...
```

### Linting

```bash
golangci-lint run
```

### Dependencies

The project uses the following key dependencies:

- **[AWS SDK for Go](https://github.com/aws/aws-sdk-go)**: AWS S3 operations
- **[Upper.io](https://github.com/upper/db)**: Database ORM
- **[MySQL Driver](https://github.com/go-sql-driver/mysql)**: MySQL connectivity
- **[godotenv](https://github.com/joho/godotenv)**: Environment variable loading

## 🔒 Security Considerations

- **Never commit `.env` file**: Contains sensitive credentials
- **Use IAM roles**: When running on AWS infrastructure, use IAM roles instead of hardcoded credentials
- **Least privilege**: Ensure AWS credentials have minimal required permissions
- **Network security**: Restrict database access to authorized IPs
- **Update dependencies**: Regularly update Go modules for security patches

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Verify database credentials in `.env`
   - Check network connectivity to MySQL host
   - Ensure database exists and user has proper permissions

2. **S3 Access Denied**
   - Verify AWS credentials are correct
   - Check IAM permissions for S3 bucket access
   - Ensure bucket names are correct

3. **File Not Found in S3**
   - Verify the file exists in the source bucket
   - Check the file path in the database matches S3 key

4. **Too Many Open Files**
   - The concurrent limit is set to 10 files
   - Increase system file descriptor limits if needed

## 📝 License

This project is proprietary software. All rights reserved.

## 👥 Contributors

- Development Team: kurmidev

## 📞 Support

For issues or questions, please contact the development team.

---

**Note**: This application handles sensitive financial data. Ensure proper security measures are in place before deployment.
