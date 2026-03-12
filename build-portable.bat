@echo off
setlocal
REM Build script to create a portable Windows executable.
REM On Windows ARM64, Go cannot build the WinUI FFI DLL with -buildmode=c-shared,
REM so this script emits the native ARM64 TUI executable instead.

set "ROOT=%~dp0"
pushd "%ROOT%"

for /f %%i in ('go env GOHOSTARCH') do set "HOST_ARCH=%%i"
if not defined HOST_ARCH (
    echo Failed to detect Go host architecture.
    popd
    exit /b 1
)

if /I "%HOST_ARCH%"=="arm64" goto BUILD_TUI

echo Building TrucoCore FFI DLL...
go env >nul 2>&1
if errorlevel 1 (
    echo Failed to find Go in PATH.
    popd
    exit /b 1
)

dotnet --version >nul 2>&1
if errorlevel 1 (
    echo Failed to find dotnet in PATH.
    popd
    exit /b 1
)

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
go build -buildmode=c-shared -o "%ROOT%bin\truco-core-ffi.dll" .\cmd\truco-core-ffi
if errorlevel 1 (
    echo Failed to build truco-core-ffi.dll.
    popd
    exit /b 1
)

echo Building WinUI GUI...
set "PROJECT=%ROOT%native\windows-winui\TrucoWinUI.csproj"
set "GUI_NAME=truco-gui-winui-windows-amd64-portable"
set "OUTPUT_DIR=%ROOT%bin\gui\winui\%GUI_NAME%"
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"

dotnet publish "%PROJECT%" ^
  -c Release ^
  -r win-x64 ^
  --self-contained true ^
  -p:PublishSingleFile=true ^
  -p:WindowsPackageType=None ^
  -p:WindowsAppSDKSelfContained=true ^
  -p:IncludeNativeLibrariesForSelfExtract=true ^
  -p:EnableCompressionInSingleFile=true ^
  -o "%OUTPUT_DIR%"
if errorlevel 1 (
    echo Failed to build GUI.
    popd
    exit /b 1
)

echo.
echo Build complete!
echo Output: "%OUTPUT_DIR%"
echo Portable bundle ready. Copy "%OUTPUT_DIR%" to another Windows x64 machine.
popd
endlocal
exit /b 0

:BUILD_TUI
echo Windows ARM64 detected.
echo Go does not support -buildmode=c-shared for Windows ARM64 on this toolchain,
echo so the WinUI native client cannot be compiled on this machine.
echo Building the native portable TUI executable instead...

if not exist "%ROOT%bin\tui" mkdir "%ROOT%bin\tui"
set GOOS=windows
set GOARCH=arm64
set "TUI_ARM64_NAME=truco-tui-core-windows-arm64-portable.exe"
go build -trimpath -ldflags="-s -w" -o "%ROOT%bin\tui\%TUI_ARM64_NAME%" .\cmd\truco
if errorlevel 1 (
    echo Failed to build the portable TUI executable.
    popd
    exit /b 1
)

echo.
echo Build complete!
echo Output: "%ROOT%bin\tui\%TUI_ARM64_NAME%"
echo This ARM64 .exe is fully portable.
popd
endlocal
