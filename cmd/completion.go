package cmd

import (
	"fmt"
	"jvm/utils"
	"os"
	"path/filepath"
	"strings"
)

// GenerateCompletion genera lo script di completamento bash
func GenerateCompletion() {
	script := `#!/bin/bash

# Bash completion script for Java Version Manager (jvm)
# To enable completion, run:
#   jvm completion >> ~/.bashrc
#   source ~/.bashrc
# Or install globally:
#   jvm completion | sudo tee /etc/bash_completion.d/jvm

_jvm_completion() {
    local cur prev words cword
    # Fallback initialization for systems without bash-completion
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    words=("${COMP_WORDS[@]}")
    cword=$COMP_CWORD

    local commands="remote-list rl download dl list l use u init fix-path fp configure-private cp config-show cs config-reset cr completion help --help -h"
    local providers="adoptium azul liberica private"
    local flags="--provider --all --latest --major-only --jdk --lts-only --output"
    
    # Special handling for use command to complete with installed JDK versions
    if [[ "$prev" == "use" || "$prev" == "u" ]]; then
        # Try to get installed JDK versions using jvm list
        if command -v jvm >/dev/null 2>&1; then
            local installed_versions=$(jvm list 2>/dev/null | grep -E "^\s*JDK-[0-9]" | sed 's/.*JDK-\([^[:space:]]*\).*/\1/' | head -20)
            if [[ -n "$installed_versions" ]]; then
                COMPREPLY=($(compgen -W "$installed_versions" -- "$cur"))
                return 0
            fi
        fi
        # Fallback to common versions if jvm list is not available
        local common_versions="8 11 17 21 23 24"
        COMPREPLY=($(compgen -W "$common_versions" -- "$cur"))
        return 0
    fi
    
    # Handle command-specific completions
    case "$prev" in
        jvm)
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
        list|l|use|u|init|fix-path|fp|configure-private|cp|config-show|cs|config-reset|cr|completion|help|--help|-h)
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
complete -F _jvm_completion jvm
`
	fmt.Print(script)
}

// GeneratePowerShellCompletion genera lo script di completamento PowerShell
func GeneratePowerShellCompletion() string {
	return generatePowerShellScript()
}

// GenerateCmdCompletion genera lo script di completamento CMD
func GenerateCmdCompletion() string {
	return generateCmdScript()
}

// generateBashScript genera lo script di completamento per Bash
func generateBashScript() string {
	return `#!/bin/bash

# Bash completion script for Java Version Manager (jvm)
_jvm_completion() {
    local cur prev words cword
    # Fallback initialization for systems without bash-completion
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    words=("${COMP_WORDS[@]}")
    cword=$COMP_CWORD

    local commands="remote-list rl download dl list l use u init fix-path fp configure-private cp config-show cs config-reset cr completion help --help -h"
    local providers="adoptium azul liberica private"
    local flags="--provider --all --latest --major-only --jdk --lts-only --output"
    
    # Special handling for use command to complete with installed JDK versions
    if [[ "$prev" == "use" || "$prev" == "u" ]]; then
        # Try to get installed JDK versions using jvm list
        if command -v jvm >/dev/null 2>&1; then
            local installed_versions=$(jvm list 2>/dev/null | grep -E "^\s*JDK-[0-9]" | sed 's/.*JDK-\([^[:space:]]*\).*/\1/' | head -20)
            if [[ -n "$installed_versions" ]]; then
                COMPREPLY=($(compgen -W "$installed_versions" -- "$cur"))
                return 0
            fi
        fi
        # Fallback to common versions if jvm list is not available
        local common_versions="8 11 17 21 23 24"
        COMPREPLY=($(compgen -W "$common_versions" -- "$cur"))
        return 0
    fi
    
    # Handle command-specific completions
    case "$prev" in
        jvm)
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
        list|l|use|u|use-user|init|fix-path|fp|configure-private|cp|config-show|cs|config-reset|cr|completion|help|--help|-h)
            return 0
            ;;
        *)
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return 0
            ;;
    esac
}

complete -F _jvm_completion jvm
`
}

// generatePowerShellScript genera lo script di completamento per PowerShell
func generatePowerShellScript() string {
	return `# PowerShell completion script for Java Version Manager (jvm)
# Add this to your PowerShell profile: Add-Content $PROFILE -Value (jvm completion powershell)

Register-ArgumentCompleter -Native -CommandName jvm -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)
    
    $commands = @('remote-list', 'rl', 'download', 'dl', 'list', 'l', 'use', 'u', 'init', 'fix-path', 'fp', 'configure-private', 'cp', 'config-show', 'cs', 'config-reset', 'cr', 'completion', 'help', '--help', '-h')
    $providers = @('adoptium', 'azul', 'liberica', 'private')
    $flags = @('--provider', '--all', '--latest', '--major-only', '--jdk', '--lts-only', '--output')
    $versions = @('8', '11', '17', '21', '23', '24')
    
    $words = $wordToComplete.Split(' ')
    $lastWord = $words[-1]
    $secondLastWord = if ($words.Length -gt 1) { $words[-2] } else { '' }
    
    # Complete main commands after 'jvm'
    if ($words.Length -le 2 -and -not $lastWord.StartsWith('--')) {
        $commands | Where-Object { $_ -like "$lastWord*" }
    }
    # Complete providers after --provider
    elseif ($secondLastWord -eq '--provider') {
        $providers | Where-Object { $_ -like "$lastWord*" }
    }
    # Complete versions for use commands or after --jdk
    elseif ($secondLastWord -eq 'use' -or $secondLastWord -eq 'u' -or $secondLastWord -eq '--jdk') {
        # Try to get installed versions first
        try {
            $installedVersions = & jvm list 2>$null | Select-String "JDK-(\d+)" | ForEach-Object { $_.Matches[0].Groups[1].Value }
            if ($installedVersions) {
                $installedVersions | Where-Object { $_ -like "$lastWord*" }
            } else {
                $versions | Where-Object { $_ -like "$lastWord*" }
            }
        } catch {
            $versions | Where-Object { $_ -like "$lastWord*" }
        }
    }
    # Complete flags
    elseif ($lastWord.StartsWith('--')) {
        $flags | Where-Object { $_ -like "$lastWord*" }
    }
}
`
}

// generateCmdScript genera lo script di completamento per CMD
func generateCmdScript() string {
	return `@echo off
REM CMD completion script for Java Version Manager (jvm)
REM This provides basic command suggestions for CMD

if "%1"=="jvm" (
    echo Available commands:
    echo   remote-list ^(rl^)     - List available JDK versions from providers
    echo   download ^(dl^)        - Download and install a JDK version
    echo   list ^(l^)             - List installed JDK versions
    echo   use ^(u^)              - Set JAVA_HOME system-wide
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

// InstallCompletionForAllShells installa il completamento per tutte le shell disponibili
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

// installBashCompletion installa il completamento per Bash
func installBashCompletion() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %v", err)
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")

	// Controlla se il completamento è già installato
	if content, err := os.ReadFile(bashrcPath); err == nil {
		if strings.Contains(string(content), "_jvm_completion") {
			return nil // Già installato
		}
	}

	script := "\n# Java Version Manager (jvm) completion\n" + generateBashScript()

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

// installPowerShellCompletion installa il completamento per PowerShell
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
		if strings.Contains(string(content), "Register-ArgumentCompleter -Native -CommandName jvm") {
			return nil // Già installato
		}
	}

	script := "\n# Java Version Manager (jvm) completion\n" + generatePowerShellScript()

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

// installCmdCompletion installa helper di completamento per CMD
func installCmdCompletion() error {
	// Per CMD, creiamo un semplice file di aiuto
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %v", err)
	}

	cmdHelpPath := filepath.Join(homeDir, ".jvm_cmd_help.bat")

	script := generateCmdScript()

	if err := os.WriteFile(cmdHelpPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("writing CMD help file: %v", err)
	}

	// Aggiungi suggerimento per l'alias
	fmt.Println("� [INFO] For CMD completion, add this alias to your environment:")
	fmt.Printf("   doskey jvm-help=%s jvm $*\n", cmdHelpPath)

	return nil
}
