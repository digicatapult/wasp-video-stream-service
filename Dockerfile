
FROM golang:1.17-alpine AS build

WORKDIR /wasp-video-stream-service
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/wasp-video-stream-service

FROM alpine
COPY --from=build /bin/wasp-video-stream-service /bin/wasp-video-stream-service
CMD ["/bin/wasp-video-stream-service"]