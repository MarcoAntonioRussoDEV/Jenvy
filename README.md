# Java Version Manager (JVM)

**Una soluzione professionale per la gestione centralizzata delle distribuzioni OpenJDK**

Java Version Manager è un'applicazione a riga di comando progettata per semplificare l'installazione, la gestione e il passaggio tra diverse versioni di OpenJDK su sistemi Windows. Il tool supporta i principali provider pubblici (Adoptium, Azul Zulu, BellSoft Liberica) e repository privati aziendali.

---

## Funzionalità Principali

### Gestione Multi-Provider

-   **Provider Pubblici**: Integrazione nativa con Adoptium (Eclipse Temurin), Azul Zulu e BellSoft Liberica
-   **Repository Privati**: Supporto completo per distribuzioni JDK aziendali personalizzate
-   **Configurazione Flessibile**: Gestione tramite file di configurazione locale o variabili d'ambiente

### Operazioni Core

-   **Esplorazione Remota**: Ricerca e visualizzazione delle versioni JDK disponibili con filtri avanzati
-   **Download Intelligente**: Scaricamento automatico con rilevamento dell'architettura di sistema
-   **Estrazione Automatica**: Opzione di estrazione immediata al completamento del download
-   **Gestione Locale**: Visualizzazione e amministrazione delle versioni JDK installate
-   **Switching Automatico**: Cambio di versione JDK attiva con elevazione automatica dei privilegi

### Caratteristiche Avanzate

-   **Autocompletamento**: Supporto nativo per Bash, PowerShell e Command Prompt
-   **Filtri Intelligenti**: Selezione automatica basata su criteri LTS, versioni maggiori e patch più recenti
-   **Gestione PATH**: Strumenti integrati per la riparazione e manutenzione delle variabili di sistema
-   **Rimozione Sicura**: Eliminazione controllea con conferme di sicurezza per operazioni distruttive

---

## Installazione

### Distribuzione Windows

1. Scaricare il file `jvm-installer.exe` dalla sezione releases
2. Eseguire l'installer con privilegi di amministratore
3. Il comando `jvm` sarà disponibile globalmente in tutti i terminali

### Compilazione da Sorgenti

```bash
# Clonare il repository
git clone https://github.com/MarcoAntonioRussoDEV/JavaVersionManager.git
cd JavaVersionManager

# Compilazione per Windows
GOOS=windows GOARCH=amd64 go build -o jvm.exe main.go

# Build completo con installer (richiede Inno Setup)
./build.bat
```

---

## Guida all'Utilizzo

### Esplorazione delle Versioni Disponibili

```bash
# Visualizzazione versioni dal provider predefinito (Adoptium)
jvm remote-list

# Esplorazione di provider specifici
jvm remote-list --provider=azul
jvm remote-list --provider=liberica
jvm remote-list --provider=private

# Filtri avanzati
jvm remote-list --lts-only          # Solo versioni Long Term Support
jvm remote-list --major-only        # Solo versioni maggiori
jvm remote-list --latest            # Solo le versioni più recenti
jvm remote-list --all               # Tutte le versioni da tutti i provider
```

### Download e Installazione

```bash
# Download di una versione specifica
jvm download 21

# Il sistema richiederà automaticamente se estrarre l'archivio:
# [?] Do you want to extract the archive now? (Y/n):
# - Y/y/Enter: Estrazione automatica immediata
# - n/N: Solo download, estrazione manuale successiva

# Estrazione manuale di archivi già scaricati
jvm extract JDK-21.0.1+12
```

### Gestione delle Versioni Installate

```bash
# Visualizzazione versioni installate
jvm list

# Attivazione di una versione specifica (richiede privilegi admin)
jvm use 21

# Configurazione della versione predefinita
jvm init
```

### Amministrazione Repository Privati

```bash
# Configurazione repository aziendale
jvm configure-private https://repository.company.com/jdk YOUR_TOKEN

# Visualizzazione configurazione corrente
jvm config-show

# Reset configurazione
jvm config-reset
```

### Rimozione e Manutenzione

```bash
# Rimozione versione specifica
jvm remove 17

# Rimozione completa (con conferma di sicurezza)
jvm remove --all

# Riparazione variabili di sistema
jvm fix-path
```

---

## Configurazione Avanzata

### Repository Privati

Il sistema supporta due modalità di configurazione per repository privati:

#### File di Configurazione

Percorso: `%USERPROFILE%\.jvm\config.json`

```json
{
    "private": {
        "endpoint": "https://repository.company.com/api/jdk",
        "token": "your-auth-token"
    }
}
```

#### Variabili d'Ambiente

```bash
set JVM_PRIVATE_ENDPOINT=https://repository.company.com/api/jdk
set JVM_PRIVATE_TOKEN=your-auth-token
```

### Struttura API Repository Privati

Il sistema richiede che i repository privati espongano un endpoint REST che restituisca un array JSON con le versioni JDK disponibili. L'endpoint deve supportare autenticazione tramite header `Authorization: Bearer <token>`.

#### Specifica dell'Endpoint

**URL:** `GET {endpoint}/api/jdk` o endpoint configurato  
**Headers:** `Authorization: Bearer {token}`  
**Content-Type:** `application/json`

#### Formato Risposta JSON

```json
[
    {
        "version": "11.0.21",
        "download": "https://repository.company.com/private-jdk/openjdk-11.0.21.zip",
        "os": "windows",
        "arch": "x64",
        "lts": true
    },
    {
        "version": "17.0.15",
        "download": "https://repository.company.com/private-jdk/openjdk-17.0.15.zip",
        "os": "windows",
        "arch": "x64",
        "lts": true
    },
    {
        "version": "21.0.7",
        "download": "https://repository.company.com/private-jdk/openjdk-21.0.7.zip",
        "os": "windows",
        "arch": "x64",
        "lts": true
    },
    {
        "version": "22.0.2",
        "download": "https://repository.company.com/private-jdk/openjdk-22.0.2.zip",
        "os": "windows",
        "arch": "x64",
        "lts": false
    }
]
```

#### Campi Obbligatori

| Campo      | Tipo    | Descrizione                                   | Valori Accettati                                         |
| ---------- | ------- | --------------------------------------------- | -------------------------------------------------------- |
| `version`  | String  | Versione semantica del JDK                    | Formato: `major.minor.patch` o `major.minor.patch+build` |
| `download` | String  | URL diretto per il download dell'archivio JDK | URL HTTPS valido                                         |
| `os`       | String  | Sistema operativo target                      | `windows`, `linux`, `macos`                              |
| `arch`     | String  | Architettura CPU                              | `x64`, `x32`, `aarch64`                                  |
| `lts`      | Boolean | Indica se è una versione Long Term Support    | `true`, `false`                                          |

#### Esempio di Implementazione Server

```javascript
// Esempio endpoint Node.js/Express
app.get("/api/jdk", authenticateToken, (req, res) => {
    const jdkVersions = [
        {
            version: "11.0.21",
            download:
                "https://repository.company.com/private-jdk/openjdk-11.0.21.zip",
            os: "windows",
            arch: "x64",
            lts: true,
        },
        // ... altre versioni
    ];

    res.json(jdkVersions);
});

function authenticateToken(req, res, next) {
    const authHeader = req.headers["authorization"];
    const token = authHeader && authHeader.split(" ")[1];

    if (!token || !isValidToken(token)) {
        return res.sendStatus(401);
    }

    next();
}
```

### Autocompletamento

```bash
# Bash
jvm completion bash >> ~/.bashrc

# PowerShell
jvm completion powershell >> $PROFILE

# Command Prompt
jvm completion cmd
```

---

## Gestione Privilegi Windows

### Elevazione Automatica UAC

Il comando `jvm use` richiede automaticamente l'elevazione dei privilegi attraverso il dialogo UAC di Windows per:

-   Modificare la variabile di sistema `JAVA_HOME`
-   Aggiornare la variabile di sistema `PATH`
-   Garantire la persistenza delle modifiche per tutti gli utenti

**Flusso operativo:**

1. Esecuzione comando `jvm use <version>`
2. Richiesta automatica elevazione privilegi
3. Conferma utente tramite dialogo UAC
4. Applicazione modifiche con privilegi amministrativi
