. .\function\handler.ps1

$stdin_json = [System.Text.StringBuilder]::new()
foreach ($i in $input) {
    [void]$stdin_json.Append($i)
}

$stdin = ConvertFrom-Json -InputObject $stdin_json

Write-Host (handle $stdin.context $stdin.input | ConvertTo-Json)