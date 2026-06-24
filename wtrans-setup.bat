@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "BIN=%~dp0wtrans.exe"
if not exist "%BIN%" (
  echo wtrans.exe was not found next to this script.
  exit /b 1
)

:menu
echo.
echo wtrans quick setup
echo 1. List visible windows
echo 2. Make an app transparent
echo 3. Turn transparent mode off for an app
echo 4. Show status
echo 5. Stop background keeper
echo 6. Reset all saved rules
echo 7. Exit
set /p CHOICE=Choose 1-7: 

if "%CHOICE%"=="1" goto list
if "%CHOICE%"=="2" goto set
if "%CHOICE%"=="3" goto restore
if "%CHOICE%"=="4" goto status
if "%CHOICE%"=="5" goto stop
if "%CHOICE%"=="6" goto reset
if "%CHOICE%"=="7" goto end
goto menu

:list
set /p PROCESS=Process name (blank for all): 
"%BIN%" list --process "!PROCESS!"
goto menu

:set
set /p PROCESS=Process name (example: WindowsTerminal.exe): 
set /p OPACITY=Opacity 20-100 (example: 85): 
set /p PERSIST=Keep future windows too? (y/n): 
if /I "!PERSIST!"=="y" (
  "%BIN%" set --process "!PROCESS!" --opacity !OPACITY! --persist
) else (
  "%BIN%" set --process "!PROCESS!" --opacity !OPACITY!
)
goto menu

:restore
set /p PROCESS=Process name: 
"%BIN%" restore --process "!PROCESS!"
goto menu

:status
"%BIN%" status
goto menu

:stop
"%BIN%" stop
goto menu

:reset
echo This will clear all saved rules.
set /p CONFIRM=Type RESET to continue: 
if /I not "!CONFIRM!"=="RESET" goto menu
"%BIN%" reset
goto menu

:end
endlocal
