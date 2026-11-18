# 一键启动三个后端服务（core / ai / bff）
# 用法：pwsh -File .\scripts\dev.ps1
# 停止：pwsh -File .\scripts\stop-dev.ps1
$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
New-Item -ItemType Directory -Force -Path logs | Out-Null

Write-Host 'building ...'
& go build -o .\bin\core.exe ./app/core/cmd/server
& go build -o .\bin\ai.exe ./app/ai/cmd/server
& go build -o .\bin\bff.exe ./app/bff/cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Error 'build failed'
}

$wd = (Get-Location).Path

$services = @(
    @{ Name = 'core'; Exe = '.\bin\core.exe'; Conf = 'configs/core.yaml' },
    @{ Name = 'ai';   Exe = '.\bin\ai.exe';   Conf = 'configs/ai.yaml' },
    @{ Name = 'bff';  Exe = '.\bin\bff.exe';  Conf = 'configs/bff.yaml' }
)

foreach ($s in $services) {
    $p = Start-Process -FilePath $s.Exe `
        -ArgumentList '-conf', $s.Conf `
        -WorkingDirectory $wd `
        -NoNewWindow `
        -PassThru `
        -RedirectStandardOutput "logs\$($s.Name).log" `
        -RedirectStandardError "logs\$($s.Name).err.log"
    Write-Host ("started {0,-5} pid={1}  log=logs\{0}.log" -f $s.Name, $p.Id)
}

Start-Sleep -Seconds 3

Write-Host ''
Write-Host 'ports:'
Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue |
    Where-Object { $_.LocalPort -in 8000, 8001, 8002, 9000, 9001 } |
    Select-Object LocalPort, OwningProcess |
    Sort-Object LocalPort |
    Format-Table -AutoSize

Write-Host 'readyz:'
try {
    (Invoke-WebRequest 'http://127.0.0.1:8000/readyz' -SkipHttpErrorCheck -TimeoutSec 5).Content
} catch {
    Write-Host ('core unreachable: ' + $_.Exception.Message)
}

Write-Host ''
Write-Host '前端：cd web; npm.cmd run dev'
