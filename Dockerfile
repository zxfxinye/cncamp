FROM golang:1.17.3-buster as builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
ADD httpserver ./httpserver
RUN go env -w GO111MODULE="on"; \
    go env -w GOPROXY="https://goproxy.io,direct"; \
    go mod vendor; \
    CGO_ENABLED=0 GOARCH=amd64 go build -o / ./httpserver

FROM alpine
ENV VERSION=1.0
COPY --from=builder /httpserver /httpserver
EXPOSE 80
ENTRYPOINT ["/httpserver"]
