#!/bin/bash
# build.sh - Script di build automatizzato per Java Version Manager (bash version)

set -e  # Exit on any error

# Determina la directory root del progetto (dove si trova go.mod)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🔧 Java Version Manager - Build Script"
echo "======================================"

echo "► Running tests..."
# Esegui i test dalla directory root del progetto
if (cd "$PROJECT_ROOT" && go test ./test/); then
    echo "✅ Tutti i test sono passati"
else
    echo "❌ Test falliti! Build interrotto."
    echo "💡 Correggi i test prima di procedere con la build"
    exit 1
fi

echo ""
echo "► Building jvm.exe..."
if (cd "$PROJECT_ROOT" && go build -o build/dist/jvm.exe ./main.go); then
    echo "✅ jvm.exe compilato con successo"
else
    echo "❌ Build fallito: go build ha restituito errore $?"
    exit 1
fi

if [ ! -f "$PROJECT_ROOT/build/dist/jvm.exe" ]; then
    echo "❌ Build fallito: jvm.exe non trovato dopo la compilazione"
    exit 1
fi

# Step 2: Sign jvm.exe with certificate (optional on Windows)
SIGNTOOL="/c/Program Files (x86)/Windows Kits/10/bin/x64/signtool.exe"
if [ -f "$SIGNTOOL" ]; then
    echo "► Signing jvm.exe..."
    if "$SIGNTOOL" sign /f "$PROJECT_ROOT/build/installer/jvm-dev-cert.pfx" /p jvm-password /tr http://timestamp.digicert.com /td sha256 "$PROJECT_ROOT/build/dist/jvm.exe"; then
        echo "✅ jvm.exe firmato con successo"
    else
        echo "⚠️ Firma fallita ma continuo (errore $?)"
    fi
else
    echo "⚠️ SignTool non trovato. Salta firma di jvm.exe."
fi

echo ""
echo "⏱️ Attendo il rilascio dei file lock..."
echo "💡 Se hai VS Code aperto con file dalla cartella distribution, chiudilo ora!"
sleep 5

# Remove old installer if exists
if [ -f "$PROJECT_ROOT/build/dist/jvm-installer.exe" ]; then
    echo "🗑️ Rimuovo il vecchio installer..."
    if rm -f "$PROJECT_ROOT/build/dist/jvm-installer.exe"; then
        echo "✅ Vecchio installer rimosso"
    else
        echo "⚠️ Non posso rimuovere il vecchio installer (potrebbe essere in uso)"
    fi
fi

# Check if jvm.exe is in use
if lsof "$PROJECT_ROOT/build/dist/jvm.exe" 2>/dev/null; then
    echo "❌ ERRORE: jvm.exe è ancora in uso da un altro processo"
    echo "💡 Chiudi tutti i terminali, VS Code e altri processi che potrebbero usare il file"
    echo "💡 Oppure riavvia il sistema e riprova"
    exit 1
fi

# Step 2.5: Convert README.md to README.txt for distribution
echo "► Converting README.md to build/dist/README.txt..."
if command -v sed >/dev/null 2>&1; then
    # Convert Markdown to plain text using sed
    sed -E '
        # Remove horizontal lines (---)
        /^-{3,}$/d
        
        # Remove headers (convert ### Title to Title)
        s/^#{1,6}\s*(.*)$/\1/g
        
        # Remove bold (**text** -> text)
        s/\*\*([^*]+)\*\*/\1/g
        
        # Remove italic (*text* -> text) 
        s/\*([^*]+)\*/\1/g
        
        # Remove inline code (`code` -> code)
        s/`([^`]+)`/\1/g
        
        # Convert links [text](url) to just text
        s/\[([^\]]+)\]\([^\)]+\)/\1/g
        
        # Convert bullet points
        s/^[\s]*-\s*/* /g
        
        # Remove code blocks (everything between ``` or ````)
        /^```/,/^```/d
        /^````/,/^````/d
        
    ' "$PROJECT_ROOT/README.md" | sed ':a;N;$!ba;s/\n\n\n*/\n\n/g' > "$PROJECT_ROOT/build/dist/README.txt"
    echo "✅ README.txt generato con successo"
else
    # Fallback: simple copy if sed is not available
    echo "⚠️ sed non disponibile, copio direttamente README.md come README.txt"
    cp "$PROJECT_ROOT/README.md" "$PROJECT_ROOT/build/dist/README.txt"
fi

# Step 3: Compile installer with Inno Setup
INNO="/c/Program Files (x86)/Inno Setup 6/ISCC.exe"
if [ -f "$INNO" ]; then
    echo "► Compiling installer..."
    if (cd "$PROJECT_ROOT/build/installer" && "$INNO" setup.iss); then
        echo "✅ Installer compilato con successo"
    else
        echo "❌ Compilazione installer fallita (errore $?)"
        echo "💡 Possibili soluzioni:"
        echo "   • Chiudi VS Code completamente"
        echo "   • Chiudi tutti i terminali che potrebbero avere handle sui file"
        echo "   • Riavvia il sistema se il problema persiste"
        exit 1
    fi
else
    echo "❌ ISCC.exe non trovato. Installa Inno Setup Compiler."
    exit 1
fi

echo ""
echo "🎉 Build completo! Controlla la cartella build/dist/"
echo "📦 File generati:"
echo "   - jvm.exe (eseguibile principale)"
echo "   - jvm-installer.exe (installer)"
echo ""
echo "🚀 Esempi di test:"
echo "   ./build/dist/jvm.exe remote-list"
echo "   ./build/dist/jvm-installer.exe /CONFIGURE_PRIVATE=0"
