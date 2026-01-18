function cb
{
    $out = & cmdbox @args --emit
    $code = $LASTEXITCODE
    if ($code -ne 0)
    {
        return
    }
    if ( [string]::IsNullOrWhiteSpace($out))
    {
        return
    }
    Invoke-Expression $out
}
