# ============================================
# 阶段 1：构建 Go 后端
# ============================================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 先设置代理
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 修复在这里：用 go install 安装 swag，不会报错
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go -o ./docs


RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o alert-ops ./cmd/server/main.go

# ============================================
# 阶段 2：构建前端
# ============================================
FROM node:22 AS frontend

WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci --registry=https://registry.npmmirror.com 2>/dev/null || npm install --registry=https://registry.npmmirror.com

COPY web/ ./
RUN npm run build

# ============================================
# 阶段 3：运行镜像
# ============================================
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/alert-ops .
COPY --from=builder /app/conf ./conf
COPY --from=frontend /web/dist ./web/dist

EXPOSE 8080

CMD ["./alert-ops"]
