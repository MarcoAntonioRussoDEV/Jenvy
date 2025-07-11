; JavaVersionManager Installer Script
; Creato con NSIS (Nullsoft Scriptable Install System)

!define APPNAME "Java Version Manager"
!define COMPANYNAME "Marco Russo"
!define DESCRIPTION "Gestore semplice per versioni Java"
!define VERSIONMAJOR 1
!define VERSIONMINOR 0
!define VERSIONBUILD 0

!include "MUI2.nsh"
!include "UAC.nsh"

; Impostazioni generali
Name "${APPNAME}"
OutFile "JavaVersionManager-Setup.exe"
InstallDir "$PROGRAMFILES\${APPNAME}"
RequestExecutionLevel admin

; Interfaccia Modern UI
!define MUI_ABORTWARNING
!define MUI_ICON "..\manager\assets\java.ico"
!define MUI_UNICON "..\manager\assets\java.ico"

; Pagine dell'installer
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "license.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; Pagine del disinstaller
!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Lingue
!insertmacro MUI_LANGUAGE "Italian"
!insertmacro MUI_LANGUAGE "English"

; Sezione principale di installazione
Section "JavaVersionManager" SecMain
    SetOutPath $INSTDIR
    
    ; Copia tutti i file
    File "..\launcher.vbs"
    File "..\launcher-cmd.vbs"
    File "..\README.txt"
    
    ; Crea cartella manager
    CreateDirectory "$INSTDIR\manager"
    SetOutPath "$INSTDIR\manager"
    File "..\manager\java-manager.ps1"
    File "..\manager\java-version.bat"
    
    ; Crea cartella assets
    CreateDirectory "$INSTDIR\manager\assets"
    SetOutPath "$INSTDIR\manager\assets"
    File "..\manager\assets\java.ico"
    
    ; Crea collegamento nel menu Start
    CreateDirectory "$SMPROGRAMS\${APPNAME}"
    CreateShortCut "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk" "$INSTDIR\launcher.vbs" "" "$INSTDIR\manager\assets\java.ico"
    CreateShortCut "$SMPROGRAMS\${APPNAME}\Uninstall.lnk" "$INSTDIR\uninstall.exe"
    
    ; Crea collegamento sul Desktop
    CreateShortCut "$DESKTOP\${APPNAME}.lnk" "$INSTDIR\launcher.vbs" "" "$INSTDIR\manager\assets\java.ico"
    
    ; Registrazione nel Pannello di Controllo
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayName" "${APPNAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "InstallLocation" "$\"$INSTDIR$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayIcon" "$\"$INSTDIR\manager\assets\java.ico$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "Publisher" "${COMPANYNAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "VersionMajor" ${VERSIONMAJOR}
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "VersionMinor" ${VERSIONMINOR}
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoRepair" 1
    
    ; Crea uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

; Sezione di disinstallazione
Section "Uninstall"
    ; Rimuovi file
    Delete "$INSTDIR\launcher.vbs"
    Delete "$INSTDIR\launcher-cmd.vbs"
    Delete "$INSTDIR\README.txt"
    Delete "$INSTDIR\manager\java-manager.ps1"
    Delete "$INSTDIR\manager\java-version.bat"
    Delete "$INSTDIR\manager\assets\java.ico"
    Delete "$INSTDIR\uninstall.exe"
    
    ; Rimuovi cartelle
    RMDir "$INSTDIR\manager\assets"
    RMDir "$INSTDIR\manager"
    RMDir "$INSTDIR"
    
    ; Rimuovi collegamenti
    Delete "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk"
    Delete "$SMPROGRAMS\${APPNAME}\Uninstall.lnk"
    RMDir "$SMPROGRAMS\${APPNAME}"
    Delete "$DESKTOP\${APPNAME}.lnk"
    Delete "$DESKTOP\Gestione Java.lnk"
    
    ; Rimuovi registro
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}"
SectionEnd
