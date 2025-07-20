# ☕ Java Version Manager (JVM)

Una CLI elegante per esplorare, filtrare e gestire versioni di OpenJDK da provider pubblici e repository privati — con interfaccia grafica testuale e comportamento smart.

---

## 🚀 Funzionalità principali

-   🔍 Elenco JDK da **Adoptium**, **Azul**, **Liberica** e **repository privati**
-   🧠 Selezione intelligente di una versione per tag (LTS → usata → patch)
-   📊 Visualizzazione tabellare con info su OS, Architettura, link di download
-   📄 Banner Figlet all’avvio + spiegazione del comportamento
-   🎛️ Supporto a flag avanzati:
    -   `--provider`, `--all`, `--major-only`, `--latest`, `--jdk`, `--lts-only`
-   🛡️ Supporto a repository privati con configurazione:
    -   Via `~/.jvm/config.json`
    -   Via variabili d’ambiente `JVM_PRIVATE_ENDPOINT` e `JVM_PRIVATE_TOKEN`
-   📦 Comandi ausiliari:
    -   `configure-private`: genera `config.json`
    -   `show-config`: visualizza configurazione attuale
    -   `reset-config`: cancella configurazione privata

---

## 📦 Installazione

### Windows (.exe globale)

1. Scarica l’installer firmato `jvm-installer.exe`
2. Esegui come amministratore
3. Il tool sarà disponibile in qualsiasi terminale come `jvm`

---

## 🔧 Esempi d’uso

```bash
jvm remote-list                    # selezione smart per Adoptium
jvm remote-list --provider=azul   # provider alternativo
jvm remote-list --all             # mostra versioni da tutti i provider
jvm remote-list --provider=private  # fetch repository aziendale

jvm configure-private <URL> <TOKEN>  # crea configurazione privata
jvm show-config
jvm reset-config
```

---

## 📒 VADEMECUM — Modifiche, Build e Distribuzione

### 🔧 Quando modifichi il codice Go

1. ✅ **Modifica i file** in `cmd/`, `providers/`, `utils/`
2. 🧪 **Testa la CLI** con:
    ```bash
    go run main.go remote-list
    ```
3. **Generazione Installer**
    ```bash
    GOOS=windows GOARCH=amd64 go build -o distribution/jvm.exe main.go
    ```

# 📦 Build automatizzato

Usa build.bat per:

-   ✅ Compilare jvm.exe in distribution/

-   🔐 Firmare con jvm-dev-cert.pfx (autofirmato)

-   📦 Compilare installer .exe via Inno Setup

Esegui dalla root del progetto:

```bat
build.bat
```

🧠 Priorità selezione versioni
✅ Versione LTS

📈 Versione più usata

🆕 Patch più recente

📎 Requisiti
Go 1.20+

Inno Setup Compiler

Windows SDK con signtool.exe

Certificato autofirmato .pfx (facoltativo)

🖋️ Creato da Marco Antonio Russo — powered by JVM CLI 💎

### NOTE

semplificare anche l’interfaccia Entry, tipo pubblico riutilizzabile per tutti i provider. Così lo standard diventa universale 🧩💼

non imposta variabile globale jvm
aggiungi repository privata nell'installer
