$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exePath   = Join-Path $scriptDir "tempod.exe"
$taskName  = "tempod"

if (-not (Test-Path $exePath)) {
    Write-Error "tempod.exe not found in $scriptDir"
    exit 1
}

$action   = New-ScheduledTaskAction -Execute $exePath
$trigger  = New-ScheduledTaskTrigger -AtLogon
$settings = New-ScheduledTaskSettingsSet `
    -ExecutionTimeLimit (New-TimeSpan -Hours 0) `
    -RestartCount 10 `
    -RestartInterval (New-TimeSpan -Minutes 1) `
    -StartWhenAvailable

Register-ScheduledTask `
    -TaskName $taskName `
    -Action $action `
    -Trigger $trigger `
    -Settings $settings `
    -RunLevel Highest `
    -Force | Out-Null

Start-ScheduledTask -TaskName $taskName

Write-Host "tempod registered as a scheduled task and started."
Write-Host "It will run automatically at every login."
