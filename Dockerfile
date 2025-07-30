# Dockerfile

# --- 第一階段：建置器 (Builder) ---
# 使用官方 Go 映像檔作為基礎，其中包含建置 Go 應用程式所需的所有工具
FROM golang:1.22-alpine AS builder

# 設定工作目錄
WORKDIR /app

# 拷貝 go.mod 和 go.sum 檔案，並下載 Go 模組
# 這樣做可以利用 Docker 層的快取機制，如果 go.mod/go.sum 沒有變化，則可以跳過模組下載步驟
COPY go.mod go.sum ./
RUN go mod download

# 拷貝應用程式的原始碼
COPY . .

# 建置主應用程式
# CGO_ENABLED=0 禁止 CGO，使建置出的二進位檔案靜態鏈接，無需依賴系統庫，更易於部署到最小化映像中
# -o main 指定輸出檔案名
# ./main.go 指定入口檔案
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o main ./main.go

# 建置 resetadmin 工具
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o resetadmin ./cmd/resetadmin/main.go

# --- 第二階段：運行器 (Runner) ---
# 使用一個更小、更安全的基礎映像檔來運行應用程式 (通常不包含建置工具)
# alpine/git 是一個輕量級的映像，包含 git，用於一些可能需要 git 的工具
FROM alpine/git AS runner

# 創建一個非 root 用戶，以增強安全性
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# 設定工作目錄
WORKDIR /app

# 拷貝第一階段建置好的可執行檔
COPY --from=builder /app/main .
COPY --from=builder /app/resetadmin .

# 拷貝資料庫遷移腳本
# 如果您確實有 db/migrations 目錄，請確保在建置時它被正確拷貝
COPY --from=builder /app/db/migrations ./db/migrations

# 暴露應用程式監聽的端口
EXPOSE 8080

# 定義容器啟動時執行的命令
# 預設執行應用程式的主可執行檔
ENTRYPOINT ["./main"]
