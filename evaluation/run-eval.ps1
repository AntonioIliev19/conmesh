<#
.SYNOPSIS
Runs one evaluation task headless with Claude Code, with or without ConMesh, and records the result.

.DESCRIPTION
Each run starts from a fresh clone of a local mirror, so runs never touch the original project.
Steps: clone -> apply seeded defect (if any) -> remove existing AI steering (CLAUDE.md, AGENTS.md,
.agents, .claude, .codex) -> for the "conmesh" condition, set up ConMesh (init, index, code-aware pack,
sync) -> run `claude -p` with the task prompt -> record time, turns, tokens and cost -> add the hidden
acceptance tests -> build, run hidden tests and the full test suite -> run grep checks -> append a row
to runs.csv.

.EXAMPLE
.\run-eval.ps1 -Mirror D:\eval\project.git -TaskDir D:\eval\tasks\T1-defect `
  -WorkRoot D:\eval -Condition conmesh -Rep 1 -Solution App.sln -TestProject tests/App.Tests -Model claude-sonnet-5-5
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string] $Mirror,
    [Parameter(Mandatory)] [string] $TaskDir,
    [Parameter(Mandatory)] [string] $WorkRoot,
    [Parameter(Mandatory)] [ValidateSet('bare', 'conmesh', 'conmesh-directed')] [string] $Condition,
    [int] $Rep = 1,
    [string] $Model = 'claude-sonnet-5-5',
    [Parameter(Mandatory)] [string] $TestProject,
    [Parameter(Mandatory)] [string] $Solution,
    [int] $TimeoutMinutes = 45,
    # ConMesh binary used for setup; sync writes its absolute path into .mcp.json.
    [string] $ConmeshExe = 'conmesh'
)

$ErrorActionPreference = 'Stop'
$env:Path = [Environment]::GetEnvironmentVariable('Path', 'Machine') + ';' + [Environment]::GetEnvironmentVariable('Path', 'User')

function Invoke-Native {
    param([string] $Exe, [string[]] $Arguments, [string] $LogFile)
    # Windows PowerShell 5.1 turns native stderr into errors; only the exit code decides failure here.
    $ErrorActionPreference = 'Continue'
    $out =& $Exe @Arguments 2>&1 | ForEach-Object { "$_" }
    if ($LogFile) { $out | Out-File -FilePath $LogFile -Encoding utf8 -Append }
    return [pscustomobject]@{ ExitCode = $LASTEXITCODE; Output = $out }
}

function Save-GitCommit {
    param([string] $Message)
    Invoke-Native git @('add', '-A') | Out-Null
    Invoke-Native git @('-c', 'user.name=eval', '-c', 'user.email=eval@localhost', 'commit', '--quiet', '--no-verify', '--allow-empty', '-m', $Message) | Out-Null
}

$task = Get-Content (Join-Path $TaskDir 'task.json') -Raw | ConvertFrom-Json
$prompt = Get-Content (Join-Path $TaskDir 'prompt.md') -Raw
$runId = '{0}-{1}-r{2}-{3}' -f $task.id, $Condition, $Rep, (Get-Date -Format 'yyyyMMdd-HHmmss')
$workDir = Join-Path $WorkRoot "runs\$runId"
$resultDir = Join-Path $WorkRoot "results\$runId"
New-Item -ItemType Directory -Force $resultDir | Out-Null
$log = Join-Path $resultDir 'run.log'

Write-Host "== $runId"
Invoke-Native git @('clone', '--quiet', $Mirror, $workDir) $log | Out-Null
Push-Location $workDir
try {
    $baseCommit = (Invoke-Native git @('rev-parse', 'HEAD')).Output | Select-Object -First 1

    # 1. Seeded defect, if the task has one.
    foreach ($seed in @($task.seeds)) {
        if (-not $seed) { continue }
        $path = Join-Path $workDir $seed.file
        $text = [IO.File]::ReadAllText($path)
        if (-not $text.Contains($seed.find)) { throw "Seed text not found in $($seed.file)" }
        [IO.File]::WriteAllText($path, $text.Replace($seed.find, $seed.replace))
    }

    # 2. Same starting point for both conditions: no hand-written AI steering.
    foreach ($p in 'CLAUDE.md', 'AGENTS.md', 'src/CLAUDE.md', '.agents', '.claude', '.codex', '.aikit', '.mcp.json') {
        if (Test-Path $p) { Remove-Item -Recurse -Force $p }
    }
    Save-GitCommit 'eval: baseline'

    # 3. ConMesh setup (timed separately: it is a one-off cost per project).
    $setupSeconds = 0
    $mcpConfig = Join-Path $resultDir 'mcp-empty.json'
    '{"mcpServers":{}}' | Out-File -FilePath $mcpConfig -Encoding ascii
    if ($Condition -like 'conmesh*') {
        $sw = [Diagnostics.Stopwatch]::StartNew()
        foreach ($step in @(
                @('--no-input', 'init', '--harness', 'claude', '--index'),
                @('--no-input', 'index'),
                @('--no-input', 'pack', 'add', 'conmesh-code-aware', '--trust', '--project'),
                @('--no-input', 'sync', '--trust'))) {
            $r = Invoke-Native $ConmeshExe $step $log
            if ($r.ExitCode -ne 0) { throw "conmesh $($step -join ' ') failed; see $log" }
        }
        $setupSeconds = [math]::Round($sw.Elapsed.TotalSeconds, 1)
        if (Test-Path '.mcp.json') { $mcpConfig = Join-Path $workDir '.mcp.json' }
        Copy-Item $mcpConfig (Join-Path $resultDir 'mcp-used.json')
    }
    if ($Condition -eq 'conmesh-directed') {
        # Same ConMesh setup, plus a project instruction that makes the index the first step of navigation.
        $directive = @'

## Code navigation (required)

Search the ConMesh index before you use Grep, Glob or read files to locate code:

- Use `conmesh_search_codebase` (ConMesh MCP server) for behaviour or concepts described in words.
- Use `conmesh_find_symbol` for named types, methods and members.
- Use Grep only to confirm exact text, or to find every occurrence of a literal pattern you already know.
'@
        Add-Content -Path 'CLAUDE.md' -Value $directive -Encoding UTF8
        Copy-Item 'CLAUDE.md' (Join-Path $resultDir 'CLAUDE.md')
    }
    Save-GitCommit 'eval: condition setup'
    $setupCommit = (Invoke-Native git @('rev-parse', 'HEAD')).Output | Select-Object -First 1

    # 4. The agent run.
    $allowed = @('Read', 'Edit', 'Write', 'Glob', 'Grep', 'TodoWrite', 'Task',
        'Bash(dotnet build:*)', 'Bash(dotnet test:*)', 'Bash(git diff:*)', 'Bash(git status:*)',
        'PowerShell(dotnet build:*)', 'PowerShell(dotnet test:*)', 'PowerShell(git diff:*)', 'PowerShell(git status:*)')
    if ($Condition -like 'conmesh*') { $allowed += @('mcp__conmesh', 'Bash(conmesh:*)', 'PowerShell(conmesh:*)') }
    # The prompt goes through stdin: Windows PowerShell 5.1 splits native arguments at embedded quotes.
    $claudeArgs = @('-p', '--output-format', 'stream-json', '--verbose', '--model', $Model,
        '--permission-mode', 'acceptEdits', '--setting-sources', 'project,local',
        '--strict-mcp-config', '--mcp-config', $mcpConfig, '--no-session-persistence',
        '--allowedTools') + $allowed
    $sw = [Diagnostics.Stopwatch]::StartNew()
    $job = Start-Job -ScriptBlock {
        param($dir, $argList, $promptText)
        Set-Location $dir
        $ErrorActionPreference = 'Continue'
        $OutputEncoding = New-Object System.Text.UTF8Encoding $false
        $env:Path = [Environment]::GetEnvironmentVariable('Path', 'Machine') + ';' + [Environment]::GetEnvironmentVariable('Path', 'User')
        $promptText | & claude @argList 2>&1 | ForEach-Object { "$_" }
    } -ArgumentList $workDir, $claudeArgs, $prompt
    $timedOut = -not (Wait-Job $job -Timeout ($TimeoutMinutes * 60))
    if ($timedOut) { Stop-Job $job }
    $agentOut = (Receive-Job $job) -join "`n"
    Remove-Job $job -Force
    $wallSeconds = [math]::Round($sw.Elapsed.TotalSeconds, 1)
    $agentOut | Out-File (Join-Path $resultDir 'transcript.jsonl') -Encoding utf8

    # The last "result" event holds totals; every assistant "tool_use" block is one tool call.
    $cj = $null
    $toolCounts = @{}
    foreach ($line in ($agentOut -split "`n")) {
        $t = $line.Trim()
        if (-not $t.StartsWith('{')) { continue }
        try { $ev = $t | ConvertFrom-Json } catch { continue }
        if ($ev.type -eq 'result') { $cj = $ev }
        if ($ev.type -eq 'assistant') {
            foreach ($block in @($ev.message.content)) {
                if ($block.type -eq 'tool_use') { $toolCounts[$block.name] = 1 + [int]$toolCounts[$block.name] }
            }
        }
    }
    $toolSummary = ($toolCounts.GetEnumerator() | Sort-Object Name | ForEach-Object { "$($_.Name)=$($_.Value)" }) -join '; '
    $conmeshCalls = ($toolCounts.GetEnumerator() | Where-Object { $_.Name -like 'mcp__conmesh*' -or $_.Name -like '*conmesh*' } | Measure-Object -Property Value -Sum).Sum
    $denials = if ($cj) { @($cj.permission_denials).Count } else { '' }

    # 5. What the agent changed (before hidden tests are added).
    Invoke-Native git @('add', '-A') | Out-Null
    $numstat = (Invoke-Native git @('diff', '--cached', '--numstat', $setupCommit)).Output
    $files = 0; $added = 0; $deleted = 0
    foreach ($line in $numstat) {
        $parts = "$line" -split "`t"
        if ($parts.Count -ge 3) { $files++; if ($parts[0] -ne '-') { $added += [int]$parts[0]; $deleted += [int]$parts[1] } }
    }
    (Invoke-Native git @('diff', '--cached', $setupCommit)).Output | Out-File (Join-Path $resultDir 'agent.diff') -Encoding utf8

    # 6. Grep checks (e.g. refactor completeness).
    $grepResults = @()
    $grepFail = $false
    foreach ($g in @($task.grepChecks)) {
        if (-not $g) { continue }
        $count = @(Get-ChildItem -Path $g.path -Recurse -File -Include $g.include |
            Where-Object { $_.FullName -notmatch '\\(bin|obj)\\' } |
            Select-String -Pattern $g.pattern).Count
        $grepResults += '{0}={1}(max {2})' -f $g.name, $count, $g.maxAllowed
        if ($count -gt $g.maxAllowed) { $grepFail = $true }
    }

    # 7. Hidden acceptance tests, then the full suite.
    foreach ($h in @($task.hiddenTests)) {
        Copy-Item (Join-Path $TaskDir $h) (Join-Path $workDir $TestProject)
    }
    $hidden = Invoke-Native dotnet @('test', $Solution, '--nologo', '--filter', $task.testFilter) (Join-Path $resultDir 'hidden-tests.log')
    $full = Invoke-Native dotnet @('test', $Solution, '--nologo', '--no-build') (Join-Path $resultDir 'full-tests.log')
    $fullSummary = ($full.Output | Where-Object { $_ -match '(Passed|Failed)!\s+-' }) -join ' | '

    $row = [pscustomobject]@{
        run_id            = $runId
        task              = $task.id
        task_type         = $task.type
        condition         = $Condition
        rep               = $Rep
        model             = $Model
        claude_version    = ((Invoke-Native claude @('--version')).Output | Select-Object -First 1)
        conmesh_version   = $(if ($Condition -like 'conmesh*') { (Invoke-Native $ConmeshExe @('--version')).Output | Select-Object -First 1 } else { '' })
        base_commit       = $baseCommit
        setup_seconds     = $setupSeconds
        wall_seconds      = $wallSeconds
        timed_out         = $timedOut
        agent_error       = $(if ($cj) { $cj.is_error } else { 'unparsed' })
        num_turns         = $(if ($cj) { $cj.num_turns } else { '' })
        cost_usd          = $(if ($cj) { $cj.total_cost_usd } else { '' })
        input_tokens      = $(if ($cj) { $cj.usage.input_tokens } else { '' })
        cache_read_tokens = $(if ($cj) { $cj.usage.cache_read_input_tokens } else { '' })
        cache_write_tokens = $(if ($cj) { $cj.usage.cache_creation_input_tokens } else { '' })
        output_tokens     = $(if ($cj) { $cj.usage.output_tokens } else { '' })
        tool_calls        = $toolSummary
        conmesh_calls     = [int]$conmeshCalls
        permission_denials = $denials
        files_changed     = $files
        lines_added       = $added
        lines_deleted     = $deleted
        hidden_tests_pass = ($hidden.ExitCode -eq 0)
        full_suite_pass   = ($full.ExitCode -eq 0)
        full_suite        = $fullSummary
        grep_checks       = ($grepResults -join '; ')
        grep_pass         = (-not $grepFail)
        accepted          = (($hidden.ExitCode -eq 0) -and ($full.ExitCode -eq 0) -and (-not $grepFail) -and (-not $timedOut))
    }
    $csv = Join-Path $WorkRoot 'results\runs.csv'
    $row | Export-Csv -Path $csv -NoTypeInformation -Append -Encoding UTF8
    $row | Format-List | Out-String | Write-Host
}
finally {
    Pop-Location
}
