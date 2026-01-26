## What IS This
MicroSms is a stupid idea I have, create a webservice that creates SMS requests and partner it with a spare Android phone/SIM. The Phone will then query the server and grab any pending requests (one by one by visiting a URL), then it will autocreate an SMS using the built in Android API. This Phone will run a custom built Android App that will do this on a timer.

### So what is THIS
This is the webserver, it's going to support the following scenarios

# MicroSMS Server

A lightweight Go-based web server that manages SMS requests and integrates with the SMSFilter service for content filtering. The server acts as the central coordinator between API requests and the Android worker that sends the actual SMS messages.

# Autogen'd README (With Hand Touch Notes)
## Overview

MicroSMS Server provides a RESTful API for creating, managing, and retrieving SMS requests. It implements a workflow where messages are:
1. Created via API
2. Filtered through the SMSFilter service for safety
3. Queued for sending by the Android worker
4. Tracked through completion

## Features

- **REST API**: Full CRUD operations for SMS requests
- **Content Filtering**: Automatic integration with SMSFilter service to block unsafe content
- **Concurrent Processing**: Configurable concurrent filter API requests with throttling
- **SQLite Database**: Persistent storage with GORM ORM
- **Status Tracking**: Track messages through their lifecycle (payment_owed → ready_to_send → taken → sent)
- **Configuration**: YAML-based configuration with environment variable overrides
- **Docker Support**: Containerized deployment with volume persistence

## Architecture

```
Client → MicroSMS Server → SMSFilter API
              ↓
         SQLite DB
              ↓
      Android Worker → SMS Network
```

## Installation

### Prerequisites

- Go 1.25 or higher
- SQLite3
- (Optional) Docker and Docker Compose

### Local Development

```bash
# Clone the repository
cd server

# Install dependencies
go mod download

# Run the server
go run main.go
```

### Docker Deployment

```bash
# Build the Docker image
docker build -t microsms-server .

# Run with Docker Compose (recommended)
docker compose up -d
```

## Configuration

Configuration is managed through `config.yaml` and can be overridden with environment variables.

### config.yaml
#### USE THE CONFIG.YAML. YES ENVs CAN WORK TOO, BUT MAKE IT EASIER ON YOURSELF

```yaml
server:
  port: "8080"
  host: "0.0.0.0"

database:
  path: "smsrequest.DB"

filter:
  enabled: true
  apiurl: "http://192.168.8.100:8000/api/filter/sms"
  maxconcurrent: 5      # Max concurrent filter API requests
  resultchansize: 10    # Result channel buffer size
```

### Environment Variables
### See note above, technically this can work, but it is more confusing than using the .yaml

All configuration values can be overridden using environment variables with the `MICROSMS_` prefix:

```bash
export MICROSMS_SERVER_PORT=8080
export MICROSMS_SERVER_HOST=0.0.0.0
export MICROSMS_DATABASE_PATH=/app/data/smsrequest.db
export MICROSMS_FILTER_APIURL=http://smsfilter:8000/api/filter/sms
export MICROSMS_FILTER_MAXCONCURRENT=5
export MICROSMS_FILTER_RESULTCHANSIZE=10
```

## API Endpoints

All endpoints are prefixed with `/api/v0`

### Create SMS Request

Create a new SMS request that will be filtered and queued for sending.

```http
POST /api/v0/create
Content-Type: application/json

{
  "number": "555-123-4567",
  "message": "Hello, this is a test message"
}
```

**Response:**
```json
{
  "message": "SMSRequest Created <uuid>",
  "smsrequest": {
    "id": "uuid-here",
    "number": "555-123-4567",
    "status": "payment_owed",
    "message": "Hello, this is a test message",
    "created": 1234567890
  }
}
```

### Get SMS Request

Retrieve a specific SMS request by ID.

```http
GET /api/v0/smsrequest?id=<uuid>
```

**Response:**
```json
{
  "message": "SMSRequest found",
  "smsrequest": {
    "id": "uuid-here",
    "number": "555-123-4567",
    "status": "ready_to_send",
    "message": "Hello, this is a test message",
    "created": 1234567890
  }
}
```

### Get Ready to Send SMS

Get the earliest SMS request that's ready to be sent. Used by the Android worker.

```http
GET /api/v0/ready
```

**Response:**
```json
{
  "message": "SMS Request <uuid> ready to send",
  "smsrequest": {
    "id": "uuid-here",
    "number": "555-123-4567",
    "status": "ready_to_send",
    "message": "Hello, this is a test message",
    "created": 1234567890
  }
}
```

### Update SMS Request

Update the status of an SMS request. Used by the Android worker to mark messages as taken/sent.

```http
PATCH /api/v0/smsrequest?id=<uuid>
Content-Type: application/json

{
  "status": "sent"
}
```

**Valid Status Values:**
- `payment_owed`: Initial state, awaiting payment (or filter check)
- `ready_to_send`: Filtered and ready for sending
- `taken`: Picked up by Android worker
- `sent`: Successfully sent
- `error`: Error occurred during processing
- `blocked`: Blocked by content filter

### Health Check

Check server health and uptime.

```http
GET /api/v0/health
```

**Response:**
```json
{
  "message": "API Healthy for 1h23m45s"
}
```

## Message Lifecycle

1. **Create**: Client creates SMS request via `/create` endpoint
2. **Filter**: Server automatically sends message to SMSFilter API for safety check
3. **Queue**: If safe, status updates to `ready_to_send`
4. **Pickup**: Android worker polls `/ready` and marks message as `taken`
5. **Send**: Android worker sends SMS and updates status to `sent`
6. **Block**: If unsafe, status updates to `blocked` and message is not sent

## Concurrency & Throttling

The server implements two levels of concurrency control:

### Filter API Throttling

Limits concurrent requests to the SMSFilter API to prevent overwhelming the service:

```yaml
filter:
  maxconcurrent: 5  # Max 5 concurrent filter API calls
```

Set to `0` for unlimited concurrent requests.

### Result Channel Buffering

Controls how many filter results can be queued for processing:

```yaml
filter:
  resultchansize: 10  # Buffer up to 10 results
```

Set to `0` for unbuffered channel (blocking).

## Database Schema

### SMSRequest Table

| Column  | Type      | Description                    |
|---------|-----------|--------------------------------|
| id      | UUID      | Primary key                    |
| number  | String    | Phone number (validated)       |
| status  | String    | Current status (enum)          |
| message | String    | SMS message content            |
| created | Int64     | Unix timestamp (auto-created)  |

## Phone Number Validation

Phone numbers must match one of these formats:
- `(555) 123-4567`
- `555-123-4567`
- `555.123.4567`
- `555 123 4567`
- `5551234567`

## Development

### Project Structure

```
server/
├── main.go              # Application entry point
├── config/
│   ├── config.go        # Configuration loader
│   └── config.yaml      # Default configuration
├── models/
│   └── SMSRequest.go    # Database models and methods
├── routes/
│   └── routes.go        # API route handlers
├── helpers/
│   └── helpers.go       # Filter API integration
├── Dockerfile           # Container definition
└── README.md           # This file
```

### Running Tests

```bash
go test ./...
```

### Building from Source

```bash
# Build for current platform
go build -o microsms .

# Build for Linux (useful for Docker)
CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o microsms .
```

## Docker

### Building with Existing Database

To include an existing database file in the Docker image:

```bash
docker build --build-arg BUILD_WITH_DB=true -t microsms-server .
```

Place your `smsrequest.db` file in the server root directory before building.

### Volume Persistence

The Docker container persists the database in `/app/data`. Mount a volume to preserve data:

```bash
docker run -v microsms-data:/app/data microsms-server
```

## Graceful Shutdown

The server waits for all pending filter checks to complete before shutting down:

```bash
# Stop the server (Ctrl+C)
# Output: Server stopped, waiting for pending filter checks to complete...
# Output: All filter checks completed, exiting cleanly
```

## Integration with Android Worker

The Android worker should:

1. Poll `/api/v0/ready` every 2-5 seconds
2. Parse the returned SMS request
3. Update status to `taken` via PATCH
4. Send the SMS
5. Update status to `sent` or `error` via PATCH

See the `android_worker` directory for the companion app.

## Troubleshooting

### Filter API Connection Issues

Check the filter API URL and ensure the SMSFilter service is running:

```bash
curl http://your-filter-api:8000/api/filter/sms \
  -H "Content-Type: application/json" \
  -d '{"sms": "test message"}'
```

### Database Lock Errors

SQLite can have concurrent write issues. If you see lock errors:
- Reduce `maxconcurrent` in config
- Ensure database file has proper permissions
- Check for other processes accessing the database

### Messages Stuck in payment_owed

If messages aren't transitioning to `ready_to_send`:
- Check server logs for filter API errors
- Verify filter API is accessible
- Ensure filter goroutines are running (check logs for "Handling filter result")

## License

Licensed under mepl. Don't do anything I wouldn't or do, I don't care unless you pay me. 

## Related Components

- **SMSFilter**: Content filtering service (see `smsfilter/README.md`)
- **Android Worker**: SMS sending client (see `android_worker/README.md`)