# 🚀 GUIDA ALLA CREAZIONE DI UN INSTALLER PROFESSIONALE

# Java Version Manager - Da Script a Programma Installabile

## 📋 OPZIONI DISPONIBILI

### 1. **NSIS (CONSIGLIATO) - Gratuito e Potente**

**Vantaggi:**
✅ Completamente gratuito
✅ Installer .exe professionale
✅ Supporto completo per Windows 10/11
✅ Interfaccia moderna con Modern UI
✅ Supporto multilingua
✅ Registrazione automatica nel Pannello di Controllo
✅ Creazione automatica di collegamenti
✅ Disinstaller automatico

**Come usare:**

1. Scarica NSIS da: https://nsis.sourceforge.io/
2. Installa NSIS sul tuo sistema
3. Usa il file `JavaVersionManager.nsi` che ho creato
4. Fai clic destro sul file .nsi → "Compile NSIS Script"
5. Otterrai un file `JavaVersionManager-Setup.exe`

**Comando da terminale:**

```bash
"C:\Program Files (x86)\NSIS\makensis.exe" JavaVersionManager.nsi
```

### 2. **Inno Setup - Alternativa User-Friendly**

**Vantaggi:**
✅ Gratuito
✅ Interfaccia grafica per creare installer
✅ Wizard integrato
✅ Ottima documentazione

**Come usare:**

1. Scarica Inno Setup da: https://jrsoftware.org/isinfo.php
2. Apri il file `JavaVersionManager.iss` con Inno Setup
3. Compila → Otterrai l'installer .exe

### 3. **Advanced Installer - Professionale**

**Vantaggi:**
✅ Interfaccia molto user-friendly
✅ Supporto per Windows Store
✅ Aggiornamenti automatici
✅ Certificazione digitale integrata

**Svantaggi:**
❌ A pagamento per funzioni avanzate
❌ Free edition limitata

### 4. **PowerShell App Deployment Toolkit**

**Vantaggi:**
✅ Specifico per ambienti aziendali
✅ Integrazione con SCCM
✅ Logging avanzato
✅ Gestione automatica delle dipendenze

**Come usare:**

1. Scarica PSADT da: https://psappdeploytoolkit.com/
2. Usa il file `Deploy-Application.ps1` che ho creato
3. Personalizza e distribuisci

## 🛠️ PASSAGGI PER CREARE L'INSTALLER CON NSIS

### Passo 1: Preparazione

```
JavaVersionManager/
├── installer/
│   ├── JavaVersionManager.nsi    ← Script principale
│   ├── license.txt               ← Licenza
│   └── build.bat                 ← Script di compilazione
├── launcher.vbs
├── launcher-cmd.vbs
├── README.txt
└── manager/
    ├── java-manager.ps1
    ├── java-version.bat
    └── assets/
        └── java.ico
```

### Passo 2: Compilazione

1. Installa NSIS
2. Apri Command Prompt nella cartella installer/
3. Esegui:

```cmd
"C:\Program Files (x86)\NSIS\makensis.exe" JavaVersionManager.nsi
```

### Passo 3: Risultato

Otterrai `JavaVersionManager-Setup.exe` - un installer professionale!

## 🔧 FUNZIONALITÀ DELL'INSTALLER

### ✨ Cosa fa l'installer:

-   **Installazione guidata** con interfaccia moderna
-   **Controlla privilegi amministratore** automaticamente
-   **Crea cartella in Program Files** con struttura corretta
-   **Registra nel Pannello di Controllo** per disinstallazione
-   **Crea collegamenti** in Menu Start e Desktop
-   **Associa icona personalizzata** java.ico
-   **Genera disinstaller automatico**
-   **Supporta installazione silenziosa** con `/S`

### 🗑️ Cosa fa il disinstaller:

-   **Rimuove tutti i file** installati
-   **Cancella collegamenti** (Menu Start, Desktop)
-   **Pulisce il registro** di Windows
-   **Rimuove cartelle** create

## 📦 ALTERNATIVE MODERNE

### **1. MSIX (Windows 10/11)**

```powershell
# Crea un pacchetto MSIX moderno
New-MsixPackage -Source "C:\JavaVersionManager" -Destination "JavaVersionManager.msix"
```

### **2. ClickOnce (.NET)**

Se ricostruisci in .NET, puoi usare ClickOnce per aggiornamenti automatici.

### **3. Windows Package Manager (winget)**

Puoi pubblicare su winget per distribuzione tramite:

```cmd
winget install JavaVersionManager
```

## 🎯 RACCOMANDAZIONE FINALE

**Per il tuo progetto, ti consiglio NSIS perché:**

1. ✅ È completamente gratuito
2. ✅ Crea installer professionali
3. ✅ È lo standard per software Windows
4. ✅ Ha un'ottima community e documentazione
5. ✅ Supporta tutte le funzionalità moderne

## 🚀 PROSSIMI PASSI

1. **Installa NSIS** dal sito ufficiale
2. **Testa il file .nsi** che ho creato
3. **Personalizza l'interfaccia** se necessario
4. **Compila l'installer**
5. **Testa su una macchina pulita**
6. **Distribuisci il tuo .exe**

## 📝 VERSIONING E DISTRIBUZIONE

Una volta che hai l'installer, puoi:

-   **Pubblicare su GitHub** con releases
-   **Creare un sito web** per il download
-   **Distribuire in azienda** via Group Policy
-   **Pubblicare su software directories** come FileHippo, Softonic

Il tuo progetto è già molto ben strutturato - con un installer diventerà un software completamente professionale! 🎉
