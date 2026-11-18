$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
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
