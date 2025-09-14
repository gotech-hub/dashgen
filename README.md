# DashGen
A powerful CLI tool and Go library for generating complete backend boilerplate code from Go struct definitions.

🚀 **Features**
- **Multi-layer Generation**: Creates repository, API handlers, action services, and client SDK
- **MongoDB Integration**: Built-in support for MongoDB with soft delete functionality
- **RESTful API**: Automatically generates CRUD endpoints with proper HTTP methods
- **Smart Validation**: Auto-generates validation code from struct tags (`validate:"required,email,min=2"`)
- **Index Management**: Auto-generates MongoDB indexes from struct tags and comments
- **Type-safe Code**: Generates fully type-safe Go code with comprehensive error handling
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

# If 'dashgen' command not found, add Go bin to PATH:
export PATH=$PATH:$(go env GOPATH)/bin

# Or add permanently to your shell profile:
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc  # for zsh
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc # for bash
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

## 🏗️ Project Structure

DashGen generates the following directory structure:

```
your-project/
├── model/
│   └── user/
│       ├── data.go          # Entity definition
│       ├── init.go          # Database initialization with indexes
│       └── repository.go    # Repository interface & implementation
├── internal/
│   ├── action/
│   │   └── user.go         # Business logic services
│   └── api/
│       └── user.go         # HTTP handlers with validation
└── client/
    └── user.go             # SDK client methods
```

**Note**: DashGen no longer generates or modifies `main.go` files. This allows the library to be used in existing projects without interfering with your main application setup.

## 📝 Usage

### 1. Define Entity

Create file `model/user/data.go`:

```go
package user

import "time"

// @entity db:users
// @index email:1 unique
// @index name:1,created_at:-1
// @index email:text
type User struct {
    ID        string    `json:"id" bson:"_id" validate:"required"`
    Name      string    `json:"name" bson:"name" validate:"required,min=2,max=100" index:"1"`
    Email     string    `json:"email" bson:"email" validate:"required,email" index:"unique"`
    Age       int       `json:"age" bson:"age" validate:"min=0,max=150"`
    IsActive  bool      `json:"is_active" bson:"is_active" index:"1"`
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
```

**Important Notes:**
- Comment `// @entity` must be placed directly before the type declaration
- No blank lines allowed between comment and type
- You can specify collection name: `// @entity db:custom_table_name`

### 2. Validation Tags

DashGen supports automatic validation code generation from struct tags:

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field is required | `validate:"required"` |
| `min=N` | Minimum length/value | `validate:"min=2"` |
| `max=N` | Maximum length/value | `validate:"max=100"` |
| `email` | Valid email format | `validate:"email"` |

**Combined validation**: `validate:"required,email,min=5,max=100"`

### 3. Index Definitions

#### Field-level Indexes (via struct tags):
```go
type User struct {
    Name  string `index:"1"`        // Ascending index
    Email string `index:"unique"`   // Unique index
    Tags  string `index:"text"`     // Text index
    Score int    `index:"-1"`       // Descending index
}
```

#### Compound Indexes (via comments):
```go
// @index field1:1,field2:-1 unique sparse name:custom_name
// @index email:1 unique
// @index name:text
type User struct {
    // ... fields
}
```

**Index Options:**
- `unique` - Creates unique index
- `sparse` - Creates sparse index
- `name:custom_name` - Sets custom index name
- Field directions: `1` (ascending), `-1` (descending)
- Special types: `text`, `2dsphere`, etc.

### 4. Generate Code

#### Generate from specific file:
```bash
./dashgen --root=/path/to/project --module=github.com/yourorg/yourapp --model=/path/to/model/user/data.go
```

#### Generate from all data.go files:
```bash
./dashgen --root=/path/to/project --module=github.com/yourorg/yourapp
```

#### Dry run (preview without creating files):
```bash
./dashgen --root=/path/to/project --module=github.com/yourorg/yourapp --dry
```

#### Force overwrite existing files:
```bash
./dashgen --root=/path/to/project --module=github.com/yourorg/yourapp --force
```

### 5. Command Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--root` | Project root directory (containing model/ folder) | `.` |
| `--module` | Go module path (used for imports) | `github.com/your-org/app` |
| `--model` | Path to specific data.go file (optional) | - |
| `--force` | Overwrite existing files | `false` |
| `--dry` | Show preview only, don't create files | `false` |

## 🔧 Generated Files

### 1. Database Initialization (`model/user/init.go`)
```go
func Init(database *mongo.Database) error {
    // Collection setup
    userCollection = collection.NewMongoDBGenericCollection[User]("users")
    userCollection.SetDatabase(database)

    // Create indexes automatically
    if err := createIndexes(); err != nil {
        return err
    }
    return nil
}

func createIndexes() error {
    // Auto-generated index creation code
    // Based on struct tags and @index comments
    err := userCollection.CreateIndex(bson.D{
        {Key: "email", Value: 1},
    }, &options.IndexOptions{
        Unique: utils.GetPointer(true),
    })
    // ... more indexes
    return nil
}
```

### 2. Repository (`model/user/repository.go`)
```go
type Repository interface {
    Create(data *User) (*User, error)
    GetByUserID(userID string) (*User, error)
    List(filter interface{}, offset, limit int64, sort map[string]int) ([]*User, error)
    Count(filter interface{}) (int64, error)
    UpdateByUserID(userID string, data *User) (*User, error)
    DeleteByUserID(userID string) error // Soft delete
}
```

### 3. API Handlers with Validation (`internal/api/user.go`)
```go
func CreateUser(req request.APIRequest, res responder.APIResponder) error {
    var userData user.User
    if err := req.ParseBody(&userData); err != nil {
        return res.Respond(common.NewErrorResponse(...))
    }

    // Auto-generated validation code
    if userData.Name == "" {
        return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid,
            "VALIDATION_FAILED", "name is required"))
    }
    if len(userData.Name) < 2 {
        return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid,
            "VALIDATION_FAILED", "name must be at least 2 characters"))
    }
    if userData.Email != "" && !isValidEmail(userData.Email) {
        return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid,
            "VALIDATION_FAILED", "email must be a valid email address"))
    }

    response := action.CreateUser(&userData)
    return res.Respond(response)
}
```

**API Endpoints:**
- `CreateUser` - POST /v1/user
- `GetUserByUserID` - GET /v1/user
- `QueryUsers` - QUERY /v1/users
- `UpdateUser` - PUT /v1/user
- `DeleteUser` - DELETE /v1/user

### 4. Client SDK (`client/user.go`)
```go
func (c *BackendServiceClient) CreateUser(data *user.User) *common.APIResponse[*user.User]
func (c *BackendServiceClient) GetUser(id string) *common.APIResponse[*user.User]
func (c *BackendServiceClient) ListUsers(query *common.Query[user.User]) *common.APIResponse[*user.User]
func (c *BackendServiceClient) UpdateUser(id string, data *user.User) *common.APIResponse[*user.User]
func (c *BackendServiceClient) DeleteUser(id string) *common.APIResponse[any]
```

## 🧪 Testing

Run the test script to verify the tool:

```bash
./test.sh
```

The script will:
1. Build dashgen
2. Test with sample data
3. Generate files
4. Display results

## 📋 Complete Example

### 1. Create project structure
```bash
mkdir myapp
cd myapp
mkdir -p model/user
```

### 2. Create entity with validation and indexes
```bash
cat > model/user/data.go << 'EOF'
package user

import "time"

// @entity db:users
// @index email:1 unique
// @index name:1,created_at:-1
type User struct {
    ID        string    `json:"id" bson:"_id" validate:"required"`
    Name      string    `json:"name" bson:"name" validate:"required,min=2,max=100" index:"1"`
    Email     string    `json:"email" bson:"email" validate:"required,email" index:"unique"`
    Age       int       `json:"age" bson:"age" validate:"min=0,max=150"`
    IsActive  bool      `json:"is_active" bson:"is_active"`
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
EOF
```

### 3. Generate code
```bash
dashgen --root=. --module=github.com/myorg/myapp
```

### 4. Results
```
✅ Generated: model/user/init.go
✅ Generated: model/user/repository.go
✅ Generated: internal/action/user.go
✅ Generated: internal/api/user.go
✅ Generated: client/user.go
✅ Generation finished.
```

### 5. Integration in your main.go
```go
package main

import (
    "log"
    "go.mongodb.org/mongo-driver/mongo"
    "gitlab.silvertiger.tech/go-sdk/go-mongodb/client"
    "github.com/myorg/myapp/model/user"
)

func main() {
    // Setup MongoDB connection
    mongoClient := client.NewMongoClient("myapp", config, onDBConnected)
    err := mongoClient.Connect()
    if err != nil {
        log.Fatal(err)
    }
}

func onDBConnected(database *mongo.Database) error {
    // Initialize all your entities
    return user.Init(database)
}
```

## ⚠️ Important Notes

1. **Comment format**: Comment `@entity` must be in correct format with no blank lines
2. **Existing files**: Tool will skip existing files (unless using `--force`)
3. **Main.go**: DashGen no longer generates or modifies main.go files
4. **Module path**: Must be accurate for imports to work correctly
5. **Validation**: Only basic validation types are supported (required, min, max, email)
6. **Indexes**: Field-level and compound indexes are automatically created during Init()

## 🐛 Troubleshooting

### Command 'dashgen' not found after `go install`
This happens when Go bin directory is not in PATH.

**Solution:**
```bash
# Check GOPATH
go env GOPATH

# Add to PATH temporarily
export PATH=$PATH:$(go env GOPATH)/bin

# Add to PATH permanently
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc  # for zsh
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc # for bash

# Reload shell
source ~/.zshrc  # or ~/.bashrc

# Verify
which dashgen
dashgen --version
```

### Error "Found 0 entities"
- Check `@entity` comment format is correct
- Ensure no blank lines between comment and type
- Verify data.go file path

### Template errors
- Check Go version >= 1.24
- Rebuild tool: `go install github.com/gotech-hub/dashgen/cmd/dashgen@latest`

### Files not generated
- Check directory write permissions
- Use `--dry` flag to debug
- Verify `--root` path is correct

### Validation not working
- Ensure validate tags are properly formatted
- Check that validation logic is imported in your API handlers
- Verify field types are supported (string, int, int32, int64)

### Index creation fails
- Check MongoDB connection is established before calling Init()
- Verify field names in index definitions match struct fields
- Ensure index syntax is correct: `field:1` or `field:-1`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push and create a Pull Request

## 🚀 Advanced Features

### Custom Validation Rules
You can extend validation by adding custom rules in the generated API handlers:

```go
// Add custom validation after auto-generated validation
if userData.Age > 0 && userData.Age < 13 {
    return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid,
        "VALIDATION_FAILED", "age must be 13 or older"))
}
```

### Complex Index Patterns
```go
// @entity db:products
// @index category:1,price:-1,created_at:-1 name:category_price_date
// @index name:text,description:text name:search_index
// @index location:2dsphere
type Product struct {
    Category    string    `json:"category" bson:"category"`
    Price       float64   `json:"price" bson:"price"`
    Name        string    `json:"name" bson:"name"`
    Description string    `json:"description" bson:"description"`
    Location    []float64 `json:"location" bson:"location"` // [lng, lat]
    CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}
```

### Environment-based Configuration
```bash
# Use environment variables for common settings
export DASHGEN_MODULE="github.com/myorg/myapp"
export DASHGEN_ROOT="/path/to/project"
dashgen --module=$DASHGEN_MODULE --root=$DASHGEN_ROOT
```

## 🔧 Build and Deployment

### 1. Build for production
```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dashgen ./cmd/dashgen

# Cross-platform builds
GOOS=windows GOARCH=amd64 go build -o dashgen.exe ./cmd/dashgen
GOOS=darwin GOARCH=amd64 go build -o dashgen-mac ./cmd/dashgen
GOOS=linux GOARCH=amd64 go build -o dashgen-linux ./cmd/dashgen
```

### 2. Docker deployment
```dockerfile
FROM golang:1.24-alpine AS builder
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
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.24'
    - name: Build dashgen
      run: go build -o dashgen ./cmd/dashgen
    - name: Generate code
      run: ./dashgen --root=. --module=${{ github.repository }}
```

## 🎯 Best Practices

### 1. Entity Design
```go
// Good: Clear, consistent naming
// @entity db:users
// @index email:1 unique
// @index created_at:-1
type User struct {
    ID        string    `json:"id" bson:"_id" validate:"required"`
    Email     string    `json:"email" bson:"email" validate:"required,email" index:"unique"`
    Name      string    `json:"name" bson:"name" validate:"required,min=2,max=100"`
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
```

### 2. Validation Strategy
- Use `required` for mandatory fields
- Set reasonable `min` and `max` limits
- Use `email` for email fields
- Combine multiple rules: `validate:"required,email,min=5"`

### 3. Index Strategy
- Create unique indexes for unique fields (email, username)
- Add compound indexes for common query patterns
- Use text indexes for search functionality
- Consider sparse indexes for optional fields

### 4. Project Organization
```
project/
├── model/
│   ├── user/data.go
│   ├── product/data.go
│   └── order/data.go
├── internal/
│   ├── api/          # Generated API handlers
│   └── action/       # Generated business logic
└── client/           # Generated SDK
```

## 📊 Performance Tips

1. **Batch Generation**: Generate multiple entities at once instead of one by one
2. **Skip Existing**: Tool automatically skips existing files for faster execution
3. **Dry Run**: Use `--dry` flag to test before actual generation
4. **Selective Generation**: Use `--model` flag to generate specific entities only

## 🔍 Validation Reference

| Rule | Type Support | Example | Generated Code |
|------|-------------|---------|----------------|
| `required` | string, int | `validate:"required"` | Checks for empty/zero values |
| `min=N` | string, int | `validate:"min=2"` | Length/value minimum check |
| `max=N` | string, int | `validate:"max=100"` | Length/value maximum check |
| `email` | string | `validate:"email"` | Email format validation |

## 🗂️ Index Reference

| Type | Syntax | Example | Description |
|------|--------|---------|-------------|
| Ascending | `index:"1"` | `Name string \`index:"1"\`` | Single field ascending |
| Descending | `index:"-1"` | `Date time.Time \`index:"-1"\`` | Single field descending |
| Unique | `index:"unique"` | `Email string \`index:"unique"\`` | Unique constraint |
| Text | `index:"text"` | `Content string \`index:"text"\`` | Full-text search |
| Compound | `@index field1:1,field2:-1` | See examples above | Multiple fields |

## 📄 License

MIT License - see LICENSE file for details.
