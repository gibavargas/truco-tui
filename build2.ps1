$ErrorActionPreference = "Stop"

$projectPath = Join-Path $PSScriptRoot "native\windows-winui\TrucoWinUI.csproj"

dotnet build $projectPath -c Release
