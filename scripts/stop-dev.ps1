# 停止三个后端服务
# 用法：pwsh -File .\scripts\stop-dev.ps1
$ports = @(8000, 8001, 8002, 9000, 9001)

$pids = Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue |
    Where-Object { $_.LocalPort -in $ports } |
    Select-Object -ExpandProperty OwningProcess -Unique

if (-not $pids) {
    Write-Host 'no service is listening on 8000/8001/8002/9000/9001'
    exit 0
}

foreach ($procId in $pids) {
    try {
        Stop-Process -Id $procId -Force -ErrorAction Stop
        Write-Host "stopped pid=$procId"
    } catch {
        Write-Host "stop pid=$procId failed: $($_.Exception.Message)"
    }
}
