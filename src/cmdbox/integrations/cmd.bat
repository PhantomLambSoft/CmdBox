@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "_tmp=%TEMP%\cmdbox_emit_%RANDOM%%RANDOM%.cmd"

cmdbox %* --emit > "%_tmp%"
if errorlevel 1 (
  del "%_tmp%" >nul 2>&1
  exit /b %errorlevel%
)

call "%_tmp%"
del "%_tmp%" >nul 2>&1
exit /b 0
