@echo off
setlocal enabledelayedexpansion

echo PrivacyCheck Go Build Script
echo ============================

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed or not in PATH
    exit /b 1
)

echo Go version:
go version

REM Clean previous builds
echo.
echo Cleaning previous builds...
if exist "dist" rmdir /s /q "dist"
mkdir "dist"

REM Download dependencies
echo.
echo Downloading dependencies...
go mod tidy
if errorlevel 1 (
    echo Error: Failed to download dependencies
    exit /b 1
)

REM Get build info
for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown
for /f "tokens=*" %%i in ('powershell -Command "Get-Date -Format 'yyyy-MM-dd_HH-mm-ss'"') do set BUILD_DATE=%%i

set LDFLAGS=-s -w -X "privacycheck/core.BuildDate=%BUILD_DATE%" -X "privacycheck/core.GitCommit=%GIT_COMMIT%"

REM Build for Windows x64
echo.
echo Building for Windows x64...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags "%LDFLAGS%" -o "dist/privacycheck-windows-x64.exe" .
if errorlevel 1 (
    echo Error: Failed to build for Windows x64
    exit /b 1
)

REM Build for Linux x64
echo.
echo Building for Linux x64...
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags "%LDFLAGS%" -o "dist/privacycheck-linux-x64" .
if errorlevel 1 (
    echo Error: Failed to build for Linux x64
    exit /b 1
)

REM Show build results
echo.
echo Build completed successfully!
echo.
echo Built files:
dir /b "dist"

echo.
echo File sizes:
for %%f in (dist\*) do (
    echo %%f: %%~zf bytes
)

echo.
echo Build script completed.
pause
