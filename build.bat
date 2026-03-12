@echo off
setlocal
dotnet build "%~dp0native\windows-winui\TrucoWinUI.csproj" -c Release
endlocal
