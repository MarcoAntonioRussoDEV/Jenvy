@echo off
title Compilazione Java Version Manager Installer
color 0B

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║              JAVA VERSION MANAGER - BUILD SCRIPT              ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

:: Verifica se NSIS è installato
set "NSIS_PATH=C:\Program Files (x86)\NSIS\makensis.exe"
if not exist "%NSIS_PATH%" (
    echo ❌ NSIS non trovato in %NSIS_PATH%
    echo.
    echo 📥 Per installare NSIS:
    echo    1. Vai su https://nsis.sourceforge.io/
    echo    2. Scarica e installa NSIS
    echo    3. Riavvia questo script
    echo.
    pause
    exit /b 1
)

echo ✅ NSIS trovato: %NSIS_PATH%
echo.

:: Verifica se il file .nsi esiste
if not exist "JavaVersionManager.nsi" (
    echo ❌ File JavaVersionManager.nsi non trovato nella cartella corrente
    echo    Assicurati di essere nella cartella installer/
    echo.
    pause
    exit /b 1
)

echo 🔧 Compilando l'installer...
echo.

:: Compila l'installer
"%NSIS_PATH%" JavaVersionManager.nsi

if %errorlevel% equ 0 (
    echo.
    echo ✅ COMPILAZIONE COMPLETATA CON SUCCESSO!
    echo.
    echo 📦 L'installer è stato creato: JavaVersionManager-Setup.exe
    echo.
    echo 🚀 Ora puoi:
    echo    • Testare l'installer su questo computer
    echo    • Distribuire il file .exe ad altri utenti
    echo    • Pubblicare su GitHub Releases
    echo.
    
    :: Chiedi se vuoi testare l'installer
    choice /C YN /M "Vuoi testare l'installer ora? (Y/N)"
    if errorlevel 2 goto :end
    if errorlevel 1 (
        if exist "JavaVersionManager-Setup.exe" (
            echo.
            echo 🧪 Avvio dell'installer per test...
            start "" "JavaVersionManager-Setup.exe"
        ) else (
            echo ❌ File installer non trovato
        )
    )
) else (
    echo.
    echo ❌ ERRORE DURANTE LA COMPILAZIONE
    echo    Controlla i messaggi di errore sopra
    echo.
)

:end
echo.
pause
