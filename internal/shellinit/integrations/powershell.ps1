function cbe
{
    $emitFile = Join-Path ([System.IO.Path]::GetTempPath()) ("cmdbox_emit_{0}.tmp" -f [guid]::NewGuid().ToString("N"))
    $previous = $env:CMDBOX_EMIT_FILE
    $env:CMDBOX_EMIT_FILE = $emitFile
    try
    {
        & cb @args --emit
        $code = $LASTEXITCODE
    }
    finally
    {
        $env:CMDBOX_EMIT_FILE = $previous
    }
    try
    {
        if ($code -eq 0 -and (Test-Path -LiteralPath $emitFile))
        {
            $script = [System.IO.File]::ReadAllText($emitFile)
            if (-not [string]::IsNullOrWhiteSpace($script))
            {
                Invoke-Expression $script
            }
        }
    }
    finally
    {
        Remove-Item -LiteralPath $emitFile -Force -ErrorAction SilentlyContinue
    }
}
