Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Assert-Command {
	param(
		[Parameter(Mandatory = $true)]
		[string]$Name
	)

	if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
		throw "未找到命令: $Name，请先安装并确保它在 PATH 中。"
	}
}

function Start-DevJob {
	param(
		[Parameter(Mandatory = $true)]
		[string]$Name,
		[Parameter(Mandatory = $true)]
		[string]$WorkingDirectory,
		[Parameter(Mandatory = $true)]
		[string]$CommandPath,
		[Parameter(Mandatory = $true)]
		[string[]]$Arguments
	)

	return Start-Job -Name $Name -ScriptBlock {
		param($wd, $exe, $argsForExe)

		Set-Location $wd
		Write-Output "cwd=$wd"
		Write-Output "exec=$exe $($argsForExe -join ' ')"

		& $exe @argsForExe 2>&1

		$exitCode = if ($null -ne $LASTEXITCODE) { $LASTEXITCODE } else { 0 }
		Write-Output "exit_code=$exitCode"
	} -ArgumentList $WorkingDirectory, $CommandPath, $Arguments
}

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$frontendDir = Join-Path $root 'frontend'
$backendDir = Join-Path $root 'backend'

if (-not (Test-Path $frontendDir)) {
	throw "前端目录不存在: $frontendDir"
}

if (-not (Test-Path $backendDir)) {
	throw "后端目录不存在: $backendDir"
}

Assert-Command -Name 'npm'
Assert-Command -Name 'go'

$npmPath = (Get-Command npm).Source
$goPath = (Get-Command go).Source

$terminalStates = @('Completed', 'Failed', 'Stopped')
$jobs = @()
$stateAnnounced = @{}

Write-Host '正在启动前后端服务...' -ForegroundColor Cyan
Write-Host "frontend: $frontendDir" -ForegroundColor DarkCyan
Write-Host "backend : $backendDir" -ForegroundColor DarkCyan
Write-Host '按 Ctrl+C 可停止全部服务。' -ForegroundColor DarkYellow

try {
	$jobs += Start-DevJob -Name 'frontend' -WorkingDirectory $frontendDir -CommandPath $npmPath -Arguments @('run', 'dev')
	$jobs += Start-DevJob -Name 'backend' -WorkingDirectory $backendDir -CommandPath $goPath -Arguments @('run', '.\cmd\server\.')

	while ($true) {
		$allTerminated = $true

		foreach ($job in $jobs) {
			$liveJob = Get-Job -Id $job.Id -ErrorAction SilentlyContinue
			if (-not $liveJob) {
				continue
			}

			$prefix = "[$($job.Name)]"
			$color = if ($job.Name -eq 'frontend') { 'Cyan' } else { 'Green' }

			$messages = Receive-Job -Job $liveJob -ErrorAction SilentlyContinue
			foreach ($m in $messages) {
				Write-Host "$prefix $m" -ForegroundColor $color
			}

			if ($liveJob.State -notin $terminalStates) {
				$allTerminated = $false
			} elseif (-not $stateAnnounced.ContainsKey($job.Id)) {
				$stateAnnounced[$job.Id] = $true
				Write-Host "$prefix 已退出，状态: $($liveJob.State)" -ForegroundColor Yellow
			}
		}

		if ($allTerminated) {
			break
		}

		Start-Sleep -Milliseconds 220
	}
}
finally {
	Write-Host '正在清理后台任务...' -ForegroundColor DarkYellow
	foreach ($job in $jobs) {
		if ($job.State -notin $terminalStates) {
			Stop-Job -Job $job -ErrorAction SilentlyContinue
		}
		Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
	}
}

