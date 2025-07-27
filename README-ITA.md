[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# Jenvy - Developer Kit Manager

<img src="assets/icons/jenvy_white.svg" alt="logo" height="400" />

### Una soluzione professionale per la gestione centralizzata delle distribuzioni OpenJDK

---

Jenvy è un'applicazione a riga di comando progettata per semplificare l'installazione, la gestione e il passaggio tra diverse versioni di OpenJDK su sistemi Windows. Il tool supporta i principali provider pubblici (Adoptium, Azul Zulu, BellSoft Liberica) e repository privati aziendali.

> **⚠️ Importante:** Questo è un progetto open source personale e indipendente. Non sono affiliato con Oracle Corporation o con i suoi prodotti. Jenvy è un tool di gestione per distribuzioni OpenJDK di terze parti e non include, distribuisce o modifica alcun software Oracle.

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
-   **Rimozione Sicura**: Eliminazione controllata con conferme di sicurezza per operazioni distruttive

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


### Amministrazione Repository Privati
```

### Repository privati

```bash
# Configurazione repository privato
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

Il sistema richiede che i repository privati espongano un endpoint REST che restituisca un array JSON con le versioni JDK disponibili. L'endpoint può supportare autenticazione tramite header `Authorization: Bearer <token>`.

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

---

## 💖 Supporta il Progetto

Jenvy è un progetto open source sviluppato nel tempo libero. Se trovi utile questo tool e vuoi supportare il suo sviluppo, considera una donazione:

### 🎯 Opzioni di Donazione

-   **GitHub Sponsors**: [Sponsorizza su GitHub](https://github.com/sponsors/MarcoAntonioRussoDEV)
-   **Ko-fi**: [Supporta su Ko-fi](https://ko-fi.com/marcoantoniorussodev)
-   **PayPal**: [Dona via PayPal](https://paypal.me/Ocrama94)

### 🚀 Come vengono utilizzate le donazioni

Le donazioni aiutano a:

-   Mantenere il progetto attivo e aggiornato
-   Aggiungere nuove funzionalità richieste dalla community
-   Migliorare la documentazione e i test

### 🤝 Altri modi per contribuire

Anche se non puoi donare, puoi supportare il progetto:

-   ⭐ Metti una stella al repository su GitHub
-   🐛 Segnala bug e problemi
-   💡 Suggerisci nuove funzionalità
-   📖 Migliora la documentazione
-   🔧 Contribuisci con pull request

---

## 📄 Licenza

Questo progetto è rilasciato sotto licenza **MIT License**.

```
MIT License

Copyright (c) 2025 Marco Antonio Russo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### 🔒 Disclaimer e Responsabilità

-   **Nessuna affiliazione**: Questo progetto non è affiliato, approvato o sponsorizzato da Oracle Corporation
-   **Software di terze parti**: Jenvy gestisce distribuzioni OpenJDK fornite da provider terzi (Eclipse Adoptium, Azul, BellSoft)
-   **Utilizzo a proprio rischio**: Il software è fornito "as-is" senza garanzie di alcun tipo
-   **Responsabilità utente**: L'utente è responsabile del rispetto delle licenze dei JDK scaricati
-   **Marchi registrati**: Java e OpenJDK sono marchi registrati di Oracle Corporation

---

## 🤝 Contribuire

Contributi, segnalazioni di bug e richieste di funzionalità sono benvenuti!

### 📋 Come contribuire

1. Fai fork del repository
2. Crea un branch per la tua feature (`git checkout -b feature/AmazingFeature`)
3. Committa le tue modifiche (`git commit -m 'Add some AmazingFeature'`)
4. Pusha il branch (`git push origin feature/AmazingFeature`)
5. Apri una Pull Request

### 🐛 Segnalare Bug

Apri una [issue su GitHub](https://github.com/MarcoAntonioRussoDEV/Jenvy/issues) includendo:

-   Versione di Windows utilizzata
-   Versione di Jenvy (`jenvy --version`)
-   Descrizione dettagliata del problema
-   Log di errore (se disponibile)
-   Passi per riprodurre il bug

### 💡 Richiedere Funzionalità

Per nuove funzionalità, apri una [discussion su GitHub](https://github.com/MarcoAntonioRussoDEV/Jenvy/discussions) specificando:

-   Caso d'uso specifico
-   Comportamento desiderato
-   Eventuali alternative considerate

---

## 📞 Contatti

-   **GitHub**: [@MarcoAntonioRussoDEV](https://github.com/MarcoAntonioRussoDEV)
-   **Email**: marcoantoniorusso94@gmail.com

---

![signature](assets/images/SVG_GRADIENT_WHITE.svg)

---
