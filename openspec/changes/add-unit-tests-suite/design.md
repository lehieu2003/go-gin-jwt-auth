## Context

Dự án được tổ chức theo cấu trúc phân tầng (`utils`, `middleware`, `service`, `repository`, `handler`). Tầng `service` phụ thuộc vào `repository` interfaces, tầng `middleware` tương tác với Gin context và HTTP headers, tầng `utils` chứa các hàm thuần túy (pure functions). Xem chi tiết tại `proposal.md`.

## Goals / Non-Goals

**Goals:**
- Triển khai unit tests sử dụng quy chuẩn Go idiom (Table-Driven Tests).
- Đạt độ bao phủ cao cho các hàm mã hóa, sinh token JWT trong `internal/utils`.
- Kiểm thử các middleware (`AuthMiddleware`, `CORSMiddleware`) với `net/http/httptest` và `gin.Context`.
- Kiểm thử logic nghiệp vụ tại `internal/service` bằng cách cô lập database thông qua in-memory mock repositories.

**Non-Goals:**
- Chưa triển khai End-to-End (E2E) hoặc Integration Test kết nối trực tiếp đến database PostgreSQL thật trong change này.
- Không sửa đổi logic nghiệp vụ trong production code trừ khi phát hiện bug nghiêm trọng trong quá trình viết test.

## Decisions

- **Decision 1: Áp dụng Table-Driven Tests cho Utils và Middleware**
  - *Rationale*: Cấu trúc bảng (slice of structs) là chuẩn mực trong Go, cho phép kiểm tra nhiều case (happy path, edge case, invalid input) trong một hàm test duy nhất một cách ngắn gọn, rõ ràng.
  - *Alternatives considered*: Viết từng hàm test riêng lẻ (dẫn đến lặp code boilerplate).

- **Decision 2: Sử dụng In-Memory Mock Structs cho Repository Dependencies**
  - *Rationale*: Các service (`AuthService`, `TodoService`) nhận interface repository (`UserRepository`, `TodoRepository`, `TokenRepository`). Việc viết struct mock gọn nhẹ trực tiếp trong file test giúp test chạy tức thì (sub-millisecond), độc lập và không cần thêm các thư viện mock nặng nề.
  - *Alternatives considered*: Cài đặt gomock/mockery (tạo code gen phức tạp không cần thiết ở quy mô hiện tại).

## Risks / Trade-offs

- **[Risk] Mock behavior không khớp 100% với GORM DB behavior** → *Mitigation*: Đảm bảo mock struct tuân thủ đúng các lỗi trả về (ví dụ `gorm.ErrRecordNotFound` hoặc `nil`) theo đúng các case thực tế mà service xử lý.
