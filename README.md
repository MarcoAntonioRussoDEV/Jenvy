# ☕ Java Version Manager

<p align="center">
  <img src="manager/assets/java.ico" alt="Java Version Manager" width="64"/>
</p>

<p align="center">
  <strong>Gestore semplice e intuitivo per le versioni Java installate su Windows</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Windows-blue.svg" alt="Platform Windows"/>
  <img src="https://img.shields.io/badge/Language-PowerShell-blue.svg" alt="Language PowerShell"/>
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License MIT"/>
  <img src="https://img.shields.io/badge/Version-1.0.0-orange.svg" alt="Version 1.0.0"/>
</p>

---

## 📋 Caratteristiche

-   ✅ **Interfaccia grafica** per la selezione delle versioni Java
-   ✅ **Cambio automatico** della variabile `JAVA_HOME`
-   ✅ **Privilegi amministratore** gestiti automaticamente
-   ✅ **Collegamenti desktop** per accesso rapido
-   ✅ **Due modalità**: GUI (PowerShell) e CLI (Batch)
-   ✅ **Logging completo** per debug e troubleshooting

## 📁 Struttura del Progetto

```
JavaVersionManager/
├── 📄 launcher.vbs              # Avvio principale (PowerShell GUI con diritti admin)
├── 📄 launcher-cmd.vbs          # Avvio alternativo (Batch con diritti admin)
├── 📄 README.md                 # Questo file!
├── 📂 manager/
│   ├── 📜 java-manager.ps1      # Script PowerShell con selezione grafica
│   ├── 📜 java-version.bat      # Script compatibile con cmd
│   └── 📂 assets/
│       └── 🎨 java.ico          # Icona personalizzata
└── 📂 installer/                # Script per creare installer professionale
    ├── 📜 JavaVersionManager.nsi # Script NSIS per installer
    ├── 📜 build.bat             # Script di compilazione
    └── 📄 README-INSTALLER.md   # Guida per creare installer
```

## 🚀 Installazione e Uso

### Metodo 1: Utilizzo Diretto

1. **Scarica** o clona questo repository
2. **Fai doppio clic** su `launcher.vbs`
3. **Conferma** i privilegi di amministratore quando richiesto
4. **Seleziona** la versione Java desiderata dalla finestra grafica
5. **Conferma** la selezione

### Metodo 2: Installer Professionale

1. Vai nella cartella `installer/`
2. Segui le istruzioni in `README-INSTALLER.md`
3. Crea un installer .exe professionale con NSIS

## 🧩 Come Funziona

### 🖼️ Modalità GUI (Consigliata)

-   **Avvia**: `launcher.vbs`
-   **Interfaccia**: Finestra grafica con Out-GridView
-   **Selezione**: Click sulla versione desiderata
-   **Risultato**: `JAVA_HOME` aggiornato + collegamento desktop creato

### 💻 Modalità CLI (Alternativa)

-   **Avvia**: `launcher-cmd.vbs`
-   **Interfaccia**: Menu testuale nel terminale
-   **Selezione**: Digitare il numero della versione
-   **Risultato**: `JAVA_HOME` aggiornato

## 📋 Prerequisiti

-   🖥️ **Windows 7** o superiore
-   ☕ **Almeno una versione di Java** installata in `C:\Program Files\Java\`
-   🔐 **Privilegi di amministratore** (per modificare variabili d'ambiente di sistema)
-   🐚 **PowerShell** (incluso in Windows per default)

## 🔧 Configurazione

Il tool cerca automaticamente le installazioni Java in:

-   `C:\Program Files\Java\`

Se hai Java installato in percorsi diversi, puoi modificare la variabile `$javaDir` in `java-manager.ps1`.

## 📸 Screenshots

### GUI Mode

![GUI Mode](https://via.placeholder.com/600x400/0078D4/FFFFFF?text=Out-GridView+Selection)

### CLI Mode

![CLI Mode](https://via.placeholder.com/600x400/000000/00FF00?text=Terminal+Selection)

## 🎯 Funzionalità Avanzate

-   **🔍 Auto-discovery**: Rileva automaticamente tutte le versioni Java installate
-   **📝 Logging**: Log completo delle operazioni in `%TEMP%\java-manager-log.txt`
-   **🔗 Shortcut automatici**: Crea collegamenti desktop con icona personalizzata
-   **⚡ UAC Handling**: Gestione automatica dei privilegi amministratore
-   **🔄 Backup**: Non modifica le installazioni esistenti, solo la variabile d'ambiente

## 🛠️ Sviluppo

### Struttura del Codice

-   **VBScript**: Launcher per elevazione privilegi
-   **PowerShell**: Logica principale e GUI
-   **Batch**: Alternativa CLI compatibile con cmd

### Build e Distribuzione

```bash
# Clona il repository
git clone https://github.com/MarcoAntonioRussoDEV/JavaVersionManager.git

# Testa il funzionamento
cd JavaVersionManager
./launcher.vbs

# Crea installer (opzionale)
cd installer
./build.bat
```

## 🐛 Troubleshooting

### Problema: "Nessuna versione Java trovata"

-   **Causa**: Java non installato in `C:\Program Files\Java\`
-   **Soluzione**: Verifica il percorso di installazione Java

### Problema: "Privilegi amministratore richiesti"

-   **Causa**: UAC non autorizzato
-   **Soluzione**: Fai clic destro → "Esegui come amministratore"

### Problema: "PowerShell bloccato da Execution Policy"

-   **Causa**: Criteri di esecuzione PowerShell restrittivi
-   **Soluzione**: Il launcher usa `-ExecutionPolicy Bypass` automaticamente

## 📞 Supporto

-   🐛 **Bug Reports**: [Issues](https://github.com/MarcoAntonioRussoDEV/JavaVersionManager/issues)
-   💡 **Feature Requests**: [Discussions](https://github.com/MarcoAntonioRussoDEV/JavaVersionManager/discussions)
-   📧 **Contatto**: [Marco Antonio Russo](mailto:your.email@example.com)

## 🤝 Contribuire

I contributi sono benvenuti! Per contribuire:

1. **Fork** il progetto
2. **Crea** un branch per la tua feature (`git checkout -b feature/AmazingFeature`)
3. **Commit** le modifiche (`git commit -m 'Add some AmazingFeature'`)
4. **Push** al branch (`git push origin feature/AmazingFeature`)
5. **Apri** una Pull Request

## 📄 Licenza

Questo progetto è distribuito sotto licenza **MIT**. Vedi il file [LICENSE](LICENSE) per maggiori dettagli.

## 🙏 Ringraziamenti

-   Microsoft per PowerShell e Out-GridView
-   Community NSIS per gli strumenti di packaging
-   Tutti i contributor e tester

---

<p align="center">
  <strong>Fatto con ❤️ da <a href="https://github.com/MarcoAntonioRussoDEV">Marco Antonio Russo</a></strong>
</p>

<p align="center">
  ⭐ Se questo progetto ti è stato utile, lascia una stella!
</p>
