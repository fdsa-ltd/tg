@echo off
setlocal

FOR /F "delims=" %%i IN ("%cd%") DO (
    set name=%%~ni
) 

if "%1" == "" (
    del go.mod
    del go.sum
    go mod init fdsa.ltd/%name%
)

REM gofmt -w src
set GOOS=linux
set GOARCH=amd64

if "%1" == "" (
    go build -o ./bin/%name% fdsa.ltd/%name%/src
) else (
    go build -ldflags="-s -w" -o ./bin/%name% fdsa.ltd/%name%/src
)
echo go: linux version is finished ok

set GOOS=windows
set GOARCH=amd64
if "%1" == "" (
    go build -o ./bin/%name%.exe fdsa.ltd/%name%/src
) else (
    go build -ldflags="-s -w" -o ./bin/%name%.exe fdsa.ltd/%name%/src
)
echo go: windows version is finished ok


@REM if "%1" == "" (
@REM     echo go: package...
@REM ) else (
@REM     upx -9 ./bin/%name%.exe
@REM     upx -9 ./bin/%name%
@REM )

@REM if not exist target (
@REM     mkdir target
@REM )
@REM cd ..

@REM tar -czvf %name%/target/%name%.tar.gz %name%/install %name%/install.bat %name%/bin %name%/*.md %name%/docs
@REM zip %name%/target/%name%.zip %name%/install %name%/install.bat %name%/bin %name%/*.md %name%/docs
@REM echo go: all is finished
