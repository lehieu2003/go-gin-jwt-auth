## 1. GitHub Actions Workflow Configuration

- [x] 1.1 Tạo thư mục `.github/workflows/` và tệp cấu hình `ci.yml` chứa triggers `push` và `pull_request` cho nhánh `main`.
- [x] 1.2 Cấu hình bước `actions/checkout@v4` và `actions/setup-go@v5` sử dụng `go-version-file: 'go.mod'` kèm bật module caching.

## 2. CI Verification Steps Implementation

- [x] 2.1 Thêm các bước kiểm tra code formatting (`gofmt -l -s .`) và static analysis (`go vet ./...`) vào workflow.
- [x] 2.2 Thêm bước thực thi unit test suite với race detector và coverage profile (`go test -v -race -coverprofile=coverage.out ./...`).
- [x] 2.3 Thêm bước biên dịch server binary (`go build -v ./cmd/server`) để đảm bảo không có lỗi build.

## 3. Workflow Validation

- [x] 3.1 Kiểm tra tính hợp lệ của tệp YAML workflow bằng cú pháp chuẩn và xác thực các câu lệnh chạy trơn tru trên môi trường Go cục bộ.
