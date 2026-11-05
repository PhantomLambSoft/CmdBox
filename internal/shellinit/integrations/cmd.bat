@echo off
setlocal EnableExtensions

set "_tmp=%TEMP%\cmdbox_emit_%RANDOM%%RANDOM%.cmd"
set "CMDBOX_EMIT_FILE=%_tmp%"

cb %* --emit
set "_rc=%errorlevel%"

if not "%_rc%"=="0" (
  if exist "%_tmp%" del "%_tmp%" >nul 2>&1
  exit /b %_rc%
)
if not exist "%_tmp%" exit /b 0

rem endlocal first so cd and set in the emitted commands persist in the caller.
rem %_tmp% is expanded when this line is parsed, before endlocal runs.
endlocal & call "%_tmp%" & del "%_tmp%" >nul 2>&1
