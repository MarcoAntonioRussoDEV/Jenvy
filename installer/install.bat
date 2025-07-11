@echo off
title Java Version Manager - Installer
color 0A

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                    JAVA VERSION MANAGER                       ║
echo ║                         INSTALLER                             ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

:: Verifica privilegi amministratore
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ ERRORE: Questo installer richiede privilegi di amministratore.
    echo    Fai clic destro sull'installer e seleziona "Esegui come amministratore"
    echo.
    pause
    exit /b 1
)

:: Variabili
set "INSTALL_DIR=%ProgramFiles%\Java Version Manager"
set "START_MENU=%ProgramData%\Microsoft\Windows\Start Menu\Programs"
set "DESKTOP=%PUBLIC%\Desktop"

echo ⚙️  Preparando l'installazione...
echo.

:: Crea la directory di installazione
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
if not exist "%INSTALL_DIR%\manager" mkdir "%INSTALL_DIR%\manager"
if not exist "%INSTALL_DIR%\manager\assets" mkdir "%INSTALL_DIR%\manager\assets"

echo 📁 Copiando i file...

:: Copia i file principali
copy /Y "launcher.vbs" "%INSTALL_DIR%\" >nul
copy /Y "launcher-cmd.vbs" "%INSTALL_DIR%\" >nul
copy /Y "README.txt" "%INSTALL_DIR%\" >nul

:: Copia i file del manager
copy /Y "manager\java-manager.ps1" "%INSTALL_DIR%\manager\" >nul
copy /Y "manager\java-version.bat" "%INSTALL_DIR%\manager\" >nul
copy /Y "manager\assets\java.ico" "%INSTALL_DIR%\manager\assets\" >nul

echo 🔗 Creando i collegamenti...

:: Crea collegamento nel menu Start
powershell -Command "$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%START_MENU%\Java Version Manager.lnk'); $s.TargetPath = '%INSTALL_DIR%\launcher.vbs'; $s.IconLocation = '%INSTALL_DIR%\manager\assets\java.ico'; $s.Description = 'Gestore versioni Java'; $s.Save()"

:: Crea collegamento sul Desktop
powershell -Command "$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%DESKTOP%\Java Version Manager.lnk'); $s.TargetPath = '%INSTALL_DIR%\launcher.vbs'; $s.IconLocation = '%INSTALL_DIR%\manager\assets\java.ico'; $s.Description = 'Gestore versioni Java'; $s.Save()"

echo 📝 Registrando nel sistema...

:: Registra nel Pannello di Controllo
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\JavaVersionManager" /v "DisplayName" /t REG_SZ /d "Java Version Manager" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\JavaVersionManager" /v "DisplayVersion" /t REG_SZ /d "1.0.0" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\JavaVersionManager" /v "Publisher" /t REG_SZ /d "Marco Russo" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\JavaVersionManager" /v "InstallLocation" /t REG_SZ /d "%INSTALL_DIR%" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\JavaVersionManager" /v "UninstallString" /t REG_SZ /d "\"%INSTALL_DIR%\uninstall.bat\"" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\JavaVersionManager" /v "DisplayIcon" /t REG_SZ /d "%INSTALL_DIR%\manager\assets\java.ico" /f >nul

:: Crea script di disinstallazione
echo @echo off > "%INSTALL_DIR%\uninstall.bat"
echo title Java Version Manager - Disinstallazione >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo echo Disinstallazione di Java Version Manager... >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo del /Q "%START_MENU%\Java Version Manager.lnk" ^>nul 2^>^&1 >> "%INSTALL_DIR%\uninstall.bat"
echo del /Q "%DESKTOP%\Java Version Manager.lnk" ^>nul 2^>^&1 >> "%INSTALL_DIR%\uninstall.bat"
echo del /Q "%DESKTOP%\Gestione Java.lnk" ^>nul 2^>^&1 >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\JavaVersionManager" /f ^>nul 2^>^&1 >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo cd /d "%ProgramFiles%" >> "%INSTALL_DIR%\uninstall.bat"
echo rmdir /S /Q "Java Version Manager" >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo echo Disinstallazione completata. >> "%INSTALL_DIR%\uninstall.bat"
echo pause >> "%INSTALL_DIR%\uninstall.bat"

echo.
echo ✅ INSTALLAZIONE COMPLETATA!
echo.
echo 📍 Il programma è stato installato in: %INSTALL_DIR%
echo 🔗 Troverai i collegamenti in:
echo    • Menu Start ^> Java Version Manager
echo    • Desktop ^> Java Version Manager
echo.
echo 🚀 Per iniziare, fai doppio clic sul collegamento o vai al menu Start.
echo.

choice /C YN /M "Vuoi avviare Java Version Manager ora? (Y/N)"
if errorlevel 2 goto :end
if errorlevel 1 start "" "%INSTALL_DIR%\launcher.vbs"

:end
echo.
echo Grazie per aver installato Java Version Manager!
pause
