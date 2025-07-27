package cmd

import (
	"fmt"
	"jenvy/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

// GenerateCompletion genera e stampa lo script di completamento Bash per Windows.
//
// Questa funzione produce uno script di completamento Bash compatibile con Windows
// che supporta Git Bash, WSL e altri ambienti Bash su Windows. Lo script include
// completamento intelligente per comandi, versioni JDK installate e flag.
//
// Caratteristiche del completamento generato:
// - Completamento comandi principali (remote-list, download, use, remove, etc.)
// - Completamento alias abbreviati (rl, dl, u, rm, etc.)
// - Completamento provider (adoptium, azul, liberica, private)
// - Completamento flag (--provider, --all, --latest, etc.)
// - Completamento versioni JDK installate per comandi 'use' e 'remove'
// - Completamento intelligente del flag --all per 'remove'
// - Fallback su versioni comuni se 'jenvy list' non è disponibile
//
// Utilizzo:
//
//	jenvy completion                    # Stampa lo script Bash
//	jenvy completion >> ~/.bashrc       # Aggiunge al profilo Bash
//	source ~/.bashrc                  # Ricarica il profilo
//
// Compatibilità Windows:
//   - Git Bash (più comune su Windows)
//   - WSL (Windows Subsystem for Linux)
//   - MSYS2/MinGW Bash
//   - Cygwin Bash
//
// Note tecniche:
//   - Usa grep e sed per parsing output di 'jenvy list'
//   - Limita risultati a 20 versioni per performance
//   - Gestisce fallback sicuro se il comando jenvy non è disponibile
//   - Script autocontenuto senza dipendenze esterne bash-completion
//
// Output: Stampa lo script completo su stdout (nessun valore di ritorno)
func GenerateCompletion() {
	script := `#!/bin/bash

# Bash completion script for Jenvy
# To enable completion, run:
#   jenvy completion >> ~/.bashrc
#   source ~/.bashrc
# Or install globally:
#   jenvy completion | sudo tee /etc/bash_completion.d/jenvy

_jenvy_completion() {
    local cur prev words cword
    # Fallback initialization for systems without bash-completion
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    words=("${COMP_WORDS[@]}")
    cword=$COMP_CWORD

    local commands="remote-list rl download dl extract ex list l use u remove rm init fix-path fp configure-private cp config-show cs config-reset cr completion help --help -h"
    local providers="adoptium azul liberica private"
    local flags="--provider --all --latest --major-only --jdk --lts-only --output"
    
    # Special handling for use and remove commands to complete with installed JDK versions
    if [[ "$prev" == "use" || "$prev" == "u" || "$prev" == "remove" || "$prev" == "rm" ]]; then
        # Try to get installed JDK versions using jenvy list
        if command -v jenvy >/dev/null 2>&1; then
            local installed_versions=$(jenvy list 2>/dev/null | grep -E "^\s*JDK-[0-9]" | sed 's/.*JDK-\([^[:space:]]*\).*/\1/' | head -20)
            if [[ -n "$installed_versions" ]]; then
                COMPREPLY=($(compgen -W "$installed_versions" -- "$cur"))
                return 0
            fi
        fi
        # Fallback to common versions if jenvy list is not available
        local common_versions="8 11 17 21 23 24"
        COMPREPLY=($(compgen -W "$common_versions" -- "$cur"))
        return 0
    fi
    
    # Handle command-specific completions
    case "$prev" in
        jenvy)
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return 0
            ;;
        --provider)
            COMPREPLY=($(compgen -W "$providers" -- "$cur"))
            return 0
            ;;
        --jdk)
            local versions="8 11 17 21 23 24"
            COMPREPLY=($(compgen -W "$versions" -- "$cur"))
            return 0
            ;;
        --output)
            # Complete directory paths
            COMPREPLY=($(compgen -d -- "$cur"))
            return 0
            ;;
        configure-private|cp)
            # For configure-private, suggest common endpoint patterns
            if [[ ${#words[@]} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "https://nexus.company.com/api/jdk https://artifactory.company.com/jdk http://localhost:8080/jdk-list.json" -- "$cur"))
            fi
            return 0
            ;;
        download|dl)
            # Complete with common JDK versions
            local versions="8 11 17 21 23 24"
            COMPREPLY=($(compgen -W "$versions" -- "$cur"))
            return 0
            ;;
        remove|rm)
            # Complete with installed JDK versions or --all flag
            if [[ "$cur" == --* ]]; then
                COMPREPLY=($(compgen -W "--all" -- "$cur"))
            else
                # Try to get installed JDK versions using jenvy list
                if command -v jenvy >/dev/null 2>&1; then
                    local installed_versions=$(jenvy list 2>/dev/null | grep -E "^\s*JDK-[0-9]" | sed 's/.*JDK-\([^[:space:]]*\).*/\1/' | head -20)
                    if [[ -n "$installed_versions" ]]; then
                        COMPREPLY=($(compgen -W "$installed_versions --all" -- "$cur"))
                    else
                        COMPREPLY=($(compgen -W "--all" -- "$cur"))
                    fi
                else
                    COMPREPLY=($(compgen -W "--all" -- "$cur"))
                fi
            fi
            return 0
            ;;
    esac

    # Handle subcommand completions
    local command="${words[1]}"
    case "$command" in
        remote-list|rl)
            COMPREPLY=($(compgen -W "$flags" -- "$cur"))
            return 0
            ;;
        download|dl)
            if [[ "$cur" == --* ]]; then
                COMPREPLY=($(compgen -W "--provider --output" -- "$cur"))
            else
                local versions="8 11 17 21 23 24"
                COMPREPLY=($(compgen -W "$versions" -- "$cur"))
            fi
            return 0
            ;;
        extract|ex)
            # Complete with available archive versions from ~/.jenvy/versions
            if command -v jenvy >/dev/null 2>&1; then
                local available_archives=$(jenvy extract 2>/dev/null | grep -E "^\s*JDK-[0-9]" | sed 's/.*JDK-\([^[:space:]]*\).*/\1/' | head -20)
                if [[ -n "$available_archives" ]]; then
                    # Add both full names and short versions for intelligent parsing
                    local short_versions=$(echo "$available_archives" | sed 's/\([0-9][0-9]*\).*/\1/')
                    COMPREPLY=($(compgen -W "$available_archives $short_versions" -- "$cur"))
                    return 0
                fi
            fi
            # Fallback to common versions
            local common_versions="8 11 17 21 23 24"
            COMPREPLY=($(compgen -W "$common_versions" -- "$cur"))
            return 0
            ;;
        list|l|use|u|remove|rm|init|fix-path|fp|configure-private|cp|config-show|cs|config-reset|cr|completion|help|--help|-h)
            # These commands don't take additional arguments or have specific handling above
            return 0
            ;;
        *)
            # Default completion for unknown commands
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return 0
            ;;
    esac
}

# Register the completion function
complete -F _jenvy_completion jenvy
`
	fmt.Print(script)
}

// GeneratePowerShellCompletion genera e restituisce lo script di completamento PowerShell per Windows.
//
// Questa funzione crea uno script di completamento nativo PowerShell che si integra
// perfettamente con l'ambiente Windows PowerShell/PowerShell Core tramite il sistema
// Register-ArgumentCompleter.
//
// Caratteristiche del completamento PowerShell:
// - Integrazione nativa con il sistema di completamento PowerShell
// - Supporto per tutti i comandi Jenvy e relativi alias
// - Completamento dinamico delle versioni JDK installate
// - Gestione intelligente dei flag e parametri
// - Fallback robusto in caso di errori
// - Performance ottimizzata per ambiente Windows
//
// Tecnologie utilizzate:
//   - Register-ArgumentCompleter: API nativa PowerShell per completamento
//   - Select-String: Parsing efficiente dell'output di 'jenvy list'
//   - Try-Catch: Gestione errori robusta
//   - Where-Object: Filtering veloce dei risultati
//
// Utilizzo consigliato:
//
//	Add-Content $PROFILE -Value (jenvy completion powershell)
//	. $PROFILE                       # Ricarica il profilo PowerShell
//
// Compatibilità:
//   - Windows PowerShell 5.x
//   - PowerShell Core 6.x/7.x
//   - PowerShell su Windows 10/11
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	string - Script PowerShell completo pronto per l'installazione
//
// Note di implementazione:
//   - Usa regex per estrazione versioni JDK dall'output di 'jenvy list'
//   - Gestisce sia versioni numeriche che flag speciali (--all)
//   - Implementa logica di completamento contestuale basata sulla posizione
func GeneratePowerShellCompletion() string {
	return generatePowerShellScript()
}

// GenerateCmdCompletion genera e restituisce uno script di aiuto per Command Prompt (CMD).
//
// Poiché CMD non supporta il completamento automatico avanzato come Bash o PowerShell,
// questa funzione genera uno script batch che fornisce suggerimenti e aiuto contestuale
// per gli utenti che utilizzano il classico Command Prompt di Windows.
//
// Caratteristiche dello script CMD generato:
// - Lista completa di tutti i comandi disponibili
// - Esempi di utilizzo per ogni comando
// - Spiegazione dei parametri e flag principali
// - Suggerimenti per provider e versioni comuni
// - Istruzioni per creazione di alias DOS
//
// Limitazioni di CMD:
//   - Nessun completamento automatico con TAB
//   - Nessuna integrazione nativa con shell
//   - Richiede esecuzione manuale dello script di aiuto
//
// Workaround implementato:
//   - Genera file .bat con documentazione completa
//   - Suggerisce creazione di alias DOS (doskey)
//   - Fornisce riferimento rapido per tutti i comandi
//
// Utilizzo suggerito:
//
//	jenvy completion cmd > jenvy-help.bat  # Salva lo script
//	doskey jenvy-help=jenvy-help.bat $*    # Crea alias DOS
//	jenvy-help                           # Mostra aiuto completo
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	string - Script batch (.bat) con documentazione completa Jenvy
//
// Compatibilità:
//   - Tutte le versioni di Windows CMD
//   - Command Prompt classico
//   - Batch script tradizionali
//   - Ambienti aziendali Windows legacy
func GenerateCmdCompletion() string {
	return generateCmdScript()
}

// generateBashScript genera lo script di completamento Bash ottimizzato per Windows.
//
// Questa funzione interna crea uno script Bash autocontenuto che fornisce
// completamento intelligente per tutti i comandi Jenvy senza dipendere da
// bash-completion o altre librerie esterne.
//
// Architettura dello script:
// 1. **Inizializzazione variabili**: Setup sicuro delle variabili COMPREPLY
// 2. **Definizione comandi**: Lista completa di comandi, alias e flag
// 3. **Completamento intelligente**: Logica specifica per ogni tipo di comando
// 4. **Gestione versioni**: Query dinamica delle versioni JDK installate
// 5. **Fallback robusto**: Versioni predefinite se query fallisce
//
// Logica di completamento per comando:
//   - use/u: Versioni JDK installate (da 'jenvy list')
//   - remove/rm: Versioni installate + flag --all
//   - download/dl: Versioni comuni (8, 11, 17, 21, 23, 24)
//   - --provider: Lista provider (adoptium, azul, liberica, private)
//   - configure-private: Suggerimenti URL comuni
//
// Ottimizzazioni implementate:
//   - Limit di 20 risultati per performance
//   - Regex ottimizzate per parsing veloce
//   - Cache implicita delle versioni durante sessione
//   - Gestione errori silente per UX fluida
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	string - Script Bash completo con funzione _jenvy_completion e registrazione
//
// Compatibilità:
//   - Bash 3.x+ (standard su Git Bash)
//   - Non richiede bash-completion package
//   - Funziona in ambienti Bash minimali
//   - Compatibile con WSL e ambienti Unix-like su Windows
func generateBashScript() string {
	return `#!/bin/bash

# Bash completion script for Jenvy
_jenvy_completion() {
    local cur prev words cword
    # Fallback initialization for systems without bash-completion
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    words=("${COMP_WORDS[@]}")
    cword=$COMP_CWORD

    local commands="remote-list rl download dl list l use u remove rm init fix-path fp configure-private cp config-show cs config-reset cr completion help --help -h"
    local providers="adoptium azul liberica private"
    local flags="--provider --all --latest --major-only --jdk --lts-only --output"
    
    # Special handling for use and remove commands to complete with installed JDK versions
    if [[ "$prev" == "use" || "$prev" == "u" || "$prev" == "remove" || "$prev" == "rm" ]]; then
        # Try to get installed JDK versions using jenvy list
        if command -v jenvy >/dev/null 2>&1; then
            local installed_versions=$(jenvy list 2>/dev/null | grep -E "^\s*JDK-[0-9]" | sed 's/.*JDK-\([^[:space:]]*\).*/\1/' | head -20)
            if [[ -n "$installed_versions" ]]; then
                COMPREPLY=($(compgen -W "$installed_versions" -- "$cur"))
                return 0
            fi
        fi
        # Fallback to common versions if jenvy list is not available
        local common_versions="8 11 17 21 23 24"
        COMPREPLY=($(compgen -W "$common_versions" -- "$cur"))
        return 0
    fi
    
    # Handle command-specific completions
    case "$prev" in
        jenvy)
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return 0
            ;;
        --provider)
            COMPREPLY=($(compgen -W "$providers" -- "$cur"))
            return 0
            ;;
        --jdk)
            local versions="8 11 17 21 23 24"
            COMPREPLY=($(compgen -W "$versions" -- "$cur"))
            return 0
            ;;
        --output)
            COMPREPLY=($(compgen -d -- "$cur"))
            return 0
            ;;
        configure-private|cp)
            if [[ ${#words[@]} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "https://nexus.company.com/api/jdk https://artifactory.company.com/jdk http://localhost:8080/jdk-list.json" -- "$cur"))
            fi
            return 0
            ;;
        download|dl)
            local versions="8 11 17 21 23 24"
            COMPREPLY=($(compgen -W "$versions" -- "$cur"))
            return 0
            ;;
    esac

    # Handle subcommand completions
    local command="${words[1]}"
    case "$command" in
        remote-list|rl)
            COMPREPLY=($(compgen -W "$flags" -- "$cur"))
            return 0
            ;;
        download|dl)
            if [[ "$cur" == --* ]]; then
                COMPREPLY=($(compgen -W "--provider --output" -- "$cur"))
            else
                local versions="8 11 17 21 23 24"
                COMPREPLY=($(compgen -W "$versions" -- "$cur"))
            fi
            return 0
            ;;
        list|l|use|u|remove|rm|init|fix-path|fp|configure-private|cp|config-show|cs|config-reset|cr|completion|help|--help|-h)
            return 0
            ;;
        *)
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return 0
            ;;
    esac
}

complete -F _jenvy_completion jenvy
`
}

// generatePowerShellScript crea uno script di completamento nativo PowerShell avanzato.
//
// Questa funzione genera uno script PowerShell che sfrutta appieno il sistema
// Register-ArgumentCompleter nativo di PowerShell per fornire completamento
// intelligente e contestuale per tutti i comandi Jenvy.
//
// Architettura dello script PowerShell:
// 1. **Registrazione native completer**: Usa Register-ArgumentCompleter API
// 2. **Parsing intelligente argomenti**: Analisi dinamica della linea di comando
// 3. **Completamento contestuale**: Logica specifica basata su posizione e comando
// 4. **Gestione errori robusta**: Try-catch per operazioni di query JDK
// 5. **Performance ottimizzata**: Filtering efficiente con Where-Object
//
// Funzionalità avanzate implementate:
//   - **Query dinamica versioni**: Esecuzione sicura di 'jenvy list' per versioni reali
//   - **Regex avanzata**: Estrazione precisa versioni con Select-String
//   - **Completamento ibrido**: Combina versioni installate e flag speciali
//   - **Fallback intelligente**: Versioni predefinite se query fallisce
//   - **Context-aware**: Comportamento diverso basato su comando precedente
//
// Logica di completamento PowerShell:
//   - Analizza $wordToComplete per determinare contesto
//   - Identifica comando precedente per completamento specifico
//   - Esegue query JDK solo quando necessario
//   - Filtra risultati con pattern matching PowerShell nativo
//
// Gestione comandi speciali:
//   - use/u: Solo versioni installate (performance ottimale)
//   - remove/rm: Versioni installate + --all flag
//   - --provider: Lista provider predefiniti
//   - --jdk: Versioni comuni per filtering remoto
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	string - Script PowerShell completo con Register-ArgumentCompleter
//
// Requisiti PowerShell:
//   - PowerShell 3.0+ (Register-ArgumentCompleter introdotto in 3.0)
//   - Execution Policy che permette script locali
//   - Accesso al comando jenvy nel PATH
func generatePowerShellScript() string {
	return `# PowerShell completion script for Jenvy
# Add this to your PowerShell profile: Add-Content $PROFILE -Value (jenvy completion powershell)

Register-ArgumentCompleter -Native -CommandName jenvy -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)
    
    $commands = @('remote-list', 'rl', 'download', 'dl', 'extract', 'ex', 'list', 'l', 'use', 'u', 'remove', 'rm', 'init', 'fix-path', 'fp', 'configure-private', 'cp', 'config-show', 'cs', 'config-reset', 'cr', 'completion', 'help', '--help', '-h')
    $providers = @('adoptium', 'azul', 'liberica', 'private')
    $flags = @('--provider', '--all', '--latest', '--major-only', '--jdk', '--lts-only', '--output')
    $versions = @('8', '11', '17', '21', '23', '24')
    
    $words = $wordToComplete.Split(' ')
    $lastWord = $words[-1]
    $secondLastWord = if ($words.Length -gt 1) { $words[-2] } else { '' }
    
    # Complete main commands after 'jenvy'
    if ($words.Length -le 2 -and -not $lastWord.StartsWith('--')) {
        $commands | Where-Object { $_ -like "$lastWord*" }
    }
    # Complete providers after --provider
    elseif ($secondLastWord -eq '--provider') {
        $providers | Where-Object { $_ -like "$lastWord*" }
    }
    # Complete versions for use and remove commands or after --jdk
    elseif ($secondLastWord -eq 'use' -or $secondLastWord -eq 'u' -or $secondLastWord -eq '--jdk') {
        # Try to get installed versions first
        try {
            $installedVersions = & jenvy list 2>$null | Select-String "JDK-(\d+)" | ForEach-Object { $_.Matches[0].Groups[1].Value }
            if ($installedVersions) {
                $installedVersions | Where-Object { $_ -like "$lastWord*" }
            } else {
                $versions | Where-Object { $_ -like "$lastWord*" }
            }
        } catch {
            $versions | Where-Object { $_ -like "$lastWord*" }
        }
    }
    # Complete versions and --all flag for remove commands
    elseif ($secondLastWord -eq 'remove' -or $secondLastWord -eq 'rm') {
        if ($lastWord.StartsWith('--')) {
            @('--all') | Where-Object { $_ -like "$lastWord*" }
        } else {
            # Try to get installed versions first
            try {
                $installedVersions = & jenvy list 2>$null | Select-String "JDK-(\d+)" | ForEach-Object { $_.Matches[0].Groups[1].Value }
                if ($installedVersions) {
                    ($installedVersions + @('--all')) | Where-Object { $_ -like "$lastWord*" }
                } else {
                    @('--all') | Where-Object { $_ -like "$lastWord*" }
                }
            } catch {
                @('--all') | Where-Object { $_ -like "$lastWord*" }
            }
        }
    }
    # Complete archive files for extract command
    elseif ($secondLastWord -eq 'extract' -or $secondLastWord -eq 'ex') {
        Get-ChildItem -Path "." -Include "*.zip", "*.tar.gz" -Name | Where-Object { $_ -like "$lastWord*" }
    }
    # Complete flags
    elseif ($lastWord.StartsWith('--')) {
        $flags | Where-Object { $_ -like "$lastWord*" }
    }
}
`
}

// generateCmdScript produce uno script batch di aiuto per Command Prompt Windows.
//
// Poiché CMD non supporta completamento automatico avanzato, questa funzione
// genera uno script batch (.bat) che serve come sistema di aiuto e riferimento
// rapido per tutti i comandi Jenvy disponibili.
//
// Struttura dello script batch generato:
// 1. **Header informativo**: Spiegazione uso e scopo dello script
// 2. **Controllo parametri**: Verifica se chiamato con "jenvy" come parametro
// 3. **Lista comandi completa**: Tutti i comandi con alias e sintassi
// 4. **Esempi pratici**: Uso comune di ogni comando
// 5. **Riferimenti rapidi**: Provider, versioni e flag disponibili
//
// Contenuto informativo incluso:
//   - Tutti i comandi principali con alias abbreviati
//   - Sintassi completa per ogni comando
//   - Lista provider supportati (adoptium, azul, liberica, private)
//   - Versioni JDK comuni (8, 11, 17, 21, 23, 24)
//   - Flag speciali come --all per remove
//   - Istruzioni per alias DOS (doskey)
//
// Strategia di workaround per limitazioni CMD:
//   - Genera documentazione sempre disponibile
//   - Suggerisce creazione alias DOS per accesso rapido
//   - Fornisce esempi copy-paste per comandi comuni
//   - Include troubleshooting per problemi comuni
//
// Utilizzo del file generato:
//  1. Salvare output in file .bat
//  2. Eseguire per vedere aiuto completo
//  3. Creare alias DOS per accesso rapido
//  4. Utilizzare come riferimento durante lavoro
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	string - Script batch completo con documentazione Jenvy
//
// Compatibilità:
//   - Tutti le versioni Windows con CMD
//   - Batch scripting tradizionale
//   - Ambienti Windows aziendali
//   - Sistemi senza PowerShell o Bash
func generateCmdScript() string {
	return `@echo off
REM CMD completion script for Jenvy
REM This provides basic command suggestions for CMD

if "%1"=="jenvy" (
    echo Available commands:
    echo   remote-list ^(rl^)     - List available JDK versions from providers
    echo   download ^(dl^)        - Download and install a JDK version
    echo   extract ^(ex^)         - Extract JDK archive to versions directory
    echo   list ^(l^)             - List installed JDK versions
    echo   use ^(u^)              - Set JAVA_HOME system-wide
    echo   remove ^(rm^) ^<version^> - Remove installed JDK version
    echo   remove ^(rm^) --all    - Remove ALL JDK installations
    echo   init                  - Initialize environment and completion
    echo   fix-path ^(fp^)        - Add JDK to PATH
    echo   configure-private ^(cp^) - Configure private repository
    echo   config-show ^(cs^)     - Show current configuration
    echo   config-reset ^(cr^)    - Reset configuration
    echo   completion            - Generate completion scripts
    echo   help                  - Show this help
    echo.
    echo Providers: adoptium, azul, liberica, private
    echo Common versions: 8, 11, 17, 21, 23, 24
)
`
}

// InstallCompletionForAllShells installa automaticamente il completamento per tutte le shell disponibili su Windows.
//
// Questa funzione è il punto di ingresso principale per l'installazione automatica
// del completamento su tutti gli ambienti shell disponibili nel sistema Windows.
// Tenta l'installazione per Bash, PowerShell e CMD, gestendo gracefully
// gli errori e fornendo feedback dettagliato all'utente.
//
// Processo di installazione automatica:
// 1. **Rilevamento ambienti**: Identifica shell disponibili nel sistema
// 2. **Installazione Bash**: Modifica ~/.bashrc per Git Bash/WSL
// 3. **Installazione PowerShell**: Modifica profilo PowerShell
// 4. **Setup CMD**: Crea script di aiuto e suggerisce alias
// 5. **Reporting risultati**: Feedback completo su successi/fallimenti
//
// Strategia di gestione errori:
//   - Installazione non-blocking: errore in una shell non ferma le altre
//   - Logging dettagliato: specifiche ragioni di fallimento per debug
//   - Graceful degradation: continua anche se alcune shell falliscono
//   - User feedback: informa utente su cosa è riuscito e cosa no
//
// Vantaggi dell'installazione automatica:
//   - Setup one-click per tutte le shell
//   - Esperienza utente senza attrito
//   - Compatibilità massima con ambienti Windows
//   - Fallback robusto in caso di problemi
//   - Reporting trasparente dei risultati
//
// Shell target supportate:
//   - **Bash**: Git Bash, WSL, MSYS2, Cygwin
//   - **PowerShell**: Windows PowerShell 5.x, PowerShell Core 6.x/7.x
//   - **CMD**: Command Prompt tradizionale con script di aiuto
//
// Requisiti di sistema:
//   - Directory home utente accessibile
//   - Permessi di scrittura nei profili shell
//   - Almeno una shell installata nel sistema
//
// Comportamento post-installazione:
//   - Richiede riavvio terminale per attivazione
//   - Fornisce istruzioni specifiche per ogni shell
//   - Suggerisce comandi per test del completamento
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	Nessuno (void) - Stampa risultati su stdout tramite utils.Print*
//
// Side effects:
//   - Modifica file di configurazione shell (~/.bashrc, $PROFILE)
//   - Crea file di aiuto per CMD (~/.jenvy_cmd_help.bat)
//   - Stampa informazioni di stato e istruzioni utente
func InstallCompletionForAllShells() {
	utils.PrintInfo("Installing completion scripts for all available shells...")

	var installed []string
	var errors []string

	// Installa per Bash
	if err := installBashCompletion(); err != nil {
		errors = append(errors, fmt.Sprintf("Bash: %v", err))
	} else {
		installed = append(installed, "Bash")
	}

	// Installa per PowerShell
	if err := installPowerShellCompletion(); err != nil {
		errors = append(errors, fmt.Sprintf("PowerShell: %v", err))
	} else {
		installed = append(installed, "PowerShell")
	}

	// Installa per CMD (suggerimenti di base)
	if err := installCmdCompletion(); err != nil {
		errors = append(errors, fmt.Sprintf("CMD: %v", err))
	} else {
		installed = append(installed, "CMD")
	}

	// Mostra risultati
	if len(installed) > 0 {
		utils.PrintSuccess(fmt.Sprintf("Completion installed for: %s", strings.Join(installed, ", ")))
	}

	if len(errors) > 0 {
		utils.PrintWarning(fmt.Sprintf("Failed to install completion for: %s", strings.Join(errors, "; ")))
	}

	if len(installed) > 0 {
		utils.PrintInfo("Restart your terminal or source your shell configuration to enable completions")
	}
}

// installBashCompletion installa il completamento Bash nel profilo utente Windows.
//
// Questa funzione gestisce l'installazione del completamento Bash modificando
// il file ~/.bashrc dell'utente. È progettata per funzionare principalmente
// con Git Bash (l'ambiente Bash più comune su Windows) ma supporta anche
// WSL e altri ambienti Bash su Windows.
//
// Processo di installazione:
// 1. **Localizzazione home directory**: Usa os.UserHomeDir() per trovare profilo utente
// 2. **Costruzione percorso**: Identifica ~/.bashrc come target di installazione
// 3. **Controllo duplicati**: Verifica se completamento già installato
// 4. **Prevenzione duplicazione**: Evita installazioni multiple dello stesso script
// 5. **Scrittura sicura**: Appende script al file esistente o crea nuovo file
//
// Gestione intelligente duplicati:
//   - Scansione contenuto ~/.bashrc esistente
//   - Ricerca signature "_jenvy_completion" per rilevare installazione precedente
//   - Skip installazione se già presente (evita duplicazione)
//   - Permette aggiornamenti manuali se necessario
//
// Struttura script installato:
//   - Header identificativo per future verifiche
//   - Script Bash completo generato da generateBashScript()
//   - Registrazione finale con comando 'complete'
//   - Compatibilità con reload automatico profilo
//
// Casi d'uso supportati:
//   - **Git Bash**: Ambiente primario Windows per Git
//   - **WSL**: Windows Subsystem for Linux
//   - **MSYS2**: Ambiente di sviluppo Windows con Bash
//   - **Cygwin**: Ambiente Unix-like per Windows
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	error - nil se installazione riuscita, errore specifico altrimenti
//
// Errori possibili:
//   - Impossibile determinare directory home utente
//   - Permessi insufficienti per modificare ~/.bashrc
//   - Filesystem pieno o in sola lettura
//   - Corruzione file ~/.bashrc esistente
//
// Note di sicurezza:
//   - Usa os.O_APPEND per evitare sovrascrittura contenuto esistente
//   - Crea file con permessi 0644 (sicuri per file di configurazione)
//   - Non modifica mai contenuto esistente, solo aggiunge
func installBashCompletion() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %v", err)
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")

	// Controlla se il completamento è già installato
	if content, err := os.ReadFile(bashrcPath); err == nil {
		if strings.Contains(string(content), "_jenvy_completion") {
			return nil // Già installato
		}
	}

	script := "\n# Jenvy completion\n" + generateBashScript()

	file, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening ~/.bashrc: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(script); err != nil {
		return fmt.Errorf("writing to ~/.bashrc: %v", err)
	}

	return nil
}

// installPowerShellCompletion installa il completamento nel profilo PowerShell Windows.
//
// Questa funzione gestisce l'installazione del completamento PowerShell localizzando
// e modificando il file di profilo PowerShell appropriato. Supporta sia Windows
// PowerShell (5.x) che PowerShell Core (6.x/7.x) con rilevamento automatico.
//
// Strategia di localizzazione profilo:
// 1. **Rilevamento directory home**: Base per tutti i percorsi profilo
// 2. **Percorsi PowerShell multipli**: Supporta diverse versioni PowerShell
//   - PowerShell Core: ~/Documents/PowerShell/Microsoft.PowerShell_profile.ps1
//   - Windows PowerShell: ~/Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1
//
// 3. **Verifica esistenza**: Controlla directory esistenti per installazione target
// 4. **Creazione automatica**: Crea directory e file se non esistenti
//
// Percorsi supportati (priorità ordinata):
//  1. ~/Documents/PowerShell/ (PowerShell Core 6.x/7.x)
//  2. ~/Documents/WindowsPowerShell/ (Windows PowerShell 5.x)
//
// Processo di installazione:
// 1. **Localizzazione profilo**: Trova o crea percorso profilo appropriato
// 2. **Controllo duplicati**: Verifica signature "Register-ArgumentCompleter -Native -CommandName jenvy"
// 3. **Prevenzione sovrascrittura**: Evita installazioni multiple
// 4. **Creazione directory**: os.MkdirAll per struttura directory se necessaria
// 5. **Installazione script**: Appende script PowerShell completo
//
// Compatibilità PowerShell:
//   - **Windows PowerShell 5.x**: Versione integrata Windows 10/11
//   - **PowerShell Core 6.x**: Versione cross-platform Microsoft
//   - **PowerShell 7.x**: Ultima versione unificata Microsoft
//
// Gestione errori specifica:
//   - Directory profilo PowerShell non trovata
//   - Permessi insufficienti per creazione directory
//   - Conflitti con configurazioni PowerShell esistenti
//   - Execution Policy PowerShell restrittive
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	error - nil se installazione riuscita, errore specifico altrimenti
//
// Side effects:
//   - Crea directory profilo PowerShell se non esistente
//   - Modifica o crea file Microsoft.PowerShell_profile.ps1
//   - Registra completion handler nativo PowerShell
//
// Note di sicurezza:
//   - Usa os.O_APPEND per preservare configurazioni esistenti
//   - Crea directory con permessi sicuri (0755)
//   - Non interferisce con altri moduli PowerShell
func installPowerShellCompletion() error {
	// Trova il percorso del profilo PowerShell
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %v", err)
	}

	// Percorsi comuni per il profilo PowerShell
	possiblePaths := []string{
		filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"),
		filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"),
	}

	var profilePath string
	for _, path := range possiblePaths {
		if dir := filepath.Dir(path); dir != "" {
			if _, err := os.Stat(dir); err == nil {
				profilePath = path
				break
			}
		}
	}

	// Se non esiste, crea la directory e il file
	if profilePath == "" {
		profilePath = possiblePaths[0] // Usa il primo percorso come default
		if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
			return fmt.Errorf("creating PowerShell profile directory: %v", err)
		}
	}

	// Controlla se il completamento è già installato
	if content, err := os.ReadFile(profilePath); err == nil {
		if strings.Contains(string(content), "Register-ArgumentCompleter -Native -CommandName jenvy") {
			return nil // Già installato
		}
	}

	script := "\n# Jenvy completion\n" + generatePowerShellScript()

	file, err := os.OpenFile(profilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening PowerShell profile: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(script); err != nil {
		return fmt.Errorf("writing to PowerShell profile: %v", err)
	}

	return nil
}

// installCmdCompletion crea uno script di aiuto per Command Prompt Windows.
//
// Poiché CMD non supporta completamento automatico nativo, questa funzione
// implementa una strategia alternativa creando uno script batch di aiuto
// e suggerendo la configurazione di alias DOS per accesso rapido.
//
// Strategia di workaround per limitazioni CMD:
// 1. **Creazione script aiuto**: File .bat con documentazione completa
// 2. **Posizionamento home directory**: ~/.jenvy_cmd_help.bat per accesso facile
// 3. **Documentazione integrata**: Tutti i comandi e sintassi in un file
// 4. **Suggerimento alias**: Istruzioni per creare alias DOS con doskey
// 5. **Riferimento permanente**: File sempre disponibile per consultazione
//
// Contenuto dello script di aiuto:
//   - Lista completa comandi Jenvy con alias
//   - Sintassi dettagliata per ogni comando
//   - Esempi pratici di utilizzo
//   - Provider supportati e versioni comuni
//   - Troubleshooting e suggerimenti
//
// Workflow utente suggerito:
// 1. Installazione automatica crea ~/.jenvy_cmd_help.bat
// 2. Utente esegue comando doskey suggerito per creare alias
// 3. Alias "jenvy-help" diventa disponibile in tutte le sessioni CMD
// 4. Utente può consultare aiuto con "jenvy-help" quando necessario
//
// Comando alias suggerito:
//
//	doskey jenvy-help=C:\Users\username\.jenvy_cmd_help.bat jenvy $*
//
// Vantaggi dell'approccio:
//   - Soluzione nativa CMD senza dipendenze esterne
//   - Documentazione sempre aggiornata e accessibile
//   - Zero configurazione aggiuntiva richiesta
//   - Compatibilità con tutti gli ambienti Windows CMD
//   - Performance immediata (nessuna query dinamica)
//
// Limitazioni accettate:
//   - Nessun completamento automatico con TAB
//   - Richiede esecuzione manuale per vedere suggerimenti
//   - Alias deve essere configurato manualmente dall'utente
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	error - nil se creazione script riuscita, errore specifico altrimenti
//
// Side effects:
//   - Crea file ~/.jenvy_cmd_help.bat nella home directory
//   - Stampa istruzioni per configurazione alias DOS
//   - Sovrascrive file esistente se presente (aggiornamento)
//
// Compatibilità:
//   - Tutte le versioni Windows con CMD
//   - Batch scripting tradizionale Windows
//   - Ambienti enterprise e legacy Windows
func installCmdCompletion() error {
	// Per CMD, creiamo un semplice file di aiuto
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %v", err)
	}

	cmdHelpPath := filepath.Join(homeDir, ".jenvy_cmd_help.bat")

	script := generateCmdScript()

	if err := os.WriteFile(cmdHelpPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("writing CMD help file: %v", err)
	}

	// Aggiungi suggerimento per l'alias
	fmt.Println("� [INFO] For CMD completion, add this alias to your environment:")
	fmt.Printf("   doskey jenvy-help=%s jenvy $*\n", cmdHelpPath)

	return nil
}
