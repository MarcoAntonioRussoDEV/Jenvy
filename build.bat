@echo off
SETLOCAL ENABLEDELAYEDEXPANSION

:: ⚙️ Step 1: Build Go binary
echo 🔧 Building jvm.exe...
go build -o distribution\jvm.exe main.go
IF NOT EXIST distribution\jvm.exe (
    echo ❌ Build fallito: jvm.exe non trovato
    goto end
)

:: 🔐 Step 2: Sign executable with certificate
SET SIGNTOOL="C:\Program Files (x86)\Windows Kits\10\bin\x64\signtool.exe"
IF EXIST %SIGNTOOL% (
    echo 🔐 Firmo jvm.exe con certificato autofirmato...
    %SIGNTOOL% sign /f distribution\jvm-dev-cert.pfx /p jvm-password /tr http://timestamp.digicert.com /td sha256 distribution\jvm.exe
) ELSE (
    echo ⚠️ SignTool non trovato. Salta firma. Assicurati di avere Windows SDK installato.
)

:: 📦 Step 3: Compile installer with Inno Setup
SET INNO="C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
IF EXIST %INNO% (
    echo 📦 Compilo installer Inno Setup...
    %INNO% distribution\setup.iss
) ELSE (
    echo ⚠️ ISCC.exe non trovato. Installa Inno Setup Compiler da https://jrsoftware.org
)

echo ✅ Build completato! Controlla la cartella distribution\

:end
pause
ENDLOCAL
