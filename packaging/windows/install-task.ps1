# Registra move-mouse como tarea programada que inicia al iniciar sesion.
# Uso (PowerShell): .\install-task.ps1 -Exe "C:\Tools\move-mouse.exe"
param([Parameter(Mandatory)][string]$Exe)
$action  = New-ScheduledTaskAction -Execute $Exe -Argument "--stop-key none"
$trigger = New-ScheduledTaskTrigger -AtLogOn
Register-ScheduledTask -TaskName "move-mouse" -Action $action -Trigger $trigger -Force
