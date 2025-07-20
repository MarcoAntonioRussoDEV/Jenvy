::  build.bat - Script di build automatizzato per J:: Step 2.5: Convert README.md to distribution/README.txt
echo ► Converting README.md to distribution/README.txt...
powershell -Command "
$md = Get-Content 'README.md' -Raw -Encoding UTF8;
# Remove horizontal lines (---)
$txt = $md -replace '^-{3,}$', '';
# Remove headers (### Title -> Title)
$txt = $txt -replace '^#{1,6}\s*(.*)$', '$1';
# Remove bold (**text** -> text)
$txt = $txt -replace '\*\*([^\*]+)\*\*', '$1';
# Remove italic (*text* -> text)  
$txt = $txt -replace '\*([^\*]+)\*', '$1';
# Remove inline code (`code` -> code)
$txt = $txt -replace '`([^`]+)`', '$1';
# Convert links [text](url) to just text
$txt = $txt -replace '\[([^\]]+)\]\([^\)]+\)', '$1';
# Convert bullet points
$txt = $txt -replace '^[\s]*-\s*', '* ';
# Remove code blocks
$txt = $txt -replace '(?s)```.*?```', '';
$txt = $txt -replace '(?s)````.*?````', '';
# Clean up multiple newlines
$txt = $txt -replace '\r?\n\r?\n\r?\n+', [char]13+[char]10+[char]13+[char]10;
$txt = $txt.Trim();
[System.IO.File]::WriteAllText('distribution\README.txt', $txt, [System.Text.Encoding]::UTF8)
"
IF !ERRORLEVEL! EQU 0 (
    echo ✅ README.txt generato con successo
) ELSE (
    echo ⚠️ Conversione README fallita ma continuo
)anager

@echo off
SETLOCAL ENABLEDELAYEDEXPANSION

echo 🔧 Java Version Manager - Build Script
echo ======================================

:: Step 1: Build Go binary
echo ► Building jvm.exe...
go build -o distribution\jvm.exe main.go
IF !ERRORLEVEL! NEQ 0 (
    echo ❌ Build fallito: go build ha restituito errore !ERRORLEVEL!
    goto error
)
IF NOT EXIST distribution\jvm.exe (
    echo ❌ Build fallito: jvm.exe non trovato dopo la compilazione
    goto error
)
echo ✅ jvm.exe compilato con successo

:: Step 2: Sign jvm.exe with certificate
SET "SIGNTOOL=C:\Program Files (x86)\Windows Kits\10\bin\x64\signtool.exe"
IF EXIST "%SIGNTOOL%" (
    echo ► Signing jvm.exe...
    "%SIGNTOOL%" sign /f distribution\jvm-dev-cert.pfx /p jvm-password /tr http://timestamp.digicert.com /td sha256 distribution\jvm.exe
    IF !ERRORLEVEL! EQU 0 (
        echo ✅ jvm.exe firmato con successo
    ) ELSE (
        echo ⚠️ Firma fallita ma continuo (errore !ERRORLEVEL!)
    )
) ELSE (
    echo ⚠️ SignTool non trovato. Salta firma di jvm.exe.
)

echo.
echo ⏱️ Attendo il rilascio dei file lock...
echo 💡 Se hai VS Code aperto con file dalla cartella distribution, chiudilo ora!
timeout /t 5 /nobreak >nul

:: Remove old installer if exists
IF EXIST distribution\jvm-installer.exe (
    echo 🗑️ Rimuovo il vecchio installer...
    del /f distribution\jvm-installer.exe >nul 2>&1
    IF !ERRORLEVEL! EQU 0 (
        echo ✅ Vecchio installer rimosso
    ) ELSE (
        echo ⚠️ Non posso rimuovere il vecchio installer (potrebbe essere in uso)
    )
)

:: Step 2.5: Convert README.md to README.txt for distribution
echo ► Converting README.md to distribution/README.txt...
powershell -Command "
$md = Get-Content 'README.md' -Raw -Encoding UTF8;
$txt = $md -replace '^#{1,6}\s*(.*)$', '$1' -replace '\*\*([^*]+)\*\*', '$1' -replace '\*([^*]+)\*', '$1' -replace '`([^`]+)`', '$1' -replace '\[([^\]]+)\]\([^\)]+\)', '$1' -replace '^[\s]*-\s*', '* ' -replace '```[^`]*```', '' -replace '`{4}[^`]*`{4}', '' -replace '\n{3,}', [char]10+[char]10;
$txt = $txt.Trim();
[System.IO.File]::WriteAllText('distribution\README.txt', $txt, [System.Text.Encoding]::UTF8)
"
IF !ERRORLEVEL! EQU 0 (
    echo ✅ README.txt generato con successo
) ELSE (
    echo ⚠️ Conversione README fallita ma continuo
)

:: Step 3: Compile installer with Inno Setup
SET "INNO=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
IF EXIST "%INNO%" (
    echo ► Compiling installer...
    "%INNO%" setup.iss
    IF !ERRORLEVEL! EQU 0 (
        echo ✅ Installer compilato con successo
    ) ELSE (
        echo ❌ Compilazione installer fallita (errore !ERRORLEVEL!)
        echo 💡 Possibili soluzioni:
        echo    • Chiudi VS Code completamente
        echo    • Chiudi tutti i terminali che potrebbero avere handle sui file
        echo    • Riavvia il sistema se il problema persiste
        goto error
    )
) ELSE (
    echo ❌ ISCC.exe non trovato. Installa Inno Setup Compiler.
    goto error
)

echo.
echo 🎉 Build completo! Controlla la cartella distribution\
echo 📦 File generati:
echo    - jvm.exe (eseguibile principale)
echo    - jvm-installer.exe (installer)
echo.
echo 🚀 Esempi di test:
echo    .\distribution\jvm.exe remote-list
echo    .\distribution\jvm-installer.exe /CONFIGURE_PRIVATE=0
goto end

:error
echo.
echo ❌ Build interrotto a causa di errori.
exit /b 1

:end
pause
ENDLOCAL