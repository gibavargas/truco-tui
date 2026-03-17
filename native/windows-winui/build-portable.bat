@echo off
setlocal

set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..\..") do set "REPO_ROOT=%%~fI"
set "OUTPUT_DIR=%REPO_ROOT%\bin\gui\winui\truco-gui-winui-windows-amd64-portable"
set "CORE_DLL=%REPO_ROOT%\bin\truco-core-ffi.dll"

if not exist "%CORE_DLL%" (
  echo Missing "%CORE_DLL%".
  echo Build the shared runtime first with: make ffi-windows
  exit /b 1
)

echo Publishing WinUI portable bundle to "%OUTPUT_DIR%"
dotnet publish "%SCRIPT_DIR%TrucoWinUI.csproj" ^
  -c Release ^
  -r win-x64 ^
  --self-contained true ^
  -p:PublishSingleFile=false ^
  -p:Platform=x64 ^
  -o "%OUTPUT_DIR%"

if errorlevel 1 exit /b 1

copy /Y "%CORE_DLL%" "%OUTPUT_DIR%\truco-core-ffi.dll" >nul
if errorlevel 1 exit /b 1

echo Portable WinUI bundle ready at "%OUTPUT_DIR%"
