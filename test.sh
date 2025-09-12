#!/bin/bash

echo "🔨 Building dashgen..."
go build -o dashgen ./cmd/dashgen

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""

echo "🧪 Testing with sample data (dry run)..."
./dashgen --root=testdata --module=gitlab.silvertiger.tech/open-platform/backend-app --dry

echo ""
echo "📝 Generating actual files..."
./dashgen --root=testdata --module=gitlab.silvertiger.tech/open-platform/backend-app

echo ""
echo "📁 Generated files:"
find testdata -name "*.go" -not -path "*/data.go" | sort
echo ""
echo "📄 Generated snippets for main.go:"
find testdata -name "*.snippet" | sort

echo ""
echo "🎉 Test completed successfully!"
echo ""
echo "Usage examples:"
echo "  # Generate from all data.go files in model/ directory:"
echo "  ./dashgen -root /path/to/project -module github.com/yourorg/yourapp"
echo ""
echo "  # Generate from a single data.go file:"
echo "  ./dashgen -model /path/to/model/user/data.go -module github.com/yourorg/yourapp"
echo ""
echo "  # Dry run (preview only):"
echo "  ./dashgen -root /path/to/project -module github.com/yourorg/yourapp -dry"
echo ""
echo "  # Force overwrite existing files:"
echo "  ./dashgen -root /path/to/project -module github.com/yourorg/yourapp -force"
