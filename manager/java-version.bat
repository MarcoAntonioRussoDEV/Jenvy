@echo off
setlocal enabledelayedexpansion

:: Abilita supporto colori ANSI
for /f %%A in ('"prompt $E & for %%B in (1) do rem"') do set "ESC=%%A"

:: Verifica se è in esecuzione come amministratore
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo %ESC%[91m✖ Devi eseguire questo script come amministratore!%ESC%[0m
    echo %ESC%[93mChiudi e avvia tramite launcher.vbs oppure usa 'Esegui come amministratore'.%ESC%[0m
    pause
    exit /b
)

echo %ESC%[93m======================================================%ESC%[0m
echo %ESC%[91m   IMPOSTAZIONE JAVA_HOME - Avvio come AMMINISTRATORE   %ESC%[0m
echo %ESC%[93m======================================================%ESC%[0m
echo.

set "JAVA_DIR=C:\Program Files\Java"
set /a INDEX=0

echo %ESC%[96mScansione delle versioni Java in:%ESC%[0m %ESC%[92m%JAVA_DIR%%ESC%[0m
echo.

if not exist "%JAVA_DIR%" (
    echo %ESC%[91m❌ Errore:%ESC%[0m cartella %JAVA_DIR% non trovata.
    pause
    exit /b
)

for /d %%D in ("%JAVA_DIR%\*") do (
    echo %ESC%[93m!INDEX!%ESC%[0m - %%~nxD
    set "JAVA_PATH_!INDEX!=%%~fD"
    set /a INDEX+=1
)

echo.
set /p CHOICE=%ESC%[96m👉 Inserisci il numero della versione da impostare:%ESC%[0m 

call set "SELECTED_PATH=%%JAVA_PATH_%CHOICE%%%"

if not exist "!SELECTED_PATH!" (
    echo %ESC%[91m❌ Selezione non valida o percorso non trovato.%ESC%[0m
    pause
    exit /b
)

setx JAVA_HOME "!SELECTED_PATH!" /M >nul

echo.
echo %ESC%[92m✅ JAVA_HOME impostato a:%ESC%[0m !SELECTED_PATH!
echo %ESC%[96m🔁 Riavvia il terminale o il sistema per applicare le modifiche.%ESC%[0m
pause
