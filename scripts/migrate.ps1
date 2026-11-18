param(
    [string]$User = "recon",
    [string]$Password = "recon123",
    [string]$Database = "recon",
    [string]$DbHost = "127.0.0.1",
    [int]$Port = 3306
)
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
$mysql = Get-Command mysql -ErrorAction SilentlyContinue
if (-not $mysql) {
    Write-Error "mysql client not found in PATH"
}
$files = Get-ChildItem -Path 'db\migrations' -Filter *.sql | Sort-Object Name
if (-not $files) {
    Write-Error 'no migration files found'
}
foreach ($f in $files) {
    Write-Host "applying $($f.Name) ..."
    $path = $f.FullName.Replace('\', '/')
    & $mysql.Source -h $DbHost -P $Port -u $User "-p$Password" $Database -e "source $path"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "migration failed: $($f.Name)"
    }
}
Write-Host 'all migrations applied.'
