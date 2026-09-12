## Purpose

Tự động hóa toàn bộ quy trình kiểm tra chất lượng mã nguồn, chạy bộ kiểm thử đơn vị và build kiểm thử trên hệ thống GitHub Actions CI khi có commit hoặc Pull Request.

## Requirements

### Requirement: Automated Trigger and Setup
Hệ thống CI SHALL tự động kích hoạt mỗi khi có mã nguồn được push hoặc Pull Request được mở/cập nhật vào nhánh `main`.

#### Scenario: Trigger on push to main branch
- **WHEN** lập trình viên push commit mới lên nhánh `main`
- **THEN** GitHub Actions SHALL khởi chạy workflow CI trên môi trường runner Ubuntu mới nhất

#### Scenario: Trigger on Pull Request to main branch
- **WHEN** lập trình viên mở hoặc cập nhật Pull Request nhắm vào nhánh `main`
- **THEN** GitHub Actions SHALL khởi chạy workflow CI để kiểm tra trước khi cho phép merge

### Requirement: Code Quality and Linting Verification
Hệ thống CI SHALL kiểm tra tính hợp lệ về định dạng và phân tích tĩnh lỗi mã nguồn Go.

#### Scenario: Verify formatting and vetting
- **WHEN** workflow CI thực thi bước kiểm tra code
- **THEN** hệ thống SHALL chạy `go vet ./...` và kiểm tra định dạng `gofmt` để đảm bảo code tuân thủ quy chuẩn Go

### Requirement: Automated Unit Testing and Coverage
Hệ thống CI SHALL thực thi toàn bộ unit test suite và phát hiện các lỗi tranh chấp bộ nhớ (race condition).

#### Scenario: Run all unit tests with race detection
- **WHEN** workflow CI thực thi bước kiểm thử
- **THEN** hệ thống SHALL chạy `go test -race -v -cover ./...` và toàn bộ các test cases phải PASS thành công

### Requirement: Application Binary Compilation
Hệ thống CI SHALL kiểm tra việc biên dịch ứng dụng chính mà không xảy ra lỗi build.

#### Scenario: Build server binary successfully
- **WHEN** workflow CI thực thi bước build
- **THEN** lệnh `go build -v ./cmd/server` SHALL hoàn thành với exit code 0
