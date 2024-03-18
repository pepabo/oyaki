# oyaki

[![CI](https://github.com/pepabo/oyaki/actions/workflows/ci.yml/badge.svg)](https://github.com/pepabo/oyaki/actions/workflows/ci.yml)

Dynamic image quality transformation proxy.

## Usage

### Docker

```
docker pull ghcr.io/pepabo/oyaki:latest
docker run -p 8080:8080 -e "OYAKI_ORIGIN_HOST=example.com" ghcr.io/pepabo/oyaki
```

## Configuration

Environment variables bellow are available.

- `OYAKI_ORIGIN_HOST`: Your origin host. Example: `example.com` (required)
- `OYAKI_ORIGIN_SCHEME`: Scheme to request to your origin. Default: `https`
- `OYAKI_QUALITY`: Image quality. Default: `90`

### Use OpenTelemetry

- `OTEL_EXPORTER_OTLP_ENDPOINT`: OpenTelemetry collector endpoint. Example: `localhost:4317`
