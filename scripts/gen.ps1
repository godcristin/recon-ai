# 生成 proto 代码（Windows PowerShell 版）
# 用法：pwsh -File .\scripts\gen.ps1
$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

# well-known types 已随仓库放在 third_party/google/protobuf 下，仓库自包含，不依赖本机 protoc 安装路径
$files = @()
foreach ($dir in @('api', 'app')) {
    if (Test-Path $dir) {
        $files += Get-ChildItem -Path $dir -Recurse -Filter *.proto |
            ForEach-Object { $_.FullName.Substring($root.Length + 1).Replace('\', '/') }
    }
}

if (-not $files) {
    Write-Error 'no proto files found under api/ or app/'
}

Write-Host "generating $($files.Count) proto files ..."

& protoc -I . -I ./third_party `
    --go_out=paths=source_relative:. `
    --go-grpc_out=paths=source_relative:. `
    --go-http_out=paths=source_relative:. `
    --go-errors_out=paths=source_relative:. `
    --validate_out=lang=go,paths=source_relative:. `
    $files

if ($LASTEXITCODE -ne 0) {
    Write-Error "protoc failed with exit code $LASTEXITCODE"
}

Write-Host 'done.'
