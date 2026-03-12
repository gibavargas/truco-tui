$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location $root
try {
    $hostArch = (go env GOHOSTARCH).Trim()
    if ($hostArch -eq "arm64") {
        Write-Host "Windows ARM64 detected. Building the portable TUI executable because Go does not support WinUI FFI DLL generation on Windows ARM64."
        $tuiDir = Join-Path $root "bin\tui"
        if (-not (Test-Path $tuiDir)) {
            New-Item -ItemType Directory -Force -Path $tuiDir | Out-Null
        }
        $tuiName = "truco-tui-core-windows-arm64-portable.exe"
        $tuiOutput = Join-Path $tuiDir $tuiName
        $env:GOOS = "windows"
        $env:GOARCH = "arm64"
        go build -trimpath -ldflags="-s -w" -o $tuiOutput .\cmd\truco
        Write-Host "Build complete. Portable executable: $tuiOutput"
        return
    }

    $projectPath = Join-Path $root "native\windows-winui\TrucoWinUI.csproj"
    $dllPath = Join-Path $root "bin\truco-core-ffi.dll"
    $outputDir = Join-Path $root "bin\gui\winui\truco-gui-winui-windows-amd64-portable"

    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "1"
    go build -buildmode=c-shared -o $dllPath .\cmd\truco-core-ffi

    if (Test-Path $outputDir) {
        Remove-Item $outputDir -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path (Split-Path $outputDir) | Out-Null

    dotnet publish $projectPath `
        -c Release `
        -r win-x64 `
        --self-contained true `
        -p:PublishSingleFile=true `
        -p:WindowsPackageType=None `
        -p:WindowsAppSDKSelfContained=true `
        -p:IncludeNativeLibrariesForSelfExtract=true `
        -p:EnableCompressionInSingleFile=true `
        -o $outputDir

    Write-Host "Build complete. Portable bundle: $outputDir"
}
finally {
    Pop-Location
}
