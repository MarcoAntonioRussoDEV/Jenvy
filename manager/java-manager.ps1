# ▶️ Inizio registrazione log per debug
Start-Transcript -Path "$env:TEMP\java-manager-log.txt" -Append

# 📁 Percorsi base
$scriptDir     = Split-Path -Parent $MyInvocation.MyCommand.Definition
$iconPath      = Join-Path $scriptDir "assets\java.ico"
$launcherPath  = Join-Path $scriptDir "..\launcher.vbs"
$linkName      = "Gestione Java.lnk"
$javaDir       = "C:\Program Files\Java"

# 🚫 Controlla se la cartella Java esiste
if (-not (Test-Path $javaDir)) {
    Write-Host "Cartella '$javaDir' non trovata." -ForegroundColor Red
    Read-Host -Prompt "Premi INVIO per uscire"
    Stop-Transcript
    exit
}

# 🔍 Recupera le versioni Java
$versions = Get-ChildItem -Path $javaDir -Directory
if (-not $versions) {
    Write-Host "Nessuna versione Java trovata." -ForegroundColor Yellow
    Read-Host -Prompt "Premi INVIO per uscire"
    Stop-Transcript
    exit
}

# 📋 Menu grafico per selezione
$selected = $versions | Select-Object Name, FullName | Out-GridView -Title "Seleziona la versione Java da impostare" -PassThru
if (-not $selected) {
    Write-Host "Nessuna selezione effettuata." -ForegroundColor Yellow
    Read-Host -Prompt "Premi INVIO per uscire"
    Stop-Transcript
    exit
}

# ⚙️ Imposta JAVA_HOME a livello macchina
[System.Environment]::SetEnvironmentVariable("JAVA_HOME", $selected.FullName, "Machine")
Write-Host ""
Write-Host "JAVA_HOME impostato a: $($selected.FullName)" -ForegroundColor Green

# 📎 Crea collegamento sul desktop
$WshShell  = New-Object -ComObject WScript.Shell
$desktop   = [Environment]::GetFolderPath("Desktop")
$shortcut  = $WshShell.CreateShortcut("$desktop\$linkName")
$shortcut.TargetPath       = $launcherPath
$shortcut.IconLocation     = $iconPath
$shortcut.WorkingDirectory = "$scriptDir\.."
$shortcut.WindowStyle      = 1
$shortcut.Description      = "Impostazione rapida JAVA_HOME"
$shortcut.Save()

Write-Host ""
Write-Host "Collegamento 'Gestione Java' creato sul Desktop con icona." -ForegroundColor Cyan

# Chiusura log e pausa finale  
Stop-Transcript
Read-Host -Prompt 'Premi INVIO per chiudere'