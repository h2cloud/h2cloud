@go install mainpkg
@if %errorlevel% EQU 0 (
    ECHO Running...
    mainpkg.exe
) else (
    ECHO CompileError
)
