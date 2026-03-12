$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectPath = Join-Path $root "native\windows-winui\TrucoWinUI.csproj"

dotnet build $projectPath -c Release
