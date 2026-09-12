## 1. Utils Unit Tests

- [x] 1.1 Viết unit test cho JWT trong `internal/utils/jwt_test.go` (bao gồm `GenerateAccessToken`, `ValidateToken`, `HashRefreshToken`) và kiểm tra bằng lệnh `go test -v ./internal/utils`.
- [x] 1.2 Viết unit test cho Password hashing trong `internal/utils/password_test.go` (bao gồm `HashPassword`, `ComparePassword`) và kiểm tra bằng lệnh `go test -v ./internal/utils`.

## 2. Middleware Unit Tests

- [x] 2.1 Viết unit test cho `AuthMiddleware` trong `internal/middleware/auth_middleware_test.go` (kiểm tra token hợp lệ, thiếu Bearer header, token hết hạn/không hợp lệ) và kiểm tra bằng lệnh `go test -v ./internal/middleware`.
- [x] 2.2 Viết unit test cho `CORSMiddleware` trong `internal/middleware/cors_middleware_test.go` (kiểm tra CORS headers và preflight OPTIONS 204) và kiểm tra bằng lệnh `go test -v ./internal/middleware`.

## 3. Service Layer Unit Tests

- [x] 3.1 Xây dựng in-memory mock repositories và viết unit test cho `AuthService` trong `internal/service/auth_service_test.go` (kiểm tra Register, Login, RefreshToken, RevokeToken) và kiểm tra bằng lệnh `go test -v ./internal/service`.
- [x] 3.2 Xây dựng in-memory mock repository và viết unit test cho `TodoService` trong `internal/service/todo_service_test.go` (kiểm tra CreateTodo, GetTodoByID, ListTodos, UpdateTodo, DeleteTodo kèm phân quyền UserID) và kiểm tra bằng lệnh `go test -v ./internal/service`.

## 4. Test Suite Verification & Coverage

- [x] 4.1 Chạy toàn bộ test suite dự án bằng lệnh `go test -v -cover ./...` và xác nhận 100% test cases đều pass.
