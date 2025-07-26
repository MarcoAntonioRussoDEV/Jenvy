# Jenvy - Developer Kit Manager

**Una soluzione professionale per la gestione centralizzata delle distribuzioni OpenJDK**

Jenvy è un'applicazione a riga di comando progettata per semplificare l'installazione, la gestione e il passaggio tra diverse versioni di OpenJDK su sistemi Windows. Il tool supporta i principali provider pubblici (Adoptium, Azul Zulu, BellSoft Liberica) e repository privati aziendali.

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

1. Scaricare il file `jenvy-installer.exe` dalla sezione releases
2. Eseguire l'installer con privilegi di amministratore
3. Il comando `jenvy` sarà disponibile globalmente in tutti i terminali

### Compilazione da Sorgenti

```bash
# Clonare il repository
git clone https://github.com/MarcoAntonioRussoDEV/Jenvy.git
cd Jenvy

# Compilazione per Windows
GOOS=windows GOARCH=amd64 go build -o jenvy.exe main.go

# Build completo con installer (richiede Inno Setup)
./build.bat
```

---

## Guida all'Utilizzo

### Esplorazione delle Versioni Disponibili

```bash
# Visualizzazione versioni dal provider predefinito (Adoptium)
jenvy remote-list

# Esplorazione di provider specifici
jenvy remote-list --provider=azul
jenvy remote-list --provider=liberica
jenvy remote-list --provider=private

# Filtri avanzati
jenvy remote-list --lts-only          # Solo versioni Long Term Support
jenvy remote-list --major-only        # Solo versioni maggiori
jenvy remote-list --latest            # Solo le versioni più recenti
jenvy remote-list --all               # Tutte le versioni da tutti i provider
```

### Download e Installazione

```bash
# Download di una versione specifica
jenvy download 21

# Il sistema richiederà automaticamente se estrarre l'archivio:
# [?] Do you want to extract the archive now? (Y/n):
# - Y/y/Enter: Estrazione automatica immediata
# - n/N: Solo download, estrazione manuale successiva

# Estrazione manuale di archivi già scaricati
jenvy extract JDK-21.0.1+12
```

### Gestione delle Versioni Installate

```bash
# Visualizzazione versioni installate
jenvy list

# Attivazione di una versione specifica (richiede privilegi admin)
jenvy use 21

# Configurazione della versione predefinita
jenvy init
```

### Amministrazione Repository Privati

```bash
# Configurazione repository aziendale
jenvy configure-private https://repository.company.com/jdk YOUR_TOKEN

# Visualizzazione configurazione corrente
jenvy config-show

# Reset configurazione
jenvy config-reset
```

### Rimozione e Manutenzione

```bash
# Rimozione versione specifica
jenvy remove 17

# Rimozione completa (con conferma di sicurezza)
jenvy remove --all

# Riparazione variabili di sistema
jenvy fix-path
```

---

## Configurazione Avanzata

### Repository Privati

Il sistema supporta due modalità di configurazione per repository privati:

#### File di Configurazione

Percorso: `%USERPROFILE%\.jenvy\config.json`

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
set JENVY_PRIVATE_ENDPOINT=https://repository.company.com/api/jdk
set JENVY_PRIVATE_TOKEN=your-auth-token
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
jenvy completion bash >> ~/.bashrc

# PowerShell
jenvy completion powershell >> $PROFILE

# Command Prompt
jenvy completion cmd
```

---

## Gestione Privilegi Windows

### Elevazione Automatica UAC

Il comando `jenvy use` richiede automaticamente l'elevazione dei privilegi attraverso il dialogo UAC di Windows per:

-   Modificare la variabile di sistema `JAVA_HOME`
-   Aggiornare la variabile di sistema `PATH`
-   Garantire la persistenza delle modifiche per tutti gli utenti

**Flusso operativo:**

1. Esecuzione comando `jenvy use <version>`
2. Richiesta automatica elevazione privilegi
3. Conferma utente tramite dialogo UAC
4. Applicazione modifiche con privilegi amministrativi
