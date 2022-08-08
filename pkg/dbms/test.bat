@echo off

for /f "usebackq tokens=*" %%a in (`where go`) do set gobin=%%a

:: check for usage
if [%1]==[] call :usage

::IF %~1 == "test" and %~2 == "page" (
if %~1=="test page" (
  CALL :page_tests
)
IF %~1 == "test" and %~2 == "disk" (
  CALL :disk_tests
)
IF %~1 == "test" and %~2 == "buffer" (
  CALL :buffer_tests
)
IF %~1 == "test" and %~2 == "all" (
  CALL :all_tests
)

EXIT /B %ERRORLEVEL%

:usage
echo "usage function"
exit /B 0

:page_tests
echo ":: running page tests..."
cd page
%gobin% test -v -run '^TestPage.*$'
cd ..
echo ":: page tests are done"
EXIT /B 0

:disk_tests
echo ":: running disk tests..."
cd disk
%gobin% test -v -run '^TestDisk.*$'
%gobin% test -v -run '^TestSegment.*$'
cd ..
echo ":: disk tests are done"
EXIT /B 0

:buffer_tests
echo ":: running buffer tests..."
cd buffer
%gobin% test -v -run "^TestBuffer.*$"
%gobin% test -v -run "^TestClockReplacer.*$"
cd ..
echo ":: buffer tests are done"
EXIT /B 0

:all_tests
CALL :page_tests
CALL :disk_tests
CALL :buffer_tests
EXIT /B 0
