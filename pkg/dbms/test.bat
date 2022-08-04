@echo off

for /f "usebackq tokens=*" %%a in (`where go`) do set gobin=%%a

set pageTests = "test page"
set diskTests = "test disk"
set bufferTests = "test buffer"

if %1%==pageTests call :runPageTests
if %1%==diskTests call :runDiskTests
if %1%==bufferTests call :runBufferTests

:runPageTests
"%gobin%" test dbms/page -run "^TestPage.*$"
exit /b 0

:runDiskTests
"%gobin%" test dbms/disk -run "^TestDisk.*$"
exit /b 0

:runBufferTests
"%gobin%" test dbms/buffer -run "^TestBuffer.*$"
exit /b 0
