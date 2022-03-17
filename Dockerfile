FROM golang:1.17-alpine AS build

# WORKDIR /wasp-ingest-rtmp
# COPY services/ services/
# COPY util/ util/
# COPY go.mod go.sum main.go Makefile LICENSE /
# RUN CGO_ENABLED=0 go build -o /bin/wasp-video-stream-service

# FROM alpine AS base
# RUN apk --no-cache add nginx-mod-rtmp ffmpeg
# RUN mkdir /etc/nginx/stat/ /etc/nginx/rtmp.d/
# COPY --from=build /bin/wasp-video-stream-service /bin/wasp-video-stream-service
# COPY ./config/default.conf /etc/nginx/conf.d/default.conf
# COPY ./config/nginx.conf /etc/nginx/nginx.conf
# COPY ./config/stat.xsl /etc/nginx/stat/stat.xsl

# FROM base as development
# COPY ./config/stream_dev.conf /etc/nginx/rtmp.d/stream_dev.conf
# EXPOSE 1935 8080
# CMD ["/usr/sbin/nginx"]

# FROM base as production
# COPY ./config/stream_prod.conf /etc/nginx/rtmp.d/stream_prod.conf
# COPY ./scripts/wasp-video-stream-service.sh /bin/ingest-rtmp.sh
# ENV ENV=production
# EXPOSE 1935 8080
# CMD ["sh", "-c", "/bin/printenv > /etc/envars ; /usr/sbin/nginx"]