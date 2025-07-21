package cmd

import (
	"fmt"
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
    _init_completion || return

    local commands="remote-list rl download dl list l configure-private cp config-show cs config-reset cr help --help -h"
    local providers="adoptium azul liberica private"
    local flags="--provider --all --latest --major-only --jdk --lts-only --output"
    
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
        list|l|configure-private|cp|config-show|cs|config-reset|cr|help|--help|-h)
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

// InstallCompletion installa automaticamente il completamento
func InstallCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Error getting home directory: %v\n", err)
		return
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")

	// Controlla se il completamento è già installato
	if content, err := os.ReadFile(bashrcPath); err == nil {
		if strings.Contains(string(content), "_jvm_completion") {
			fmt.Println("✅ JVM completion is already installed in ~/.bashrc")
			fmt.Println("💡 Run 'source ~/.bashrc' to reload completions")
			return
		}
	}

	// Genera e appende il completamento
	script := `

# Java Version Manager (jvm) completion
_jvm_completion() {
    local cur prev words cword
    _init_completion || return

    local commands="remote-list rl download dl list l configure-private cp config-show cs config-reset cr help --help -h"
    local providers="adoptium azul liberica private"
    local flags="--provider --all --latest --major-only --jdk --lts-only --output"
    
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
        list|l|configure-private|cp|config-show|cs|config-reset|cr|help|--help|-h)
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

	file, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("❌ Error opening ~/.bashrc: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(script); err != nil {
		fmt.Printf("❌ Error writing to ~/.bashrc: %v\n", err)
		return
	}

	fmt.Println("✅ JVM completion installed successfully!")
	fmt.Println("💡 Run 'source ~/.bashrc' to enable completions, or restart your terminal")
	fmt.Println("🔧 Now you can use Tab to autocomplete jvm commands and options")
}
