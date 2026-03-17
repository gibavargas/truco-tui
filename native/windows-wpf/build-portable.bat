@echo off
setlocal

set "ROOT=%~dp0\..\..\"
pushd "%ROOT%"

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

echo Building truco-core-ffi.dll...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
go build -buildmode=c-shared -o "%ROOT%bin\truco-core-ffi.dll" .\cmd\truco-core-ffi
if errorlevel 1 (
    echo Failed to build truco-core-ffi.dll.
    popd
    exit /b 1
)

set "PROJECT=%ROOT%native\windows-wpf\TrucoWPF.csproj"
set "OUTPUT_DIR=%ROOT%bin\gui\wpf\truco-gui-wpf-windows-amd64-portable"

if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"

echo Publishing WPF portable bundle...
dotnet publish "%PROJECT%" ^
  -c Release ^
  -r win-x64 ^
  --self-contained true ^
  -p:PublishSingleFile=true ^
  -p:IncludeNativeLibrariesForSelfExtract=true ^
  -p:EnableCompressionInSingleFile=true ^
  -o "%OUTPUT_DIR%"
if errorlevel 1 (
    echo Failed to publish the WPF client.
    popd
    exit /b 1
)

echo.
echo Build complete.
echo Output: "%OUTPUT_DIR%"
echo Portable runtime library layout:
echo   "%OUTPUT_DIR%\lib\truco-core-ffi.dll"

popd
endlocal
exit /b 0
