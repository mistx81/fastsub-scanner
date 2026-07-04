@echo off
REM build.bat - بناء FastSub-Scanner لجميع الأنظمة من Windows CMD
REM يتطلب فقط تثبيت Go على جهاز البناء (المطوّر) - وليس على جهاز المستخدم النهائي

setlocal enabledelayedexpansion

set BINARY_NAME=fastsub-scanner
set VERSION=1.0.0
set BUILD_DIR=dist
set LDFLAGS=-s -w -X main.version=%VERSION%

echo تنظيف مجلد البناء السابق...
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
mkdir %BUILD_DIR%

echo.
echo بناء نسخة Windows (amd64)...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags="%LDFLAGS%" -o %BUILD_DIR%\%BINARY_NAME%-windows-amd64.exe main.go

echo بناء نسخة Linux (amd64)...
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags="%LDFLAGS%" -o %BUILD_DIR%\%BINARY_NAME%-linux-amd64 main.go

echo بناء نسخة Linux (arm64)...
set GOOS=linux
set GOARCH=arm64
go build -trimpath -ldflags="%LDFLAGS%" -o %BUILD_DIR%\%BINARY_NAME%-linux-arm64 main.go

echo بناء نسخة macOS (amd64 - Intel)...
set GOOS=darwin
set GOARCH=amd64
go build -trimpath -ldflags="%LDFLAGS%" -o %BUILD_DIR%\%BINARY_NAME%-macos-amd64 main.go

echo بناء نسخة macOS (arm64 - Apple Silicon)...
set GOOS=darwin
set GOARCH=arm64
go build -trimpath -ldflags="%LDFLAGS%" -o %BUILD_DIR%\%BINARY_NAME%-macos-arm64 main.go

echo.
echo اكتمل البناء! الملفات موجودة في مجلد %BUILD_DIR%
dir %BUILD_DIR%

endlocal
