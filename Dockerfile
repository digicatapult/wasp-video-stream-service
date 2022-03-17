
FROM golang:1.17-alpine AS build

WORKDIR /wasp-videeo-stream-service
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/wasp-videeo-stream-service

FROM alpine
COPY --from=build /bin/wasp-videeo-stream-service /bin/wasp-videeo-stream-service
CMD ["/bin/wasp-video-stream-service"]