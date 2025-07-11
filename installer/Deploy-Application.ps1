# PowerShell App Deployment Toolkit - Deploy-Application.ps1
# Java Version Manager Installation Script

[CmdletBinding()]
Param (
    [Parameter(Mandatory=$false)]
    [ValidateSet('Install','Uninstall','Repair')]
    [string]$DeploymentType = 'Install',
    [Parameter(Mandatory=$false)]
    [ValidateSet('Interactive','Silent','NonInteractive')]
    [string]$DeployMode = 'Interactive'
)

#region Initialization
$scriptDirectory = Split-Path -Parent $MyInvocation.MyCommand.Definition
$appName = 'Java Version Manager'
$appVersion = '1.0.0'
$appVendor = 'Marco Russo'
$appScriptVersion = '1.0.0'
$appScriptDate = '11/07/2025'
$appScriptAuthor = 'Marco Russo'

# Importa le funzioni del toolkit (assumendo che sia disponibile)
# . "$scriptDirectory\AppDeployToolkit\AppDeployToolkitMain.ps1"

Write-Log -Message "Starting deployment of $appName version $appVersion" -Severity 1

#endregion

#region Pre-Installation
If ($deploymentType -ieq 'Install') {
    [string]$installPhase = 'Pre-Installation'
    
    # Verifica i prerequisiti
    Write-Log -Message "Verifico i prerequisiti..." -Severity 1
    
    # Controlla se PowerShell è disponibile (dovrebbe sempre esserci)
    if (-not (Get-Command powershell.exe -ErrorAction SilentlyContinue)) {
        Write-Log -Message "PowerShell non trovato nel sistema." -Severity 3
        Exit-Script -ExitCode 1601
    }
    
    # Controlla se l'utente ha privilegi di amministratore
    if (-not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
        Write-Log -Message "Privilegi di amministratore richiesti." -Severity 2
        Exit-Script -ExitCode 1603
    }
    
    Write-Log -Message "Prerequisiti verificati con successo." -Severity 1
}

#endregion

#region Installation
If ($deploymentType -ieq 'Install') {
    [string]$installPhase = 'Installation'
    
    # Crea la directory di installazione
    $installPath = "$env:ProgramFiles\$appName"
    New-Item -ItemType Directory -Path $installPath -Force | Out-Null
    New-Item -ItemType Directory -Path "$installPath\manager" -Force | Out-Null
    New-Item -ItemType Directory -Path "$installPath\manager\assets" -Force | Out-Null
    
    Write-Log -Message "Copiando i file dell'applicazione..." -Severity 1
    
    # Copia i file principali
    Copy-File -Path "$scriptDirectory\Files\launcher.vbs" -Destination "$installPath\launcher.vbs"
    Copy-File -Path "$scriptDirectory\Files\launcher-cmd.vbs" -Destination "$installPath\launcher-cmd.vbs"
    Copy-File -Path "$scriptDirectory\Files\README.txt" -Destination "$installPath\README.txt"
    
    # Copia i file del manager
    Copy-File -Path "$scriptDirectory\Files\manager\java-manager.ps1" -Destination "$installPath\manager\java-manager.ps1"
    Copy-File -Path "$scriptDirectory\Files\manager\java-version.bat" -Destination "$installPath\manager\java-version.bat"
    Copy-File -Path "$scriptDirectory\Files\manager\assets\java.ico" -Destination "$installPath\manager\assets\java.ico"
    
    # Crea i collegamenti
    Write-Log -Message "Creando i collegamenti..." -Severity 1
    
    # Menu Start
    New-Shortcut -Path "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\$appName.lnk" `
                 -TargetPath "$installPath\launcher.vbs" `
                 -IconLocation "$installPath\manager\assets\java.ico" `
                 -Description "Gestore versioni Java"
    
    # Desktop (opzionale)
    New-Shortcut -Path "$env:Public\Desktop\$appName.lnk" `
                 -TargetPath "$installPath\launcher.vbs" `
                 -IconLocation "$installPath\manager\assets\java.ico" `
                 -Description "Gestore versioni Java"
    
    # Registrazione nel registro di Windows
    Write-Log -Message "Registrando l'applicazione nel sistema..." -Severity 1
    
    $uninstallKey = "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\$appName"
    New-Item -Path $uninstallKey -Force | Out-Null
    Set-ItemProperty -Path $uninstallKey -Name "DisplayName" -Value $appName
    Set-ItemProperty -Path $uninstallKey -Name "DisplayVersion" -Value $appVersion
    Set-ItemProperty -Path $uninstallKey -Name "Publisher" -Value $appVendor
    Set-ItemProperty -Path $uninstallKey -Name "InstallLocation" -Value $installPath
    Set-ItemProperty -Path $uninstallKey -Name "UninstallString" -Value "powershell.exe -File `"$installPath\Uninstall.ps1`""
    Set-ItemProperty -Path $uninstallKey -Name "DisplayIcon" -Value "$installPath\manager\assets\java.ico"
    
    Write-Log -Message "Installazione completata con successo." -Severity 1
}

#endregion

#region Uninstallation
If ($deploymentType -ieq 'Uninstall') {
    [string]$installPhase = 'Uninstallation'
    
    $installPath = "$env:ProgramFiles\$appName"
    
    Write-Log -Message "Rimuovendo l'applicazione..." -Severity 1
    
    # Rimuovi i collegamenti
    Remove-File -Path "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\$appName.lnk"
    Remove-File -Path "$env:Public\Desktop\$appName.lnk"
    Remove-File -Path "$env:Public\Desktop\Gestione Java.lnk"  # Collegamento creato dall'app
    
    # Rimuovi i file dell'applicazione
    Remove-Folder -Path $installPath -Recurse
    
    # Rimuovi la registrazione dal registro
    Remove-RegistryKey -Key "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\$appName"
    
    Write-Log -Message "Disinstallazione completata con successo." -Severity 1
}

#endregion

#region Post-Installation
If ($deploymentType -ieq 'Install') {
    [string]$installPhase = 'Post-Installation'
    
    # Mostra un messaggio di completamento
    if ($deployMode -ne 'Silent') {
        Show-InstallationPrompt -Message "Java Version Manager è stato installato con successo!`n`nPuoi trovare il programma nel menu Start o utilizzare il collegamento sul desktop." `
                               -ButtonRightText "OK" -Icon Information
    }
}

#endregion

Write-Log -Message "Deployment di $appName completato." -Severity 1
Exit-Script -ExitCode 0
