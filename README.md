# wasp-video-stream-service

## Getting Started

### Requirements

### Building

```
make build
```

### Testing

```
make test
```

### Linting

```
make lint
```

### Cleanbuild

```
make cleanbuild
```

### Environment Variables

- `IN_TOPIC_NAME_KEY` (default: "video")
- `KAFKA_BROKERS` (default: "localhost:9092") - comma separated list
- `ENV` (default: "development") - `development|production`
- `LOG_LEVEL` (default: "debug") - `debug|info|warn|error|fatal`
- `HOST_ADDRESS` (default: "localhost:9999")

### Logging
The `ENV` environment variable is required to enable log level filtering via the `production` value:
```
ENV=production LOG_LEVEL=info HOST_ADDRESS=localhost:8080 ./wasp-video-stream-service ./wasp-video-stream-service
```

### Example

```
$ make build
$ HOST_ADDRESS=localhost:8080 ./wasp-video-stream-service
```
