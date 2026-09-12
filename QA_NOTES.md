# 📚 Tổng Hợp Hỏi & Đáp: Kiến Trúc Golang, Routing, DI & SOLID

Tài liệu này tổng hợp toàn bộ các câu hỏi thảo luận và giải đáp chi tiết về kiến trúc ứng dụng Go, Dependency Injection, cách tổ chức Router và nguyên lý SOLID trong dự án.

---

## 📑 Mục Lục
1. [Câu hỏi 1: Có cần tách API v1 routes trong `main.go` ra 1 file riêng không?](#1-câu-hỏi-1-có-cần-tách-api-v1-routes-trong-maingo-ra-1-file-riêng-không)
2. [Câu hỏi 2: DI (Dependency Injection) layer là gì? Tại sao phải có nó? Có cần tách ra không?](#2-câu-hỏi-2-di-dependency-injection-layer-là-gì-tại-sao-phải-có-nó-có-cần-tách-ra-không)
3. [Câu hỏi 3: Triển khai tách DI Container khi codebase mở rộng](#3-câu-hỏi-3-triển-khai-tách-di-container-khi-codebase-mở-rộng)
4. [Câu hỏi 4: Code Go có cần tuân theo nguyên lý SOLID không?](#4-câu-hỏi-4-code-go-có-cần-tuân-theo-nguyên-lý-solid-không)
5. [Câu hỏi 5: Giải thích trực quan DI trong codebase này dùng để làm gì?](#5-câu-hỏi-5-giải-thích-trực-quan-di-trong-codebase-này-dùng-để-làm-gì)

---

## 1. Câu hỏi 1: Có cần tách API v1 routes trong `main.go` ra 1 file riêng không?

### 💬 Trả lời:
**Nên tách (Rất khuyến khích)**, đặc biệt là khi dự án bắt đầu mở rộng thêm các tính năng mới.

### Lý do nên tách:
* **Single Responsibility Principle (SRP):** `main.go` chỉ nên đóng vai trò là điểm khởi động (entrypoint) - đọc config, init database, wiring dependency injection, khởi động server.
* **Khả năng mở rộng (Scalability):** Khi thêm nhiều module (`users`, `products`, `orders`, `v2`, `v3`...), `main.go` sẽ phình to hàng trăm/hàng nghìn dòng nếu không tách.
* **Tái sử dụng & Unit/Integration Test:** Dễ dàng viết test cho router/endpoint mà không cần chạy toàn bộ hàm `main()`.

### Code triển khai (`internal/router/router.go`):
```go
package router

import (
	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/container"
	"github.com/example/go-gin-jwt-auth/internal/handler"
	"github.com/example/go-gin-jwt-auth/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(c *container.Container) *gin.Engine {
	r := gin.Default()

	// Global middlewares
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status":  "ok",
			"service": "go-gin-auth-service",
		})
	})

	// API v1
	api := r.Group("/api/v1")
	{
		setupAuthRoutes(api, c.AuthHandler, c.Config)
		setupTodoRoutes(api, c.TodoHandler, c.Config)
	}

	return r
}

func setupAuthRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, cfg *config.Config) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)

		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.GET("/me", h.GetMe)
		}
	}
}

func setupTodoRoutes(rg *gin.RouterGroup, h *handler.TodoHandler, cfg *config.Config) {
	todos := rg.Group("/todos")
	todos.Use(middleware.AuthMiddleware(cfg))
	{
		todos.POST("", h.Create)
		todos.GET("", h.List)
		todos.GET("/:id", h.Get)
		todos.PATCH("/:id", h.Update)
		todos.DELETE("/:id", h.Delete)
	}
}
```

---

## 2. Câu hỏi 2: DI (Dependency Injection) layer là gì? Tại sao phải có nó? Có cần tách ra không?

### 💬 Trả lời:

### 1. DI là gì?
**Dependency Injection (Tiêm phụ thuộc)**: Thay vì một component tự khởi tạo (`new`) các đối tượng mà nó cần bên trong nó, thì bên ngoài (người gọi) sẽ tạo sẵn và "truyền vào" (inject) cho nó thông qua constructor function (`New...`).

* ❌ **Không dùng DI (Tight Coupling):**
  ```go
  type AuthService struct {}
  func (s *AuthService) Login(...) {
      db := database.Connect() // Tự tạo kết nối DB bên trong
      userRepo := repository.NewUserRepository(db)
  }
  ```
* ✅ **Dùng DI (Loose Coupling):**
  ```go
  type AuthService struct {
      userRepo  repository.UserRepository
      tokenRepo repository.TokenRepository
      cfg       *config.Config
  }
  func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, cfg *config.Config) *AuthService {
      return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, cfg: cfg}
  }
  ```

### 2. Tại sao phải có DI?
1. **Dễ viết Unit Test:** Dễ dàng truyền Mock Repository vào để test mà không cần bật Database thật.
2. **Dùng chung tài nguyên:** Một kết nối `*gorm.DB` hoặc `*config.Config` được tạo 1 lần duy nhất và chia sẻ cho tất cả repository/service.
3. **Tuân thủ Clean Architecture:** `Handler -> Service -> Repository -> Database`.

### 3. Có cần tách ra không?
* **Dự án nhỏ:** Giữ Manual DI ngay trong `main.go` là chuẩn phong cách Go (rõ ràng, không magic).
* **Dự án mở rộng (Lớn):** Tách ra `internal/container/container.go` hoặc dùng Google Wire (`google/wire`) / Uber Fx (`uber-go/fx`).

---

## 3. Câu hỏi 3: Triển khai tách DI Container khi codebase mở rộng

Để chuẩn bị cho việc mở rộng dự án lâu dài, toàn bộ khâu khởi tạo DI được chuyển vào package `internal/container`.

### File `internal/container/container.go`:
```go
package container

import (
	"gorm.io/gorm"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/handler"
	"github.com/example/go-gin-jwt-auth/internal/repository"
	"github.com/example/go-gin-jwt-auth/internal/service"
)

type Container struct {
	Config      *config.Config
	DB          *gorm.DB
	AuthHandler *handler.AuthHandler
	TodoHandler *handler.TodoHandler
}

func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
	// 1. Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	// 2. Services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	todoService := service.NewTodoService(todoRepo)

	// 3. Handlers
	authHandler := handler.NewAuthHandler(authService, cfg)
	todoHandler := handler.NewTodoHandler(todoService)

	return &Container{
		Config:      cfg,
		DB:          db,
		AuthHandler: authHandler,
		TodoHandler: todoHandler,
	}
}
```

### File `cmd/server/main.go` sau khi refactor:
```go
package main

import (
	"log"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/container"
	"github.com/example/go-gin-jwt-auth/internal/database"
	"github.com/example/go-gin-jwt-auth/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 1. Khởi tạo DI Container
	c := container.NewContainer(cfg, db)

	// 2. Setup Router
	r := router.SetupRouter(c)

	// 3. Chạy Server
	log.Printf("Starting server on port %s (env: %s)...", cfg.Port, cfg.Environment)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

---

## 4. Câu hỏi 4: Code Go có cần tuân theo nguyên lý SOLID không?

### 💬 Trả lời:
**CÓ, RẤT NÊN**, nhưng theo **"phong cách của Go" (Go-idiomatic)** chứ không nên rập khuôn theo kiểu Java/C#.

| Ký tự | Nguyên lý | Ứng dụng trong Go |
| :--- | :--- | :--- |
| **S** | **Single Responsibility** (Đơn trách nhiệm) | Mỗi package/struct/hàm chỉ làm 1 việc duy nhất (`repository` chỉ lo DB, `service` lo logic, `handler` lo HTTP). |
| **O** | **Open/Closed** (Mở rộng, đóng sửa đổi) | Dùng **Interface** và **Composition (Struct Embedding)** thay vì kế thừa (`extends`). Ví dụ: `io.Reader`, `io.Writer`. |
| **L** | **Liskov Substitution** (Thay thế Liskov) | **Implicit Interface (Duck typing)**: Struct chỉ cần hiện thực đủ methods là có thể thay thế cho Interface ở mọi nơi (VD: Mock Repo thay cho Real Repo). |
| **I** | **Interface Segregation** (Phân tách Interface) | Interface trong Go nên **càng nhỏ càng tốt** (1-2 methods). *"The bigger the interface, the weaker the abstraction"* (Rob Pike). |
| **D** | **Dependency Inversion** (Đảo ngược phụ thuộc) | High-level module phụ thuộc vào Interface, không phụ thuộc cụ thể vào DB/struct. Áp dụng qua Dependency Injection. |

> ⚠️ **Lưu ý tránh Over-engineering trong Go:** *"Accept interfaces, return structs"* — Nhận vào interface, trả về struct cụ thể. Không nên tạo Interface thừa thãi khi chưa cần mock test hay đa hình.

---

## 5. Câu hỏi 5: Giải thích trực quan DI trong codebase này dùng để làm gì?

### 💬 Trả lời:

### 1. Ẩn dụ thực tế:
* **Không dùng DI (Hàn chết linh kiện):** Động cơ bị hàn chết vào sườn xe máy. Muốn tháo động cơ ra kiểm tra riêng phải đập nát cả khung xe.
* **Có dùng DI (Lắp ghép lego / ốc vít):** Khung xe chừa sẵn các ốc bắt nối. Xưởng lắp ráp ([`container.go`](file:///C:/Users/daong/Downloads/golang/internal/container/container.go)) sẽ ráp từng bộ phận vào nhau.

### 2. Sơ đồ chuỗi tiêm phụ thuộc trong dự án:

```
[PostgreSQL Database]
        ↓ (tiêm vào)
[Repositories: UserRepo, TokenRepo, TodoRepo] + [Config]
        ↓ (tiêm vào)
[Services: AuthService, TodoService] + [Config]
        ↓ (tiêm vào)
[Handlers: AuthHandler, TodoHandler]
        ↓ (tiêm vào)
[Router: Gin Engine]
```

### 3. Ba lợi ích lớn nhất trong dự án:
1. **Dùng chung 1 kết nối Database duy nhất:** Không bị mở kết nối tràn lan gây sập database.
2. **Viết Unit Test siêu tốc:** Xem file `internal/handler/todo_handler_test.go` — Test chạy chỉ mất 0.3s bằng Mock Service mà không cần bật PostgreSQL.
3. **Dễ đổi công nghệ trong tương lai:** Nếu đổi lưu token từ PostgreSQL sang Redis, chỉ cần tạo `RedisTokenRepository` và cắm vào `AuthService` trong `container.go`, toàn bộ code `AuthService` và `AuthHandler` không cần sửa 1 dòng nào.
