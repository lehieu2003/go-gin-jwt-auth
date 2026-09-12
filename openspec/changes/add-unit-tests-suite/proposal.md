## Why

Dự án Todo & Auth backend bằng Go hiện tại đang thiếu hệ thống kiểm thử tự động (chỉ có 1 test đơn lẻ ở handler). Việc thiếu unit test khiến dự án có rủi ro cao khi refactor, tiềm ẩn lỗi logic liên quan đến bảo mật (JWT, Password Hashing, Authentication Middleware) và các quy tắc nghiệp vụ trong Service layer. Việc triển khai bộ unit test toàn diện giúp đảm bảo tính ổn định, độ tin cậy và đáp ứng yêu cầu chất lượng mã nguồn.

## What Changes

- Bổ sung bộ Unit Test cho tầng tiện ích (`internal/utils`): bao gồm `jwt_test.go` và `password_test.go`.
- Bổ sung bộ Unit Test cho tầng middleware (`internal/middleware`): bao gồm `auth_middleware_test.go` và `cors_middleware_test.go`.
- Bổ sung bộ Unit Test cho tầng nghiệp vụ (`internal/service`): bao gồm `auth_service_test.go` và `todo_service_test.go` (sử dụng in-memory mock repository).
- Cung cấp tài liệu và lệnh chạy test tiêu chuẩn với Go toolchain (`go test ./... -v -cover`).

## Capabilities

### New Capabilities
- `unit-tests`: Quy định tiêu chuẩn kiểm thử đơn vị cho toàn bộ các thành phần lõi (Utils, Middleware, Service) của hệ thống.

### Modified Capabilities
<!-- No modified capabilities -->

## Impact

- **Affected code**:
  - `internal/utils/*_test.go` (mới)
  - `internal/middleware/*_test.go` (mới)
  - `internal/service/*_test.go` (mới)
  - Không làm thay đổi logic thực thi hiện tại của mã nguồn chính (không gây breaking changes).
- **Dependencies**: Sử dụng thư viện chuẩn của Go (`testing`, `net/http/httptest`, `github.com/stretchr/testify` nếu cần hoặc thuần chuẩn Go stdlib).
