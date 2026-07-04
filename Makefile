# Makefile - FastSub-Scanner Cross-Compilation
# يبني ملفات تنفيذية "نظيفة" (Static Binaries) لا تحتاج تثبيت Go على جهاز المستخدم

BINARY_NAME=fastsub-scanner
VERSION=1.0.0
BUILD_DIR=dist
LDFLAGS=-s -w -X main.version=$(VERSION)

# متغيرات لضمان بناء ثابت (Static) بدون أي اعتماد خارجي (خصوصًا مهم لـ Linux)
BUILD_ENV=CGO_ENABLED=0

.PHONY: all clean windows linux linux-arm macos macos-arm

all: clean windows linux linux-arm macos macos-arm
	@echo ""
	@echo "✅ تم بناء جميع الملفات التنفيذية في مجلد $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

clean:
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)

# ---- Windows (64-bit) ----
windows:
	@echo "🔨 بناء نسخة Windows (amd64)..."
	$(BUILD_ENV) GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go

# ---- Linux (64-bit Intel/AMD) ----
linux:
	@echo "🔨 بناء نسخة Linux (amd64)..."
	$(BUILD_ENV) GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go

# ---- Linux (ARM64 - سيرفرات حديثة / Raspberry Pi) ----
linux-arm:
	@echo "🔨 بناء نسخة Linux (arm64)..."
	$(BUILD_ENV) GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go

# ---- macOS (Intel) ----
macos:
	@echo "🔨 بناء نسخة macOS (amd64 - Intel)..."
	$(BUILD_ENV) GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 main.go

# ---- macOS (Apple Silicon M1/M2/M3) ----
macos-arm:
	@echo "🔨 بناء نسخة macOS (arm64 - Apple Silicon)..."
	$(BUILD_ENV) GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 main.go

# فحص سريع: هل الملفات التنفيذية ثابتة (Static) بدون اعتماديات ديناميكية؟
verify:
	@echo "🔍 فحص اعتماديات نسخة Linux..."
	@file $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	@ldd $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 2>&1 || echo "✅ ممتاز: الملف ثابت (Static) ولا يحتاج مكتبات خارجية"
