## Purpose

Cung cấp bộ kiểm thử đơn vị tự động toàn diện nhằm xác thực tính chính xác và an toàn của các module tiện ích mã hóa/JWT, bộ lọc trung gian xác thực và các tầng xử lý nghiệp vụ.

## Requirements

### Requirement: Utility JWT and Password verification
Hệ thống SHALL cung cấp các bài kiểm thử đơn vị cho việc tạo token JWT, giải mã/xác thực token, mã hóa password và đối chiếu mật khẩu.

#### Scenario: Generate and validate access token successfully
- **WHEN** một access token được tạo với UserID, Email, secret hợp lệ và thời hạn xác định
- **THEN** hàm kiểm tra token SHALL trả về claim chứa đúng UserID, Email và không báo lỗi

#### Scenario: Validate expired access token
- **WHEN** một token JWT đã hết hạn được truyền vào hàm xác thực
- **THEN** hàm xác thực SHALL trả về lỗi token hết hạn hoặc không hợp lệ

#### Scenario: Hash and compare password correctly
- **WHEN** một mật khẩu được băm và sau đó đối chiếu với chính mật khẩu đó
- **THEN** kết quả đối chiếu SHALL trả về trùng khớp và không báo lỗi

#### Scenario: Compare password with wrong input
- **WHEN** một mật khẩu đã băm được đối chiếu với một chuỗi mật khẩu sai
- **THEN** kết quả đối chiếu SHALL trả về lỗi không khớp

### Requirement: Authentication and CORS Middleware verification
Hệ thống SHALL cung cấp các bài kiểm thử cho middleware xác thực Bearer token và middleware cấu hình CORS.

#### Scenario: Auth middleware passes with valid Bearer token
- **WHEN** HTTP request gửi kèm header Authorization chứa Bearer token hợp lệ
- **THEN** middleware SHALL cho phép request đi tiếp và thiết lập UserID vào context

#### Scenario: Auth middleware rejects missing or invalid header
- **WHEN** HTTP request không có header Authorization hoặc token sai định dạng
- **THEN** middleware SHALL dừng request và trả về HTTP status 401 Unauthorized

#### Scenario: CORS middleware sets appropriate headers and handles preflight
- **WHEN** HTTP OPTIONS request (preflight) hoặc request thông thường được gửi đến
- **THEN** middleware SHALL đính kèm các CORS headers tương ứng và trả về HTTP 204 cho preflight request

### Requirement: Service layer business logic verification
Hệ thống SHALL cung cấp các bài kiểm thử đơn vị cho tầng AuthService và TodoService sử dụng mock repository để cô lập logic nghiệp vụ.

#### Scenario: Register new user with duplicate email
- **WHEN** AuthService xử lý đăng ký tài khoản với email đã tồn tại trong repository
- **THEN** service SHALL từ chối và trả về lỗi xung đột/đã tồn tại

#### Scenario: Todo pagination and ownership isolation
- **WHEN** TodoService lấy danh sách công việc của một người dùng
- **THEN** service SHALL chỉ trả về các todo thuộc về UserID đó và tuân thủ giới hạn phân trang
