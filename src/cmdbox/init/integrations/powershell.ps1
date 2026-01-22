function cbe
{
    $out = & cb @args --emit
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
