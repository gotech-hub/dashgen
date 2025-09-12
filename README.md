# DashGen
A powerful CLI tool and Go library for generating complete backend boilerplate code from Go struct definitions.

🚀 **Features**
- **Multi-layer Generation**: Creates repository, API handlers, action services, client SDK, and main.go
- **MongoDB Integration**: Built-in support for MongoDB with soft delete functionality
- **RESTful API**: Automatically generates CRUD endpoints with proper HTTP methods
- **Type-safe Code**: Generates fully type-safe Go code with validation
- **Template-based**: Easily customizable templates for different architectures
- **Cross-platform**: Available for Linux, macOS, and Windows

## 📦 Installation

### Option 1: Download Pre-built Binary
Download the latest release for your platform from the [Releases page](https://github.com/gotech-hub/dashgen/releases).

```bash
# Linux
wget https://github.com/gotech-hub/dashgen/releases/latest/download/dashgen-linux-amd64
chmod +x dashgen-linux-amd64
sudo mv dashgen-linux-amd64 /usr/local/bin/dashgen

# macOS
wget https://github.com/gotech-hub/dashgen/releases/latest/download/dashgen-darwin-amd64
chmod +x dashgen-darwin-amd64
sudo mv dashgen-darwin-amd64 /usr/local/bin/dashgen

# Windows
# Download dashgen-windows-amd64.exe and add to PATH
```

### Option 2: Build from Source
```bash
# Clone the repository
git clone https://github.com/gotech-hub/dashgen.git
cd dashgen

# Build the CLI tool
go build -o dashgen ./cmd/dashgen

# Install globally (optional)
sudo mv dashgen /usr/local/bin/
```

### Option 3: Quick Install Script
```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/your-org/dashgen/main/scripts/install.sh | bash

# Install specific version
curl -sSL https://raw.githubusercontent.com/your-org/dashgen/main/scripts/install.sh | VERSION=v1.0.0 bash
```

### Option 4: Using Go
```bash
go install github.com/gotech-hub/dashgen/cmd/dashgen@latest
```

### Option 5: Docker
```bash
# Run directly
docker run --rm -v $(pwd):/workspace ghcr.io/your-org/dashgen:latest --root=/workspace --module=github.com/yourorg/yourapp

# Create alias for easier usage
alias dashgen='docker run --rm -v $(pwd):/workspace ghcr.io/your-org/dashgen:latest --root=/workspace'
```

### Verify Installation
```bash
dashgen --version
```

## 🏗️ Cấu trúc dự án

DashGen sẽ tạo ra cấu trúc thư mục như sau:

```
your-project/
├── model/
│   └── user/
│       ├── data.go          # Entity definition
│       ├── init.go          # Database initialization
│       └── repository.go    # Repository interface & implementation
├── internal/
│   ├── action/
│   │   └── user.go         # Business logic services
│   └── api/
│       └── user.go         # HTTP handlers
├── client/
│   └── user.go             # SDK client methods
├── generated/
│   ├── router_user.go.snippet  # Router snippets
│   └── init_user.go.snippet    # Init snippets
└── main.go                 # Main application file
```

## 📝 Cách sử dụng

### 1. Định nghĩa Entity

Tạo file `model/user/data.go`:

```go
package user

import "time"

// @entity db:users
type User struct {
    ID        string    `json:"id" bson:"_id" validate:"required"`
    Name      string    `json:"name" bson:"name" validate:"required,min=2,max=100"`
    Email     string    `json:"email" bson:"email" validate:"required,email"`
    Age       int       `json:"age" bson:"age" validate:"min=0,max=150"`
    IsActive  bool      `json:"is_active" bson:"is_active"`
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
```

**Lưu ý quan trọng:**
- Comment `// @entity` phải đứng ngay trước type declaration
- Không được có dòng trống giữa comment và type
- Có thể chỉ định tên collection: `// @entity db:custom_table_name`

### 2. Generate Code

#### Generate từ một file cụ thể:
```bash
./dashgen --root=/path/to/project --module=github.com/yourorg/yourapp --model=/path/to/model/user/data.go
```

#### Generate từ tất cả file data.go:
```bash
./dashgen --root=/path/to/project --module=github.com/yourorg/yourapp
```

#### Dry run (xem trước không tạo file):
```bash
./dashgen --root=/path/to/project --module=github.com/yourorg/yourapp --dry
```

#### Force overwrite (ghi đè file đã tồn tại):
```bash
./dashgen --root=/path/to/project --module=github.com/yourorg/yourapp --force
```

### 3. Các tham số

| Tham số | Mô tả | Mặc định |
|---------|-------|----------|
| `--root` | Thư mục gốc của project (nơi chứa thư mục model/) | `.` |
| `--module` | Go module path (dùng cho imports) | `github.com/your-org/app` |
| `--model` | Đường dẫn đến file data.go cụ thể (tùy chọn) | - |
| `--force` | Ghi đè file đã tồn tại | `false` |
| `--dry` | Chỉ hiển thị preview, không tạo file | `false` |

## 🔧 Các file được generate

### 1. Repository (`model/user/repository.go`)
```go
type Repository interface {
    Create(data *User) (*User, error)
    GetByUserID(userID string) (*User, error)
    List(filter interface{}, offset, limit int64, sort map[string]int) ([]*User, error)
    Count(filter interface{}) (int64, error)
    UpdateByUserID(userID string, data *User) (*User, error)
    DeleteByUserID(userID string) error
}
```

### 2. API Handlers (`internal/api/user.go`)
- `CreateUser` - POST /v1/user
- `GetUserByUserID` - GET /v1/user
- `QueryUsers` - QUERY /v1/users
- `UpdateUser` - PUT /v1/user
- `DeleteUser` - DELETE /v1/user

### 3. Client SDK (`client/user.go`)
```go
func (c *BackendServiceClient) CreateUser(data *user.User) *common.APIResponse[*user.User]
func (c *BackendServiceClient) GetUser(id string) *common.APIResponse[*user.User]
// ... other methods
```

### 4. Main Application (`main.go`)
- Database initialization
- Router setup
- All CRUD endpoints registration

## 🧪 Testing

Chạy script test để kiểm tra tool:

```bash
./test.sh
```

Script sẽ:
1. Build dashgen
2. Test với sample data
3. Generate files
4. Hiển thị kết quả

## 📋 Ví dụ hoàn chỉnh

### 1. Tạo project structure
```bash
mkdir myapp
cd myapp
mkdir -p model/user
```

### 2. Tạo entity
```bash
cat > model/user/data.go << 'EOF'
package user

import "time"

// @entity db:users
type User struct {
    ID    string `json:"id" bson:"_id"`
    Name  string `json:"name" bson:"name"`
    Email string `json:"email" bson:"email"`
}
EOF
```

### 3. Generate code
```bash
dashgen --root=. --module=github.com/myorg/myapp
```

### 4. Kết quả
```
✅ Generated: model/user/init.go
✅ Generated: model/user/repository.go
✅ Generated: internal/action/user.go
✅ Generated: internal/api/user.go
✅ Generated: client/user.go
✅ Generated: main.go
```

## ⚠️ Lưu ý quan trọng

1. **Comment format**: Comment `@entity` phải đúng format và không có dòng trống
2. **File đã tồn tại**: Tool sẽ skip file đã tồn tại (trừ khi dùng `--force`)
3. **Main.go**: Sẽ được tạo mới hoặc skip nếu đã tồn tại
4. **Module path**: Phải chính xác để imports hoạt động đúng

## 🐛 Troubleshooting

### Lỗi "Found 0 entities"
- Kiểm tra comment `@entity` đúng format
- Đảm bảo không có dòng trống giữa comment và type
- Kiểm tra đường dẫn file data.go

### Lỗi template
- Kiểm tra Go version >= 1.21
- Rebuild tool: `go build -o dashgen ./cmd/dashgen`

### File không được generate
- Kiểm tra quyền ghi thư mục
- Sử dụng `--dry` để debug
- Kiểm tra đường dẫn `--root`

## 🤝 Đóng góp

1. Fork repository
2. Tạo feature branch
3. Commit changes
4. Push và tạo Pull Request

## � Deploy và Production

### 1. Build cho production
```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dashgen ./cmd/dashgen

# Hoặc build cho nhiều platform
GOOS=windows GOARCH=amd64 go build -o dashgen.exe ./cmd/dashgen
GOOS=darwin GOARCH=amd64 go build -o dashgen-mac ./cmd/dashgen
GOOS=linux GOARCH=amd64 go build -o dashgen-linux ./cmd/dashgen
```

### 2. Docker deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o dashgen ./cmd/dashgen

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dashgen .
CMD ["./dashgen"]
```

### 3. CI/CD Integration
```yaml
# GitHub Actions example
name: Generate Code
on: [push]
jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
    - name: Build dashgen
      run: go build -o dashgen ./cmd/dashgen
    - name: Generate code
      run: ./dashgen --root=. --module=${{ github.repository }}
```

## 🔧 Advanced Usage

### Custom Templates
Bạn có thể tùy chỉnh templates trong `internal/templates/templates.go`:

```go
// Sửa template để thay đổi format output
var ModelRepository = `package {{.Entity | lower}}
// Your custom template here
`
```

### Multiple Entities
```bash
# Generate cho nhiều entities cùng lúc
./dashgen --root=. --module=github.com/myorg/myapp
# Tool sẽ tự động tìm tất cả file data.go trong model/
```

### Environment Variables
```bash
# Có thể sử dụng env vars
export DASHGEN_MODULE="github.com/myorg/myapp"
export DASHGEN_ROOT="/path/to/project"
./dashgen --module=$DASHGEN_MODULE --root=$DASHGEN_ROOT
```

## 📊 Performance Tips

1. **Batch Generation**: Generate nhiều entities cùng lúc thay vì từng cái một
2. **Skip Existing**: Tool tự động skip file đã tồn tại để tăng tốc độ
3. **Dry Run**: Sử dụng `--dry` để test trước khi generate thực sự

## �📄 License

MIT License - xem file LICENSE để biết thêm chi tiết.
