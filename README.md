# 🧪 Lambda S3 Dispatcher - PROOF OF CONCEPT

> ⚠️ **TEST PROJECT** - This is experimental code to demonstrate S3 + OCR API integration

## What does it do?

A simple dispatcher that:

1. **Reads images** from an S3 bucket 
2. **Generates presigned URLs** for each image
3. **Sends URLs** to an external OCR API
4. **Collects and structures** the responses

## Basic flow

```
S3 Bucket → List files → Presigned URLs → OCR API → JSON Results
```

## Required environment variables

```bash
export BUCKET_NAME="your-s3-bucket"
export API_URL="http://localhost:8080"  # Your OCR API URL
export AWS_REGION="us-east-1"           # Bucket region
```

## Optional variables

```bash
export PREFIX="images/"                 # Filter files by prefix
export MAX_CONCURRENCY="10"            # Parallel request limit (default: 10)
export BATCH_SIZE="50"                  # Batch processing size (default: 50)
export USE_BATCH="true"                 # Enable batch mode (default: false)
export AWS_PROFILE="your-profile"      # Specific AWS profile
```

## Usage

### Local test execution
```bash
go run .
```

### As Lambda (comment testHandler, uncomment lambda.Start)
```go
func main() {
    // testHandler()  // ← comment this line
    lambda.Start(Handler)  // ← uncomment this line
}
```

## Processing Modes

The dispatcher supports two processing modes:

### Individual Mode (Default: `USE_BATCH=false`)
- Each file is processed individually with concurrent requests
- Uses `golang.org/x/sync/errgroup` for optimized concurrency control
- Built-in concurrency limiting with `MAX_CONCURRENCY` (default: 10)
- Improved error handling and context cancellation
- Calls `POST /ocr` endpoint for each file
- Compatible with existing OCR APIs

### Batch Mode (`USE_BATCH=true`)
- Files are processed in sequential batches of size `BATCH_SIZE` (default: 50)
- Each batch is sent as a single API request to `/ocr/batch` endpoint
- Reduces API overhead and provides better resource management for large datasets
- Requires API to support batch processing endpoint

## Response structure

```json
{
  "bucket": "test-images-bucket",
  "prefix": "images/",
  "processed": 30,
  "api_responses": [
    {
      "key": "images/photo1.png",
      "status_code": 200,
      "result": {
        "key": "images/photo1.png",
        "source_url": "https://...",
        "full_text": "Extracted text from image"
      }
    }
  ],
  "errors": []
}
```

## Requirements

- **Go 1.19+**
- **AWS credentials** configured
- **OCR API** running that accepts:

**Individual Mode:** `POST /ocr` with format:
  ```json
  {"key": "file.png", "url": "https://presigned-url..."}
  ```

**Batch Mode:** `POST /ocr/batch` with format:
  ```json
  {
    "items": [
      {"key": "file1.png", "url": "https://presigned-url1..."},
      {"key": "file2.png", "url": "https://presigned-url2..."}
    ]
  }
  ```

  And returns:
  ```json
  {
    "results": [
      {
        "key": "file1.png",
        "status_code": 200,
        "result": {"key": "file1.png", "source_url": "https://...", "full_text": "..."}
      },
      {
        "key": "file2.png", 
        "status_code": 200,
        "result": {"key": "file2.png", "source_url": "https://...", "full_text": "..."}
      }
    ]
  }
  ```

## ⚠️ Important note

This is **test code** to validate the architecture. For production you would need:

- More robust error handling
- Structured logging
- Metrics and monitoring
- Configurable timeouts
- Rate limiting
- Unit tests

---

*Developed as POC for S3 + OCR API integration*