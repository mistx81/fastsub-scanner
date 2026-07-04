#!/usr/bin/env bash
# build.sh - بناء FastSub-Scanner لجميع الأنظمة دفعة واحدة
# استخدم هذا السكربت إذا لم يكن "make" متوفرًا على جهازك (مثلاً على Windows بدون Git Bash / WSL)

set -e

BINARY_NAME="fastsub-scanner"
VERSION="1.0.0"
BUILD_DIR="dist"
LDFLAGS="-s -w -X main.version=${VERSION}"

echo "🧹 تنظيف مجلد البناء السابق..."
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# مصفوفة الأنظمة والمعماريات المطلوبة: GOOS/GOARCH/اسم-الملف-الناتج
targets=(
  "windows amd64 ${BINARY_NAME}-windows-amd64.exe"
  "linux   amd64 ${BINARY_NAME}-linux-amd64"
  "linux   arm64 ${BINARY_NAME}-linux-arm64"
  "darwin  amd64 ${BINARY_NAME}-macos-amd64"
  "darwin  arm64 ${BINARY_NAME}-macos-arm64"
)

for target in "${targets[@]}"; do
  read -r os arch out <<< "${target}"
  echo "🔨 بناء ${os}/${arch} -> ${BUILD_DIR}/${out}"
  CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" go build -trimpath -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/${out}" main.go
done

echo ""
echo "✅ اكتمل بناء جميع الملفات التنفيذية:"
ls -la "${BUILD_DIR}/"
