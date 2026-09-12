## Why

Dự án hiện chưa có quy trình Tự động hóa Tích hợp Liên tục (CI - Continuous Integration). Việc kiểm tra lỗi cú pháp, linting, chạy unit test và build binary vẫn đang phụ thuộc vào thao tác thủ công của lập trình viên. Thiết lập GitHub Actions CI workflow giúp tự động kiểm tra tính đúng đắn của mã nguồn, phát hiện lỗi hồi quy sớm trên mọi commit và Pull Request trước khi merge.

## What Changes

- Thiết lập GitHub Actions workflow (`.github/workflows/ci.yml`) kích hoạt tự động trên các sự kiện `push` và `pull_request` vào nhánh `main`.
- Cấu hình môi trường Go tương thích với phiên bản trong `go.mod` kèm cơ chế caching module/build cache để tối ưu thời gian chạy.
- Thực hiện kiểm tra chất lượng mã nguồn: `go vet`, `gofmt` (formatting check), và `golangci-lint` (tùy chọn/cơ bản).
- Chạy toàn bộ bộ kiểm thử tự động (Unit test suite) với cờ phát hiện xung đột dữ liệu race condition (`go test -race -v -cover ./...`) và xuất báo cáo coverage.
- Kiểm tra tính toàn vẹn của quá trình build binary ứng dụng (`go build -v ./cmd/server`).

## Capabilities

### New Capabilities
- `ci-pipeline`: Tự động hóa kiểm tra mã nguồn, chạy unit test và build kiểm thử trên GitHub Actions khi có commit hoặc Pull Request.

### Modified Capabilities
<!-- None -->

## Impact

- Thêm thư mục và tệp cấu hình mới `.github/workflows/ci.yml`.
- Không ảnh hưởng tiêu cực hoặc gây breaking changes đối với logic hiện có của ứng dụng.
- Tăng độ tin cậy và tự động hóa quy trình phát triển cho dự án.
