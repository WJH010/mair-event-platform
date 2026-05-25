# ============ 构建阶段 ============
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server ./cmd/main.go

# ============ 运行阶段 ============
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/config.docker.yaml ./config.yaml

RUN mkdir -p /app/logs

EXPOSE 8080

ENTRYPOINT ["./server"]