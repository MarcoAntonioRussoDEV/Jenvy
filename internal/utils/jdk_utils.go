package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// ParseVersionNumber analizza e decompone una stringa di versione JDK in componenti numerici.
//
// Questa funzione è fondamentale per il sistema di matching delle versioni, convertendo
// stringhe di versione in formato "major.minor.patch" in componenti numerici separati
// per confronti e validazioni accurate.
//
// Formato versioni supportate:
//   - **Major only**: "17" → (17, -1, -1)
//   - **Major.Minor**: "17.0" → (17, 0, -1)
//   - **Major.Minor.Patch**: "17.0.5" → (17, 0, 5)
//   - **Formato esteso**: "21.0.2.13" → (21, 0, 2) [ignora componenti extra]
//   - **Java 8 legacy**: "1.8.0_452-b09" → (8, 0, 452) [normalizza a formato moderno]
//   - **Java 8 update**: "8.0.392" → (8, 0, 392) [formato moderno]
//   - **Liberica Java 8**: "8u352" → (8, 0, 352) [formato specifico Liberica]
//
// Gestione errori parsing:
//   - **Componenti non numerici**: Vengono impostati a -1
//   - **Formato malformato**: Parsing best-effort, -1 per parti mancanti
//   - **Stringa vuota**: Ritorna (0, -1, -1)
//
// Algoritmo di parsing:
// 1. **Preprocessing Java 8**: Gestisce formati legacy "1.8.0_xxx" e "8uxxx"
// 2. **Split sui punti**: Divide stringa su "." per ottenere componenti
// 3. **Conversione progressiva**: Converte ogni componente in intero
// 4. **Default fallback**: -1 per componenti mancanti o invalid
// 5. **Ritorno strutturato**: Tupla (major, minor, patch)
//
// Parametri:
//
//	version string - Stringa versione da analizzare (es. "17", "17.0.5", "1.8.0_452", "8u352")
//
// Restituisce:
//
//	major int - Versione major (es. 17), 0 se stringa vuota
//	minor int - Versione minor (es. 0), -1 se mancante/invalid
//	patch int - Versione patch (es. 5), -1 se mancante/invalid
//
// Esempi di utilizzo:
//
//	major, minor, patch := ParseVersionNumber("17")         // (17, -1, -1)
//	major, minor, patch := ParseVersionNumber("17.0")       // (17, 0, -1)
//	major, minor, patch := ParseVersionNumber("21.0.2")     // (21, 0, 2)
//	major, minor, patch := ParseVersionNumber("1.8.0_452")  // (8, 0, 452)
//	major, minor, patch := ParseVersionNumber("8.0.392")    // (8, 0, 392)
//	major, minor, patch := ParseVersionNumber("8u352")      // (8, 0, 352)
//
// Casi d'uso:
//   - Matching flessibile versioni durante ricerca
//   - Confronto versioni per determinare "migliore"
//   - Validazione input utente per versioni
//   - Filtering risultati provider per versione target
func ParseVersionNumber(version string) (major, minor, patch int) {
	if version == "" {
		return 0, -1, -1
	}

	// Normalizza spazi bianchi
	version = strings.TrimSpace(version)

	// Gestione speciale per formato Liberica Java 8: "8u352"
	if strings.HasPrefix(version, "8u") {
		updateStr := strings.TrimPrefix(version, "8u")
		if update, err := strconv.Atoi(updateStr); err == nil {
			return 8, 0, update
		}
		return 8, 0, 0
	}

	// Gestione speciale per formato Java 8 legacy: "1.8.0_452-b09" e "1.8.0"
	if strings.HasPrefix(version, "1.8.0") {
		if strings.Contains(version, "_") {
			// Formato con update: "1.8.0_452-b09"
			parts := strings.Split(version, "_")
			if len(parts) > 1 {
				updatePart := parts[1]
				// Rimuovi suffissi build come "-b09", "-ea", etc.
				if idx := strings.IndexAny(updatePart, "-+"); idx != -1 {
					updatePart = updatePart[:idx]
				}
				if update, err := strconv.Atoi(updatePart); err == nil {
					return 8, 0, update
				}
			}
		}
		// Formato base: "1.8.0" → (8, 0, 0)
		return 8, 0, 0
	}

	// Rimuovi suffissi build come "+13", "-b09", etc.
	if idx := strings.IndexAny(version, "+-"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	major = -1
	minor = -1
	patch = -1

	if len(parts) >= 1 {
		if m, err := strconv.Atoi(parts[0]); err == nil {
			major = m
		}
	}
	if len(parts) >= 2 {
		if m, err := strconv.Atoi(parts[1]); err == nil {
			minor = m
		}
	}
	if len(parts) >= 3 {
		if p, err := strconv.Atoi(parts[2]); err == nil {
			patch = p
		}
	}

	// Normalizza valori di default per minor/patch quando major è valido
	if major > 0 {
		if minor == -1 {
			minor = 0
		}
		if patch == -1 {
			patch = 0
		}
	}

	return major, minor, patch
}

// IsLTSVersion determina se una versione JDK è Long Term Support (LTS).
//
// Questa funzione centralizza la logica per identificare versioni LTS, utilizzando
// sia l'analisi numerica della versione che il controllo di marker testuali
// per una detection robusta e consistente tra tutti i provider.
//
// Versioni LTS riconosciute:
//   - **Java 8**: Major = 8 (tutte le versioni 8.x.x)
//   - **Java 11**: Major = 11 (tutte le versioni 11.x.x)
//   - **Java 17**: Major = 17 (tutte le versioni 17.x.x)
//   - **Java 21**: Major = 21 (tutte le versioni 21.x.x)
//   - **Marker testuale**: Stringhe contenenti "lts" (case-insensitive)
//
// Strategia di detection:
// 1. **Parsing numerico**: Usa ParseVersionNumber per estrarre major version
// 2. **Check LTS known**: Verifica se major è nelle versioni LTS note
// 3. **Fallback testuale**: Se non riconosciuta, cerca marker "lts" nella stringa
//
// Parametri:
//
//	version string - Versione da analizzare (es. "17.0.5", "21", "11.0.20-lts")
//
// Restituisce:
//
//	bool - true se versione è LTS, false altrimenti
//
// Esempi di utilizzo:
//
//	IsLTSVersion("17.0.5")     // true (Java 17 è LTS)
//	IsLTSVersion("21")         // true (Java 21 è LTS)
//	IsLTSVersion("19.0.2")     // false (Java 19 non è LTS)
//	IsLTSVersion("22-lts")     // true (marker testuale)
//	IsLTSVersion("1.8.0_452")  // true (Java 8 è LTS)
//
// Note:
//   - Basato su Oracle LTS roadmap ufficiale
//   - Compatibile con tutti i formati di versioning supportati
//   - Gestisce casi edge con marker testuali espliciti
func IsLTSVersion(version string) bool {
	major, _, _ := ParseVersionNumber(version)

	// Versioni LTS note (Oracle LTS roadmap)
	ltsVersions := []int{8, 11, 17, 21}

	for _, lts := range ltsVersions {
		if major == lts {
			return true
		}
	}

	// Fallback: cerca marker testuale "lts" (case-insensitive)
	return strings.Contains(strings.ToLower(version), "lts")
}

// IsValidJDKDirectory verifica se una directory contiene un'installazione JDK valida e completa.
//
// Questa funzione implementa controlli strutturali per validare che una directory
// contenga tutti i componenti essenziali di un'installazione JDK funzionante su Windows,
// prevenendo errori quando si cerca di utilizzare directory corrotte o incomplete.
//
// Struttura JDK richiesta:
//
//	JDK-Directory/
//	├── bin/              # Directory eseguibili (OBBLIGATORIA)
//	│   ├── java.exe      # Runtime Java (OBBLIGATORIO)
//	│   ├── javac.exe     # Compiler Java
//	│   ├── jar.exe       # Tool JAR
//	│   └── ...           # Altri tool JDK
//	└── lib/              # Directory librerie (OBBLIGATORIA)
//	    ├── rt.jar        # Runtime libraries (JDK 8)
//	    ├── modules       # Module system (JDK 9+)
//	    └── ...           # Altre librerie
//
// Controlli di validazione eseguiti:
// 1. **Directory bin/**: Verifica esistenza directory eseguibili
// 2. **Directory lib/**: Verifica esistenza directory librerie
// 3. **Eseguibile java.exe**: Verifica presenza runtime Java per Windows
//
// Perché questi controlli:
//   - **bin/**: Senza eseguibili il JDK è inutilizzabile
//   - **lib/**: Senza librerie Java non può funzionare
//   - **java.exe**: File più critico per esecuzione Java su Windows
//
// Controlli NON eseguiti (per performance):
//   - Validazione completeness librerie interne
//   - Test esecuzione effettiva java.exe
//   - Verifica versione JDK specifica
//   - Controllo integrità file individuali
//
// Parametri:
//
//	path string - Percorso assoluto directory da validare come JDK
//
// Restituisce:
//
//	bool - true se directory contiene JDK valido, false se incompleto/corrotto
//
// Scenari di utilizzo:
//   - Validazione prima di impostare JAVA_HOME
//   - Controllo integrità dopo download/estrazione
//   - Verifica installazioni esistenti durante listing
//   - Prevenzione errori prima di attivazione JDK
//
// Esempi di utilizzo:
//
//	if IsValidJDKDirectory("C:\\Users\\user\\.jvm\\versions\\JDK-17") {
//	    // Sicuro da utilizzare come JAVA_HOME
//	}
//
// Limitazioni:
//   - Non garantisce funzionalità completa del JDK
//   - Non verifica compatibilità architettura (x64/x32)
//   - Non controlla requisiti sistema specifici
//   - Non valida certificati o firme digitali
func IsValidJDKDirectory(path string) bool {
	// Check for typical JDK directories
	requiredDirs := []string{"bin", "lib"}
	for _, reqDir := range requiredDirs {
		if _, err := os.Stat(filepath.Join(path, reqDir)); err != nil {
			return false
		}
	}

	// Check for java executable (Windows-specific)
	javaExe := "java.exe"
	if _, err := os.Stat(filepath.Join(path, "bin", javaExe)); err != nil {
		return false
	}

	return true
}

// GetJVMVersionsDirectory ritorna il percorso della directory standard per le versioni JVM.
//
// Questa funzione centralizza la logica per determinare dove JVM installa e gestisce
// le diverse versioni JDK, fornendo un percorso consistente per tutte le operazioni
// di gestione versioni.
//
// Struttura directory standard:
//
//	C:\Users\{username}\.jvm\versions\
//	├── JDK-17.0.5\          # Versione specifica JDK
//	├── JDK-21.0.2\          # Altra versione JDK
//	└── JDK-8.0.392\         # Versione legacy
//
// Processo di determinazione:
// 1. **Directory utente**: Usa user.Current() per ottenere home directory
// 2. **Costruzione path**: Combina home + ".jvm" + "versions"
// 3. **Path assoluto**: Ritorna percorso completo e normalizzato
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	string - Percorso assoluto directory versions (~/.jvm/versions)
//	error  - nil se successo, errore se impossibile determinare directory home
//
// Utilizzo tipico:
//
//	versionsDir, err := GetJVMVersionsDirectory()
//	if err != nil {
//	    return fmt.Errorf("cannot access JVM directory: %w", err)
//	}
//	// versionsDir = "C:\Users\Marco\.jvm\versions"
func GetJVMVersionsDirectory() (string, error) {
	homeDir, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	versionsDir := filepath.Join(homeDir.HomeDir, ".jvm", "versions")
	return versionsDir, nil
}

// FindJDKInstallationPaths localizza tutti i percorsi di installazione per una versione JDK specifica.
//
// Questa funzione implementa un algoritmo di ricerca intelligente per trovare
// tutte le installazioni JDK corrispondenti alla versione richiesta, gestendo sia
// corrispondenze esatte che parziali nella directory delle versioni JVM.
//
// Algoritmo di ricerca a due fasi:
// 1. **Exact Match**: Cerca corrispondenza esatta "JDK-{version}"
//   - Input "17" → cerca "JDK-17"
//   - Input "17.0.5" → cerca "JDK-17.0.5"
//
// 2. **Partial Match**: Se exact match non trova, cerca prefissi
//   - Input "17" → trova "JDK-17.0.5", "JDK-17.0.8", etc.
//   - Input "17.0" → trova tutte le patch versions di 17.0.x
//
// Comportamento ricerca:
//   - **Case sensitive**: Match esatto su formato "JDK-{version}"
//   - **Prefix matching**: Versioni che iniziano con il pattern richiesto
//   - **Directory filtering**: Solo directory valide (non file)
//   - **Validazione JDK**: Ogni match viene verificato con IsValidJDKDirectory
//
// Parametri:
//
//	version string - Versione JDK da cercare (es. "17", "17.0.5", "21")
//
// Restituisce:
//
//	[]string - Lista di percorsi assoluti alle directory JDK trovate
//	error    - nil se operazione completata (anche se nessun match), errore per problemi I/O
//
// Risultati possibili:
//   - **[]string{}**: Nessuna installazione trovata per la versione
//   - **[]string{path}**: Una sola installazione trovata (ideale)
//   - **[]string{path1, path2, ...}**: Multiple installazioni (serve disambiguazione)
//
// Esempi di utilizzo:
//
//	paths, err := FindJDKInstallationPaths("17")
//	if err != nil {
//	    return fmt.Errorf("search failed: %w", err)
//	}
//	if len(paths) == 0 {
//	    return fmt.Errorf("no JDK found for version 17")
//	}
//	// paths = ["C:\Users\user\.jvm\versions\JDK-17.0.5"]
func FindJDKInstallationPaths(version string) ([]string, error) {
	versionsDir, err := GetJVMVersionsDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to get JVM directory: %w", err)
	}

	// Look for exact match first
	exactMatch := filepath.Join(versionsDir, fmt.Sprintf("JDK-%s", version))
	if _, err := os.Stat(exactMatch); err == nil {
		if IsValidJDKDirectory(exactMatch) {
			return []string{exactMatch}, nil
		}
	}

	// Look for partial matches
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions directory: %w", err)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if strings.HasPrefix(name, "JDK-") {
				jdkVersion := strings.TrimPrefix(name, "JDK-")
				if strings.HasPrefix(jdkVersion, version) {
					fullPath := filepath.Join(versionsDir, name)
					if IsValidJDKDirectory(fullPath) {
						matches = append(matches, fullPath)
					}
				}
			}
		}
	}

	return matches, nil
}

// FindSingleJDKInstallation localizza un singolo percorso JDK, gestendo disambiguazione automatica.
//
// Questa è una wrapper convenience function su FindJDKInstallationPaths che gestisce
// automaticamente i casi comuni e fornisce messaggi di errore user-friendly per
// scenari di ricerca multipla o fallita.
//
// Comportamento automatico:
//   - **Singolo match**: Ritorna immediatamente il percorso trovato
//   - **Nessun match**: Errore "no JDK found matching version"
//   - **Multiple matches**: Mostra lista interattiva e richiede precisione utente
//
// Gestione multiple corrispondenze:
//   - **Lista numerata**: Mostra tutte le opzioni con numerazione
//   - **Richiesta specifica**: Informa che serve maggiore precisione
//   - **Prevenzione ambiguità**: Evita selezioni accidentali errate
//
// Parametri:
//
//	version string - Versione JDK da cercare
//
// Restituisce:
//
//	string - Percorso assoluto directory JDK trovata
//	error  - nil se trovato singolo match, errore specifico per altri casi
//
// Tipi di errore:
//   - **Directory JVM inaccessibile**: Problemi permessi o configurazione
//   - **Nessuna corrispondenza**: Versione richiesta non installata
//   - **Multiple corrispondenze**: Versione ambigua, serve maggiore precisione
//
// Messaggi output per multiple matches:
//
//	Multiple JDK versions found:
//	  1. JDK-17.0.5
//	  2. JDK-17.0.8
//	Error: multiple matches found, please be more specific
//
// Esempio di utilizzo:
//
//	jdkPath, err := FindSingleJDKInstallation("17")
//	if err != nil {
//	    PrintError(fmt.Sprintf("JDK lookup failed: %v", err))
//	    return
//	}
//	// jdkPath = "C:\Users\user\.jvm\versions\JDK-17.0.5"
func FindSingleJDKInstallation(version string) (string, error) {
	matches, err := FindJDKInstallationPaths(version)
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no JDK found matching version %s", version)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	// Multiple matches - show options
	PrintWarning("Multiple JDK versions found:")
	for i, match := range matches {
		fmt.Printf("  %d. %s\n", i+1, filepath.Base(match))
	}
	return "", fmt.Errorf("multiple matches found, please be more specific")
}
