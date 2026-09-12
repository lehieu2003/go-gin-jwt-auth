# Go + Gin + PostgreSQL JWT Authentication Boilerplate

Hệ thống xác thực (Authentication System) hoàn chỉnh được xây dựng bằng **Go (Golang)**, **Gin Framework**, **PostgreSQL (GORM)**, **JWT Access Tokens**, và cơ chế **Refresh Token Rotation (RTR)** với **Phát hiện tái sử dụng token (Reuse & Breach Detection)** lưu trong **HttpOnly Cookie**.

> 📖 **Tài liệu cho người mới bắt đầu (Beginner Guide)**: [LEARNING_GUIDE.md](file:///C:/Users/daong/Downloads/golang/LEARNING_GUIDE.md)  
> 🏛️ **Tài liệu kiến trúc & công nghệ chi tiết**: [ARCHITECTURE.md](file:///C:/Users/daong/Downloads/golang/ARCHITECTURE.md)

---

## 🛠️ Tech Stack & Kiến trúc tóm tắt

- **Ngôn ngữ & Framework**: Go 1.22+, Gin Gonic
- **Cơ sở dữ liệu**: PostgreSQL 16 (GORM Auto-Migration)
- **Kiến trúc**: **Clean Architecture / Layered Architecture** (Handler -> Service -> Repository -> Model)
- **Xác thực & Bảo mật**:
  - **Access Token (JWT)**: Thời hạn ngắn (15 phút), truyền qua header `Authorization: Bearer <token>`.
  - **Refresh Token**: Chuỗi ngẫu nhiên 256-bit an toàn, lưu bản **băm SHA-256** trong PostgreSQL.
  - **HttpOnly Cookie**: Lưu Refresh Token ở cookie với cờ `HttpOnly`, `SameSite=Lax`, `Secure`, chống tấn công XSS.
  - **Refresh Token Rotation (RTR)**: Mỗi lần refresh sẽ hủy token cũ và cấp token mới cùng `family_id`.
  - **Reuse & Breach Detection**: Nếu phát hiện token cũ bị dùng lại (dấu hiệu bị đánh cắp), hệ thống sẽ **hủy toàn bộ chuỗi token trong `family_id`** để bảo vệ tài khoản.

---

## 🚀 Hướng dẫn Cài đặt & Chạy ứng dụng

### 1. Khởi động PostgreSQL qua Docker

```bash
docker compose up -d
```

### 2. Cấu hình biến môi trường (`.env`)

File `.env` đã được tạo sẵn mẫu:

```ini
PORT=8080
ENV=development

# PostgreSQL Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth_db
DB_SSLMODE=disable

# JWT Secret Keys
ACCESS_TOKEN_SECRET=dev-access-token-secret-change-in-production-123456789
REFRESH_TOKEN_SECRET=dev-refresh-token-secret-change-in-production-123456789

# Thời hạn Token
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=168h # 7 ngày

# Cookie Config
COOKIE_DOMAIN=
COOKIE_SECURE=false # Đặt thành true khi chạy Production với HTTPS
COOKIE_SAMESITE=Lax
```

### 3. Chạy Server

```bash
go run cmd/server/main.go
```

Server sẽ lắng nghe tại: `http://localhost:8080` (GORM sẽ tự động tạo bảng `users` và `refresh_tokens`).

---

## 📡 Danh sách API & Ví dụ cURL Test

### 1. Đăng ký tài khoản (Register)
- **URL**: `POST /api/v1/auth/register`
- **Body**:
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecretPassword123!"
  }' -c cookies.txt
```
*Kết quả: Trả về Access Token trong JSON body và tự động lưu Refresh Token vào `cookies.txt` (HttpOnly cookie).*

---

### 2. Đăng nhập (Login)
- **URL**: `POST /api/v1/auth/login`
- **Body**:
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecretPassword123!"
  }' -c cookies.txt
```

---

### 3. Lấy thông tin người dùng hiện tại (Protected Route)
- **URL**: `GET /api/v1/auth/me`
- **Header**: `Authorization: Bearer <ACCESS_TOKEN>`
```bash
curl -i -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>"
```

---

### 4. Làm mới Token với cơ chế xoay vòng (Refresh Token Rotation)
- **URL**: `POST /api/v1/auth/refresh`
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt
```
*Token cũ sẽ bị hủy (revoked) và một Refresh Token mới cùng Access Token mới sẽ được cấp phát.*

---

### 5. Đăng xuất (Logout)
- **URL**: `POST /api/v1/auth/logout`
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/logout \
  -b cookies.txt \
  -c cookies.txt
```
*Hủy chuỗi session trong database và xóa HttpOnly cookie khỏi trình duyệt.*

---

## 📂 Sơ đồ Cấu trúc Thư mục

```
golang/
├── cmd/server/main.go               # Entry point, Dependency Injection, Router
├── internal/
│   ├── config/config.go            # Đọc cấu hình .env
│   ├── database/database.go        # Kết nối PostgreSQL & GORM AutoMigrate
│   ├── handler/auth_handler.go      # HTTP Controller & Cookie Set/Clear
│   ├── middleware/
│   │   ├── auth_middleware.go      # Middleware xác thực JWT Bearer
│   │   └── cors_middleware.go      # CORS hỗ trợ credentials/cookies
│   ├── model/
│   │   ├── user.go                 # Entity User
│   │   └── refresh_token.go        # Entity RefreshToken
│   ├── repository/
│   │   ├── user_repository.go      # Interface & DB logic cho Users
│   │   └── token_repository.go     # Interface & DB logic cho Tokens
│   ├── service/auth_service.go      # Business Logic: Auth, RTR, Breach Detection
│   └── utils/
│       ├── jwt.go                  # JWT Sign/Verify, Crypto Random, SHA-256
│       └── password.go             # Bcrypt Hash & Compare
├── .env / .env.example              # Cấu hình môi trường
├── docker-compose.yml               # PostgreSQL 16 Alpine
├── ARCHITECTURE.md                  # Tài liệu phân tích kiến trúc & luồng bảo mật
└── README.md                        # Hướng dẫn sử dụng
```
