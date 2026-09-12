# 🎓 CẨM NANG HỌC GO TỪ CON SỐ 0 QUA DỰ ÁN AUTHENTICATION

> Tài liệu này được thiết kế riêng cho **người mới bắt đầu (Beginner)** học Golang. Hướng dẫn giải thích cặn kẽ từng khái niệm cốt lõi, cú pháp Go, tư duy thiết kế, và cơ chế bảo mật được áp dụng trong toàn bộ dự án này.

---

## 🗺️ MỤC LỤC BÀI HỌC

1. [Các khái niệm Go cốt lõi trong dự án](#1-các-khái-niệm-go-cốt-lõi-trong-dự-án)
   - Package & Thư mục đặc biệt `internal/`
   - Struct & Struct Tags (`json`, `gorm`, `binding`)
   - Con trỏ (Pointers: `*` và `&`)
   - Interface & Tính đa hình (Polymorphism)
   - Constructor Pattern (`New...`) & Dependency Injection
   - Xử lý lỗi (Error Handling: `if err != nil`, `errors.Is`)
2. [Làm chủ Gin Web Framework](#2-làm-chủ-gin-web-framework)
   - `gin.Context` là gì?
   - Data Binding (`ShouldBindJSON`)
   - Middleware trong Gin hoạt động ra sao? (`c.Next()`, `c.Abort()`)
   - Cookie Management (`HttpOnly`, `SameSite`, `MaxAge`)
3. [Làm việc với Database & GORM](#3-làm-việc-với-database--gorm)
   - Chuỗi kết nối DSN
   - GORM Hooks (`BeforeCreate`)
   - Các thao tác ORM cơ bản
4. [Bảo mật chuyên sâu (Security & Auth Deep Dive)](#4-bảo-mật-chuyên-sâu)
   - Tại sao hash mật khẩu phải dùng Bcrypt?
   - Cấu tạo và cơ chế ký số của JWT
   - LocalStorage vs. HttpOnly Cookie
   - Refresh Token Rotation (RTR) & Token Family
5. [Đọc & Hiểu toàn bộ luồng code theo từng file](#5-đọc--hiểu-toàn-bộ-luồng-code-từng-file)

---

## 1. CÁC KHÁI NIỆM GO CỐT LÕI TRONG DỰ ÁN

### 1.1. Package & Thư mục đặc biệt `internal/`
- Trong Go, mọi file code bắt đầu bằng `package <tên_package>`.
- **Quy tắc viết hoa chữ cái đầu (Exported vs. Unexported)**:
  - Nếu một hàm, biến hoặc struct bắt đầu bằng **chữ hoa** (ví dụ: `Register`, `User`, `Config`), nó là **Public (Exported)** — các package khác có thể gọi được.
  - Nếu bắt đầu bằng **chữ thường** (ví dụ: `authService`, `userRepo`, `getEnv`), nó là **Private (Unexported)** — chỉ dùng trong nội bộ file/package đó.
- **Thư mục `internal/` trong Go**:
  - Đây là cơ chế đặc biệt của Go. Bất kỳ code nào đặt trong thư mục `internal/` thì chỉ có các package bên trong dự án này được phép import. Các dự án bên ngoài không thể import được.

---

### 1.2. Struct & Struct Tags
Struct trong Go tương đương với `Class` hoặc `Object Schema` trong các ngôn ngữ khác.

Hãy xem file [`internal/model/user.go`](file:///C:/Users/daong/Downloads/golang/internal/model/user.go):
```go
type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
    Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
    PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
}
```

**Struct Tag là các đoạn text nằm trong dấu backtick \`...\`:**
- `json:"id"`: Khi serialize struct này thành JSON trả về cho Client, trường `ID` sẽ đổi tên thành `"id"`.
- `json:"-"`: **Ẩn trường này khi chuyển thành JSON**. Cực kỳ quan trọng để không bao giờ để lộ mật khẩu `PasswordHash` ra ngoài response API!
- `gorm:"..."`: Hướng dẫn GORM tạo cột trong PostgreSQL với kiểu dữ liệu, ràng buộc unique hay khóa chính.
- `binding:"required,email"`: (Ở file handler) Hướng dẫn Gin tự động kiểm tra định dạng email và không được để trống.

---

### 1.3. Con trỏ (Pointers: `*` và `&`)
Trong Go, khi bạn truyền một biến vào hàm, Go sẽ **sao chép (copy)** giá trị của biến đó.
- `&` (Address-of operator): Lấy địa chỉ ô nhớ của biến (tạo con trỏ).
- `*` (Pointer type / Dereference): Kiểu dữ liệu con trỏ hoặc đọc giá trị tại địa chỉ ô nhớ.

**Ví dụ:**
```go
func (r *userRepository) Create(user *model.User) error
```
- `user *model.User`: Truyền địa chỉ của struct `User` vào hàm thay vì copy toàn bộ struct. Việc này giúp tiết kiệm bộ nhớ và cho phép hàm `Create` sửa đổi trực tiếp dữ liệu (ví dụ GORM gán `ID` mới vào biến `user` ban đầu).
- Trả về `*model.User, error`: Nếu có lỗi, trả về con trỏ `nil` (tương đương `null` trong JS/Java) và `err`.

---

### 1.4. Interface & Tính đa hình (Polymorphism)
Trong Go, Interface không cần từ khóa `implements`. Một struct tự động thỏa mãn một Interface nếu nó có đủ tất cả các method mà Interface đó định nghĩa (**Duck Typing**).

Xem [`internal/repository/user_repository.go`](file:///C:/Users/daong/Downloads/golang/internal/repository/user_repository.go):
```go
// 1. Khai báo Interface (Hợp đồng)
type UserRepository interface {
    Create(user *model.User) error
    FindByEmail(email string) (*model.User, error)
    FindByID(id uuid.UUID) (*model.User, error)
}

// 2. Struct thực thi interface này
type userRepository struct {
    db *gorm.DB
}

// 3. Viết các method cho struct
func (r *userRepository) Create(user *model.User) error { ... }
func (r *userRepository) FindByEmail(email string) (*model.User, error) { ... }
func (r *userRepository) FindByID(id uuid.UUID) (*model.User, error) { ... }
```
**Tại sao phải dùng Interface?**
- Giúp tầng Service không phụ thuộc vào GORM. Khi viết Unit Test, bạn có thể tạo một `mockUserRepository` giả lập mà không cần kết nối database thật.

---

### 1.5. Constructor Pattern (`New...`) & Dependency Injection
Go không có hàm khởi tạo (`constructor`) tự động như `constructor()` trong JavaScript hay `__init__` trong Python. Thay vào đó, lập trình viên Go tạo hàm bắt đầu bằng chữ `New...`:

```go
func NewAuthService(
    userRepo repository.UserRepository,
    tokenRepo repository.TokenRepository,
    cfg *config.Config,
) AuthService {
    return &authService{
        userRepo:  userRepo,
        tokenRepo: tokenRepo,
        cfg:       cfg,
    }
}
```
Trong [`main.go`](file:///C:/Users/daong/Downloads/golang/cmd/server/main.go), ta khởi tạo từ dưới lên:
`DB` ➡️ `Repositories` ➡️ `Services` ➡️ `Handlers` ➡️ `Router`. Đây chính là **Dependency Injection** thủ công, cực kỳ tường minh và dễ debug.

---

### 1.6. Xử lý lỗi (Error Handling)
Go không sử dụng `try / catch / throw`. Mọi hàm có khả năng phát sinh lỗi đều trả về giá trị `error` là phần tử cuối cùng:

```go
user, err := s.userRepo.FindByEmail(email)
if err != nil {
    return nil, err // Nếu có lỗi, dừng lại và trả lỗi về tầng trên
}
```

Để phân biệt các loại lỗi cụ thể, dự án dùng `errors.New` và kiểm tra bằng `errors.Is`:
```go
// Khai báo lỗi sẵn (Sentinel Errors)
var ErrUserAlreadyExists = errors.New("user with this email already exists")

// Kiểm tra ở Handler
if errors.Is(err, service.ErrUserAlreadyExists) {
    c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // Trả về 409 Conflict
    return
}
```

---

## 2. LÀM CHỦ GIN WEB FRAMEWORK

### 2.1. `gin.Context` là gì?
`c *gin.Context` là trái tim của Gin. Nó chứa toàn bộ thông tin về HTTP Request hiện tại (Headers, Body, Cookies, Params) và cung cấp các hàm để tạo HTTP Response.

### 2.2. Data Binding (`c.ShouldBindJSON`)
Thay vì phải tự đọc body chuỗi rồi parse JSON:
```go
var req RegisterRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```
Gin tự động đọc JSON từ body, ép kiểu vào struct `RegisterRequest`, và kiểm tra validation (`binding:"required,email"`).

### 2.3. Middleware trong Gin
Middleware là các hàm trung gian chạy trước khi request đến Handler (Controller):
- `c.Next()`: Cho phép request đi tiếp đến middleware tiếp theo hoặc handler.
- `c.Abort()` / `c.AbortWithStatus(...)`: Chặn đứng request ngay lập tức (ví dụ token không hợp lệ).
- `c.Set("userID", claims.UserID)`: Gán dữ liệu vào context để các handler phía sau lấy ra dùng qua `c.Get("userID")`.

### 2.4. Cookie Management
Xem cách lưu Refresh Token trong [`auth_handler.go`](file:///C:/Users/daong/Downloads/golang/internal/handler/auth_handler.go):
```go
c.SetCookie(
    "refresh_token",  // Tên cookie
    tokenString,      // Giá trị token
    maxAge,           // Thời gian sống (giây)
    "/",              // Path có hiệu lực
    "",               // Domain
    false,            // Secure (true nếu dùng HTTPS)
    true,             // HttpOnly = TRUE (JavaScript không thể đọc được!)
)
```

---

## 3. LÀM VIỆC VỚI DATABASE & GORM

### 3.1. GORM Hooks (`BeforeCreate`)
Hook là hàm tự động chạy trước hoặc sau một sự kiện database:
```go
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
    if u.ID == uuid.Nil {
        u.ID = uuid.New() // Tự động sinh UUID mới trước khi INSERT vào bảng
    }
    return nil
}
```

### 3.2. Các thao tác ORM cơ bản
- **Thêm mới**: `r.db.Create(user).Error`
- **Tìm kiếm 1 bản ghi**: `r.db.Where("email = ?", email).First(&user).Error`
- **Cập nhật 1 trường**: `r.db.Model(&model.RefreshToken{}).Where("id = ?", id).Update("is_revoked", true).Error`
- **Xóa**: `r.db.Where("expires_at < ?", time.Now()).Delete(&model.RefreshToken{}).Error`
- **Tìm kiếm kèm quan hệ (Join/Preload)**: `r.db.Preload("User").Where("token_hash = ?", hash).First(&token).Error`

---

## 4. BẢO MẬT CHUYÊN SÂU

### 4.1. Tại sao hash mật khẩu bằng Bcrypt?
- Các thuật toán như MD5, SHA-1, SHA-256 chạy rất nhanh (hàng triệu phép tính/giây), hacker có thể dùng GPU thử hàng tỷ mật khẩu (Brute-force / Rainbow table).
- **Bcrypt**:
  - Là thuật toán **chậm có chủ đích (Slow Hash)**.
  - Tự động sinh chuỗi **Salt (muối ngẫu nhiên)** gộp vào chuỗi hash. Hai người dùng đặt mật khẩu giống nhau nhưng chuỗi hash trong database vẫn hoàn toàn khác nhau.
  - Hàm `bcrypt.CompareHashAndPassword` an toàn trước tấn công phân tích thời gian (Timing Attack).

---

### 4.2. JWT (JSON Web Token)
JWT gồm 3 phần ngăn cách bởi dấu chấm `.`:
`Header.Payload.Signature`
- **Header**: Thuật toán ký (ví dụ: `HS256`).
- **Payload**: Thông tin người dùng (`user_id`, `email`, `exp` - thời hạn hết hạn). Lưu ý: Payload chỉ được mã hóa Base64 chứ **không mã hóa bí mật**, không được để thông tin nhạy cảm (như mật khẩu) vào đây.
- **Signature**: Chữ ký số tạo từ `HMACSHA256(Base64(Header) + "." + Base64(Payload), SecretKey)`. Server dùng `SecretKey` để đảm bảo Payload không bị chỉnh sửa giả mạo.

---

### 4.3. Tại sao Refresh Token phải lưu ở `HttpOnly Cookie`?
- Nếu lưu Token ở `localStorage` của trình duyệt: Bất kỳ mã JavaScript độc hại nào (tấn công XSS - Cross-Site Scripting) cũng có thể đọc được `localStorage.getItem('token')` và gửi về cho hacker.
- Khi bật cờ **`HttpOnly = true`**: Trình duyệt **cấm tuyệt đối mã JavaScript đọc cookie này**. Cookie sẽ tự động được trình duyệt đính kèm vào mỗi request gửi lên server mà không ai có thể can thiệp từ JS.

---

### 4.4. Refresh Token Rotation (RTR) & Token Family
**Vấn đề:** Nếu Refresh Token có hạn 7 ngày và bị hacker đánh cắp một lần, hacker có thể dùng nó suốt 7 ngày.
**Giải pháp - RTR:**
1. Mỗi khi người dùng gọi API `/refresh`, server sẽ **hủy (revoke)** refresh token hiện tại.
2. Server cấp một cặp token mới toanh (Access Token mới + Refresh Token mới).
3. Các token thuộc cùng một phiên đăng nhập sẽ có chung **`family_id`**.
4. **Phát hiện tái sử dụng (Reuse Detection)**: Nếu một token đã bị thu hồi (`is_revoked = true`) đột nhiên được gửi lên, server hiểu ngay rằng **token này đã bị đánh cắp** và kẻ trộm đang cố dùng lại nó. Lúc này server lập tức **hủy toàn bộ chuỗi token có cùng `family_id`**, ép tất cả các phiên phải đăng nhập lại.

---

## 5. ĐỌC & HIỂU TOÀN BỘ LUỒNG CODE TỪNG FILE

Để bắt đầu đọc hiểu dự án, hãy đi theo đúng thứ tự sau:

1. 📄 **[`internal/model/user.go`](file:///C:/Users/daong/Downloads/golang/internal/model/user.go)** & **[`refresh_token.go`](file:///C:/Users/daong/Downloads/golang/internal/model/refresh_token.go)**: Xem cấu trúc dữ liệu.
2. 📄 **[`internal/utils/password.go`](file:///C:/Users/daong/Downloads/golang/internal/utils/password.go)** & **[`jwt.go`](file:///C:/Users/daong/Downloads/golang/internal/utils/jwt.go)**: Xem các hàm tiện ích băm mật khẩu và ký JWT.
3. 📄 **[`internal/repository/user_repository.go`](file:///C:/Users/daong/Downloads/golang/internal/repository/user_repository.go)** & **[`token_repository.go`](file:///C:/Users/daong/Downloads/golang/internal/repository/token_repository.go)**: Xem cách truy vấn dữ liệu từ database.
4. 📄 **[`internal/service/auth_service.go`](file:///C:/Users/daong/Downloads/golang/internal/service/auth_service.go)**: Xem cách kết hợp repository và logic nghiệp vụ.
5. 📄 **[`internal/middleware/auth_middleware.go`](file:///C:/Users/daong/Downloads/golang/internal/middleware/auth_middleware.go)**: Xem cách chặn các request chưa đăng nhập.
6. 📄 **[`internal/handler/auth_handler.go`](file:///C:/Users/daong/Downloads/golang/internal/handler/auth_handler.go)**: Xem cách parse request và trả cookie/JSON.
7. 📄 **[`cmd/server/main.go`](file:///C:/Users/daong/Downloads/golang/cmd/server/main.go)**: Xem cách ráp tất cả các mảnh ghép lại với nhau và khởi động server.
