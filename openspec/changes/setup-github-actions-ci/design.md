## Context

Xem [proposal.md](file:///C:/Users/daong/Downloads/golang/openspec/changes/setup-github-actions-ci/proposal.md) để biết động lực và bối cảnh tổng quan.
Dự án viết bằng Go 1.22+ với kiến trúc Clean Architecture, sử dụng PostgreSQL, Gin framework và GORM. Đã có bộ unit test độc lập không phụ thuộc vào database thật (dùng in-memory mocks).

## Goals / Non-Goals

**Goals:**
- Tạo workflow GitHub Actions chuẩn hóa trong `.github/workflows/ci.yml`.
- Tự động hóa các bước: Checkout -> Setup Go với cache -> Dependency download -> Formatting / Linting -> Unit testing (race detector + coverage) -> Compile build.
- Thời gian thực thi nhanh (dưới 2 phút) nhờ tận dụng caching Go modules và build artifacts của `actions/setup-go`.

**Non-Goals:**
- Tự động deploy lên production (CD - Continuous Deployment) hoặc cấu hình cloud hosting (sẽ triển khai ở thay đổi riêng).
- Khởi chạy PostgreSQL container cho integration test (tất cả unit tests hiện tại đều dùng mock repository).

## Decisions

### 1. Sử dụng GitHub Actions Native Runners (`ubuntu-latest`)
- **Quyết định**: Sử dụng `ubuntu-latest` với action `actions/checkout@v4` và `actions/setup-go@v5`.
- **Lý do**: Được GitHub hỗ trợ tối ưu, miễn phí, tích hợp sẵn tính năng tự động caching `go.sum` và module download mà không cần cấu hình `actions/cache` phức tạp.
- **Giải pháp thay thế**: Cấu hình tự build Docker image để test (chậm hơn và tốn tài nguyên hơn).

### 2. Thứ tự thực thi trong Job CI
- **Quyết định**: Gộp các bước kiểm tra vào một job `build-and-test` tuần tự:
  1. `Check formatting` (`gofmt -l -s .`)
  2. `Static analysis` (`go vet ./...`)
  3. `Run Unit Tests` (`go test -race -v -coverprofile=coverage.out ./...`)
  4. `Build Binary` (`go build -v ./cmd/server`)
- **Lý do**: Thứ tự từ bước nhẹ nhất (formatting, linting) -> kiểm thử (test) -> biên dịch (build), giúp phát hiện lỗi cú pháp sớm và ngắt sớm (fail fast).

## Risks / Trade-offs

- **[Risk]** Phiên bản Go trong CI không đồng nhất với máy local → **Mitigation**: Chỉ định `go-version-file: 'go.mod'` trong `actions/setup-go` để CI luôn tự động đọc phiên bản Go từ `go.mod`.
- **[Risk]** Lỗi Race condition chạy trên môi trường đa nhân trong CI bị phát hiện mà local không thấy → **Mitigation**: Luôn bật cờ `-race` trong câu lệnh `go test`.
