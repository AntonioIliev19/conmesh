<#
.SYNOPSIS
Summarizes runs.csv by task and condition (bare vs conmesh): acceptance rate, median time, turns, cost, tokens.
.EXAMPLE
.\summarize.ps1 -Csv D:\eval\results\runs.csv
#>
param([Parameter(Mandatory)] [string] $Csv)

function Get-Median([double[]] $values) {
    if (-not $values -or $values.Count -eq 0) { return $null }
    $s = $values | Sort-Object
    $n = $s.Count
    if ($n % 2) { return $s[($n - 1) / 2] }
    return ($s[$n / 2 - 1] + $s[$n / 2]) / 2
}

$rows = Import-Csv $Csv
$rows | Group-Object task, condition | ForEach-Object {
    $g = $_.Group
    [pscustomobject]@{
        task            = $g[0].task
        condition       = $g[0].condition
        runs            = $g.Count
        accepted        = '{0}/{1}' -f @($g | Where-Object { $_.accepted -eq 'True' }).Count, $g.Count
        median_minutes  = [math]::Round((Get-Median ($g | ForEach-Object { [double]$_.wall_seconds })) / 60, 1)
        median_turns    = Get-Median ($g | Where-Object num_turns | ForEach-Object { [double]$_.num_turns })
        median_cost_usd = [math]::Round((Get-Median ($g | Where-Object cost_usd | ForEach-Object { [double]$_.cost_usd })), 3)
        median_out_tok  = Get-Median ($g | Where-Object output_tokens | ForEach-Object { [double]$_.output_tokens })
        median_lines    = Get-Median ($g | ForEach-Object { [double]$_.lines_added + [double]$_.lines_deleted })
    }
} | Sort-Object task, condition | Format-Table -AutoSize
