# Fastener API2 (Go 後端服務)

這個專案是一個基於 Go 語言和 Echo 框架構建的後端 RESTful API 服務。它提供了多個資源的管理功能，支援多人登入、基於 JWT 的身份驗證和細粒度的多級權限管理。

## 架構概覽

本專案採用分層架構，包括：

* **Handler 層**：處理 HTTP 請求和響應，調用 Service 層。
* **Service 層**：封裝業務邏輯，協調 Repository 層。
* **Repository 層**：處理資料庫操作，封裝資料庫細節。
* **Middleware 層**：處理跨域 (CORS)、JWT 身份驗證、細粒度授權。
* **Models 層**：定義資料模型（Go 結構體）。
* **Config 層**：集中管理應用程式配置。
* **Utils 層**：提供通用工具函數和統一錯誤處理。

## 技術棧

* **語言**：Go (Golang)
* **Web 框架**：[Echo](https://echo.labstack.com/)
* **資料庫**：PostgreSQL (透過 `github.com/lib/pq` 驅動)
* **身份驗證**：JWT (JSON Web Tokens)，支援 Access Token 和 Refresh Token
* **密碼雜湊**：Bcrypt
* **環境變數**：`godotenv`
* **驗證**：`go-playground/validator`
* **日誌**：`go.uber.org/zap` (結構化日誌)
* **容器化**：Docker

## 功能列表

* **用戶管理**：帳戶的 CRUD 操作，包括創建、查詢、更新、刪除。
* **身份驗證**：用戶登入、註冊，JWT Access Token 和 Refresh Token 的簽發與刷新。
* **權限管理**：基於角色的細粒度權限控制，配置哪些角色可以執行哪些操作。
* **選單管理**：系統選單的 CRUD 操作，支援角色與選單的關聯。
* **公司管理**：對公司資訊的 CRUD 操作。
* **客戶管理**：對客戶資訊的 CRUD 操作。
* **產品定義管理**：對產品類別和產品定義的 CRUD 操作。
* **重設管理員工具**：獨立的命令列工具，用於安全地重設管理員密碼。

## 環境變數配置

在專案根目錄創建 `.env` 檔案，並配置以下變數：
