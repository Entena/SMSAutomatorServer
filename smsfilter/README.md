# SMSFilter
A lightweight backend for filtering SMS messages based on safety constraints.

## Forked from https://github.com/walkernr/SMSFilter/tree/qwen
Forked due to divergent build systems. Will likely remerge later but this is cheap and easy

# SMSFilter

A lightweight Python backend for filtering SMS messages based on safety constraints using Meta's Llama Guard 3 model. The service provides a REST API that analyzes text content and blocks messages containing harmful, unsafe, or inappropriate content across 13 configurable categories.

## Overview

SMSFilter uses a quantized version of Llama Guard 3 (1B parameters) to classify SMS messages as safe or unsafe. The model runs entirely locally with CPU-only inference via `llama-cpp-python`, making it suitable for edge deployments without requiring GPU resources.

The service is designed to integrate with the MicroSMS server to automatically filter messages before they're queued for sending.

## Features

- **13 Safety Categories**: Configurable filtering across violent crimes, hate speech, sexual content, and more
- **Fast Inference**: Quantized GGUF models for efficient CPU-only operation
- **REST API**: Simple FastAPI endpoint for message filtering
- **Docker Support**: Containerized deployment with multi-stage builds
- **Async Processing**: Non-blocking inference with asyncio
- **Flexible Configuration**: Enable/disable categories via environment variables
- **CORS Support**: Configurable allowed origins for web integration

## Forked from

https://github.com/walkernr/SMSFilter/tree/qwen

Forked due to divergent build systems. Will likely remerge later but this is cheap and easy.

## Installation

### Prerequisites

- Python 3.12
- Docker (optional, recommended)
- 2-4 GB RAM for model inference

### Local Development

```bash
cd smsfilter

# Install dependencies with uv
pip install uv==0.6.2

# Build and install all modules
cd backend/settings
uv sync --no-editable --no-dev

cd ../predictor
uv sync --no-editable --no-dev

cd ../api
uv sync --no-editable --no-dev

# Run the server
cd backend/api
uvicorn api.webapp:app --host 0.0.0.0 --port 8000
```

### Docker Deployment (Recommended)

```bash
# Build the image
docker build -t smsfilter-backend .

# Run with docker compose
docker compose up -d
```

## Configuration

Configuration is managed through environment variables, typically set in a `.env` file.

### Environment Variables

Create a `.env` file in the `smsfilter` directory:

```bash
# User and permissions
UID=10001
GID=10001

# Networking - CORS allowed origins
ALLOWED_HOSTS='["http://localhost:5173", "http://127.0.0.1:5173", "http://10.0.0.119:5173"]'

# Application directory
APP_DIR=/app

# Model quantization level (affects speed vs accuracy)
# Options: Q2_K, Q3_K_L, Q3_K_M, Q3_K_S, Q4_0, Q4_1, Q4_K_M, Q4_K_S,
#          Q5_0, Q5_1, Q5_K_M, Q5_K_S, Q6_K, Q8_0
QUANT=Q5_1

# Safety categories (set to true to enable filtering, false to disable)
# Note that we DO NOT recommend disabling anything, make sure you know what you are doing
VIOLENT_CRIMES=true
NONVIOLENT_CRIMES=true
SEX_RELATED_CRIMES=true
CHILD_SEXUAL_EXPLOITATION=true
DEFAMATION=true
SPECIALIZED_ADVICE=true
PRIVACY=true
INTELLECTUAL_PROPERTY=true
INDISCRIMINATE_WEAPONS=true
HATE=true
SUICIDE_AND_SELF_HARM=true
SEXUAL_CONTENT=true
ELECTIONS=true
```

### Safety Categories

The 13 configurable categories from Llama Guard 3:

| Category | Code | Description |
|----------|------|-------------|
| S1 | VIOLENT_CRIMES | Violence, murder, assault |
| S2 | NONVIOLENT_CRIMES | Theft, fraud, vandalism |
| S3 | SEX_RELATED_CRIMES | Sexual assault, trafficking |
| S4 | CHILD_SEXUAL_EXPLOITATION | Child abuse material |
| S5 | DEFAMATION | Libel, slander, false accusations |
| S6 | SPECIALIZED_ADVICE | Unlicensed legal/medical/financial advice |
| S7 | PRIVACY | Doxxing, privacy violations |
| S8 | INTELLECTUAL_PROPERTY | Copyright infringement |
| S9 | INDISCRIMINATE_WEAPONS | Bombs, WMDs, bioweapons |
| S10 | HATE | Hate speech, discrimination |
| S11 | SUICIDE_AND_SELF_HARM | Self-harm encouragement |
| S12 | SEXUAL_CONTENT | Explicit sexual content |
| S13 | ELECTIONS | Election interference |

Set any category to `false` to disable filtering for that category.

### Quantization Levels

The `QUANT` setting controls the model quantization level, trading off between speed and accuracy:

| Level | Size | Speed | Accuracy | Recommended For |
|-------|------|-------|----------|-----------------|
| Q2_K | Smallest | Fastest | Lower | Testing only |
| Q4_K_M | Small | Fast | Good | Edge devices |
| Q5_1 | Medium | Balanced | Better | **Default, recommended** |
| Q8_0 | Large | Slower | Best | High accuracy needs |

## API Documentation

### Filter SMS Endpoint

Filter an SMS message for safety concerns.

**Endpoint:** `POST /api/filter/sms`

**Request Body:**
```json
{
  "sms": "Your message text here",
  "violent_crimes": true,
  "nonviolent_crimes": true,
  "sex_related_crimes": true,
  "child_sexual_exploitation": true,
  "defamation": false,
  "specialized_advice": false,
  "privacy": true,
  "intellectual_property": false,
  "indiscriminate_weapons": true,
  "hate": true,
  "suicide_and_self_harm": true,
  "sexual_content": true,
  "elections": false
}
```

All category fields are optional. If not provided, the service uses the defaults from environment variables.

**Response (Safe Message):**
```json
{
  "blocked": false,
  "reason": null,
  "included_categories": [
    "VIOLENT_CRIMES",
    "HATE",
    "SEXUAL_CONTENT",
    "UNKNOWN"
  ],
  "excluded_categories": [
    "DEFAMATION",
    "SPECIALIZED_ADVICE"
  ]
}
```

**Response (Unsafe Message):**
```json
{
  "blocked": true,
  "reason": "HATE",
  "included_categories": [
    "VIOLENT_CRIMES",
    "HATE",
    "SEXUAL_CONTENT",
    "UNKNOWN"
  ],
  "excluded_categories": [
    "DEFAMATION",
    "SPECIALIZED_ADVICE"
  ]
}
```

### Testing with curl

```bash
# Test a safe message
curl -X POST http://localhost:8000/api/filter/sms \
  -H "Content-Type: application/json" \
  -d '{"sms": "Hello, how are you today?"}'

# Test with custom categories
curl -X POST http://localhost:8000/api/filter/sms \
  -H "Content-Type: application/json" \
  -d '{
    "sms": "Your message here",
    "hate": true,
    "sexual_content": false
  }'
```

## Architecture

### Project Structure

```
smsfilter/
├── backend/
│   ├── settings/          # Configuration management
│   │   ├── src/settings/
│   │   │   ├── __init__.py
│   │   │   └── settings.py
│   │   └── pyproject.toml
│   ├── predictor/         # Model inference logic
│   │   ├── src/predictor/
│   │   │   ├── __init__.py
│   │   │   ├── predict.py
│   │   │   └── schemas.py
│   │   └── pyproject.toml
│   └── api/               # FastAPI application
│       ├── src/api/
│       │   ├── __init__.py
│       │   ├── webapp.py
│       │   ├── schemas.py
│       │   ├── dependencies.py
│       │   └── routers/
│       │       └── sms_filter.py
│       └── pyproject.toml
├── docker-compose.yaml
├── Dockerfile
├── .env
└── README.md
```

### Module Dependencies

```
api → predictor → settings
```

- **settings**: Base configuration module with Pydantic settings
- **predictor**: Model loading and inference logic
- **api**: FastAPI web application and routes

### Async Inference

The predictor uses `asyncio` with a lock to handle concurrent requests safely:

```python
async with self.lock:
    # Only one inference at a time (model is CPU-bound)
    result = await loop.run_in_executor(None, _inference)
```

This prevents multiple simultaneous inferences from overwhelming CPU resources.

## Docker

### Multi-Stage Build

The Dockerfile uses a multi-stage build to minimize image size:

1. **builder_base**: Creates Python virtual environment
2. **builder**: Installs dependencies and builds modules
3. **runtime**: Minimal runtime image with only necessary files

### Image Details

- Base: `python:3.12-slim-trixie`
- Runtime user: Non-root user (UID/GID configurable)
- Entrypoint: `uvicorn api.webapp:app`
- Port: 8000
- Model cache: `/app/.cache/huggingface`

### Running with Docker

```bash
# Build image
docker build -t smsfilter-backend .

# Run standalone
docker run -p 8000:8000 \
  --env-file .env \
  smsfilter-backend

# Run with docker compose
docker compose up -d

# View logs
docker compose logs -f smsfilter
```

## Performance

### Inference Speed

Typical inference times (Q5_1 quantization on modern CPU):

- Short messages (< 50 chars): 100-200ms
- Medium messages (50-200 chars): 200-400ms
- Long messages (> 200 chars): 400-800ms

### Resource Usage

- **Memory**: 1.5-2.5 GB (model + overhead)
- **CPU**: 1-2 cores recommended
- **Disk**: ~1.5 GB for Q5_1 model

### Concurrency

The service handles one inference at a time due to CPU-bound nature. For higher throughput:

1. Run multiple instances behind a load balancer
2. Use lower quantization (Q4_K_M) for faster inference
3. Consider GPU inference for high-volume scenarios

## Development

### Build (Docker)

The project uses habushu for Maven/Python integration:

```bash
# Clean and build
## Note this is a python image, not much needs to be done in terms of actual
## compilation, instead we just bundle an image and drop python deps. If you
## want to launch the python cleanly on your machine refer to the Dockerfile
docker build -t smsfilter-backend .
```

### Running Tests

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=api --cov=predictor --cov=settings
```

## Integration with MicroSMS Server

The MicroSMS server automatically calls this service when new SMS requests are created:

1. Client creates SMS via `/api/v0/create`
2. MicroSMS server sends message to `/api/filter/sms`
3. SMSFilter analyzes content and returns result
4. MicroSMS updates status to `ready_to_send` or `blocked`

Configure the MicroSMS server to point to this service:

```yaml
# In server/config.yaml
filter:
  apiurl: "http://smsfilter:8000/api/filter/sms"
```

## Troubleshooting
* Odds are any issues you have are releated to Out Of Memory or OOM errors. Check your Docker control panel to allocate more if needed
* The service isn't the best, but can be run alongside on relatively "performant" machines. For your reference my 2.4GHz I9 with 32GB RAM macbook ran this and my microsms server (along with 3 web browsers, multiple IDEs, and Spotify), but she wasn't happy. Abandon weak boxes yee who enter


### Model Download Issues
* Odds are if you are facing model download issues your Docker container probably doesn't have network access

If the model fails to download:

```bash
# Manually download model
python -c "from transformers import AutoTokenizer; AutoTokenizer.from_pretrained('QuantFactory/Llama-Guard-3-1B-GGUF', gguf_file='Llama-Guard-3-1B.Q5_1.gguf')"

python -c "from llama_cpp import Llama; Llama.from_pretrained(repo_id='QuantFactory/Llama-Guard-3-1B-GGUF', filename='Llama-Guard-3-1B.Q5_1.gguf')"
```

### Out of Memory Errors

Reduce quantization level or increase available memory:

```bash
# Use lower quantization
QUANT=Q4_K_M
```

### Slow Inference

Try a more aggressive quantization:

```bash
# Faster but less accurate
QUANT=Q4_0
```

### Permission Errors in Docker

Ensure UID/GID match your user:

```bash
# Set in .env
UID=$(id -u)
GID=$(id -g)
```

## API Health Check

Check if the service is running:

```bash
curl http://localhost:8000/docs
```

This opens the automatic FastAPI documentation interface.

## License

Stolen from the forked repo. Refer to their notes

## Related Components

- **MicroSMS Server**: Main API server (see `server/README.md`)
- **Android Worker**: SMS sending client (see `android_worker/README.md`)