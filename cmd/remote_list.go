package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"jvm/providers/adoptium"
	"jvm/providers/azul"
	"jvm/providers/liberica"
	"jvm/providers/private"
	"jvm/utils"
)

func RemoteList(defaultProvider string) {
	// 👇 Usa il valore ricevuto da main.go come default
	provider := flag.String("provider", defaultProvider, "provider: adoptium | azul | liberica | private")
	all := flag.Bool("all", false, "Mostra versioni da tutti i provider")
	majorOnly := flag.Bool("major-only", false, "Mostra solo le major release")
	latestOnly := flag.Bool("latest", false, "Mostra solo la versione più recente")
	jdkFilter := flag.Int("jdk", 0, "Filtra solo una versione JDK (es. --jdk=17)")
	ltsOnly := flag.Bool("lts-only", false, "Mostra solo versioni LTS")
	flag.CommandLine.Parse(os.Args[2:])

	defaultMode := !*all && !*majorOnly && !*latestOnly && *jdkFilter == 0 && !*ltsOnly

	if *all && defaultMode {
		fmt.Println("🧠 Selezione filtrata con versione raccomandata per ciascun provider\n")
		printRecommendedAdoptium()
		printRecommendedAzul()
		printRecommendedLiberica()
		return
	}

	if defaultMode {
		fmt.Println("🧠 Selezione filtrata con versione raccomandata per provider:", *provider, "\n")
		switch strings.ToLower(*provider) {
		case "adoptium":
			printRecommendedAdoptium()
		case "azul":
			printRecommendedAzul()
		case "liberica":
			printRecommendedLiberica()
		case "private":
			printRecommendedPrivate()
		default:
			fmt.Printf("❌ Provider '%s' non valido. Usa --provider=adoptium | azul | liberica | private\n", *provider)
		}
		return
	}

	if *all {
		fmt.Println("🧭 Recupero JDK da tutti i provider...\n")
		printAdoptium(*majorOnly, *latestOnly, *jdkFilter, *ltsOnly)
		printAzul(*majorOnly, *latestOnly, *jdkFilter, *ltsOnly)
		printLiberica(*majorOnly, *latestOnly, *jdkFilter, *ltsOnly)
		return
	}

	switch strings.ToLower(*provider) {
	case "adoptium":
		printAdoptium(*majorOnly, *latestOnly, *jdkFilter, *ltsOnly)
	case "azul":
		printAzul(*majorOnly, *latestOnly, *jdkFilter, *ltsOnly)
	case "liberica":
		printLiberica(*majorOnly, *latestOnly, *jdkFilter, *ltsOnly)
	case "private":
		printPrivate(*majorOnly, *latestOnly, *jdkFilter, *ltsOnly)
	default:
		fmt.Printf("❌ Provider '%s' non valido. Usa --provider=adoptium | azul | liberica | private\n", *provider)
	}
}

func printRecommendedAdoptium() {
	fmt.Println("🔄 Recupero dati da Adoptium...")
	list, err := adoptium.GetAllJDKs()
	if err != nil {
		fmt.Println("❌ Errore Adoptium:", err)
		return
	}
	fmt.Println("🟢 Adoptium")
	recommended := adoptium.GetRecommendedJDKs(list)
	var data [][]string
	for _, j := range recommended {
		data = append(data, []string{j.Version, j.OS, j.Arch, j.LTS, j.Link})
	}
	utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
}

func printRecommendedAzul() {
	fmt.Println("🔄 Recupero dati da Azul...")
	list, err := azul.GetAzulJDKs()
	if err != nil {
		fmt.Println("❌ Errore Azul:", err)
		return
	}
	fmt.Println("🔵 Azul")
	recommended := azul.GetRecommendedJDKs(list)
	var data [][]string
	for _, j := range recommended {
		data = append(data, []string{j.Version, j.OS, j.Arch, j.LTS, j.DownloadURL})
	}
	utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
}

func printRecommendedLiberica() {
	fmt.Println("🔄 Recupero dati da Liberica...")
	list, err := liberica.GetLibericaJDKs()
	if err != nil {
		fmt.Println("❌ Errore Liberica:", err)
		return
	}
	fmt.Println("🟣 Liberica")
	recommended := liberica.GetRecommendedJDKs(list)
	var data [][]string
	for _, j := range recommended {
		data = append(data, []string{j.Version, j.OS, j.Arch, j.LTS, j.DownloadURL})
	}
	utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
}

func printRecommendedPrivate() {
	fmt.Println("🔄 Recupero dati dal repository privato...")
	list, err := private.GetPrivateJDKs()
	if err != nil {
		fmt.Println("❌ Errore Private:", err)
		return
	}
	fmt.Println("🔒 Private")
	recommended := private.GetRecommendedJDKs(list)
	var data [][]string
	for _, j := range recommended {
		data = append(data, []string{j.Version, j.OS, j.Arch, j.LTS, j.DownloadURL})
	}
	utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
}

func printAdoptium(majorOnly, latestOnly bool, jdkFilter int, ltsOnly bool) {
	list, err := adoptium.GetAllJDKs()
	if err != nil {
		fmt.Println("Errore nel recupero da Adoptium:", err)
		return
	}

	var data [][]string
	if latestOnly {
		latest := adoptium.GetLatestAdoptium(list, majorOnly)
		for _, j := range latest {
			if jdkFilter != 0 && j.Major != jdkFilter {
				continue
			}
			if ltsOnly && j.LTS != utils.IfBool(true) {
				continue
			}
			data = append(data, []string{j.Version, j.OS, j.Arch, j.LTS, j.Link})
		}
		utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
		return
	}

	for _, j := range list {
		version := j.VersionData.OpenJDKVersion
		isLTS := strings.Contains(strings.ToLower(version), "lts")
		major, _, _ := adoptium.ParseVersion(version)

		if majorOnly && !strings.Contains(version, ".0.0") {
			continue
		}
		if jdkFilter != 0 && major != jdkFilter {
			continue
		}
		if ltsOnly && !isLTS {
			continue
		}

		for _, b := range j.Binaries {
			data = append(data, []string{
				version,
				b.OS,
				b.Arch,
				utils.IfBool(isLTS),
				b.Package.Link,
			})
		}
	}
	utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
}

func printAzul(majorOnly, latestOnly bool, jdkFilter int, ltsOnly bool) {
	list, err := azul.GetAzulJDKs()
	if err != nil {
		fmt.Println("Errore nel recupero da Azul:", err)
		return
	}

	var data [][]string
	if latestOnly {
		latest := azul.GetLatestAzul(list, majorOnly)
		for _, j := range latest {
			if jdkFilter != 0 && j.Major != jdkFilter {
				continue
			}
			if ltsOnly && j.LTS != utils.IfBool(true) {
				continue
			}
			data = append(data, []string{j.Version, j.OS, j.Arch, j.LTS, j.DownloadURL})
		}
		utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
		return
	}

	for _, j := range list {
		if majorOnly && len(j.JavaVersion) > 1 && j.JavaVersion[1] != 0 {
			continue
		}
		if !strings.HasSuffix(j.DownloadURL, ".zip") {
			continue
		}

		major := j.JavaVersion[0]
		if jdkFilter != 0 && major != jdkFilter {
			continue
		}
		isLTS := major == 11 || major == 17 || major == 21 || major == 24
		if ltsOnly && !isLTS {
			continue
		}

		version := utils.FormatVersion(j.JavaVersion)
		os, arch := utils.InferPlatform(j.Name)

		data = append(data, []string{
			version,
			os,
			arch,
			utils.IfBool(isLTS),
			j.DownloadURL,
		})
	}
	utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
}

func printLiberica(majorOnly, latestOnly bool, jdkFilter int, ltsOnly bool) {
	list, err := liberica.GetLibericaJDKs()
	if err != nil {
		fmt.Println("Errore nel recupero da Liberica:", err)
		return
	}

	var data [][]string
	if latestOnly {
		latest := liberica.GetLatestLiberica(list, majorOnly)
		for _, j := range latest {
			if jdkFilter != 0 && j.Major != jdkFilter {
				continue
			}
			if ltsOnly && j.LTS != utils.IfBool(true) {
				continue
			}
			data = append(data, []string{j.Version, j.OS, j.Arch, j.LTS, j.DownloadURL})
		}
		utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
		return
	}

	for _, j := range list {
		major, _, _ := liberica.ParseLibericaVersion(j.Version)
		if majorOnly && !(strings.Contains(j.Version, ".0.0") || strings.Contains(j.Version, "+")) {
			continue
		}
		if jdkFilter != 0 && major != jdkFilter {
			continue
		}
		isLTS := strings.HasPrefix(j.Version, "17.") || strings.HasPrefix(j.Version, "21.") || strings.Contains(strings.ToLower(j.Version), "lts")
		if ltsOnly && !isLTS {
			continue
		}

		data = append(data, []string{
			j.Version,
			j.OS,
			j.Arch,
			utils.IfBool(isLTS),
			j.DownloadURL,
		})
	}
	utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
}

func printPrivate(majorOnly, latestOnly bool, jdkFilter int, ltsOnly bool) {
	list, err := private.GetPrivateJDKs()
	if err != nil {
		fmt.Println("❌ Errore Private:", err)
		return
	}

	// 🔄 Converti in []RecommendedEntry e poi in []utils.Entry
	converted := private.ConvertToRecommended(list)
	var all []utils.Entry
	for _, r := range converted {
		all = append(all, r)
	}

	var data [][]string

	if latestOnly {
		latest := utils.LatestForEachMajor(all, majorOnly)
		for _, entry := range latest {
			if jdkFilter != 0 && entry.MajorValue() != jdkFilter {
				continue
			}
			if ltsOnly && !entry.LtsValue() {
				continue
			}
			data = append(data, []string{
				entry.(private.RecommendedEntry).Version,
				entry.(private.RecommendedEntry).OS,
				entry.(private.RecommendedEntry).Arch,
				entry.(private.RecommendedEntry).LTS,
				entry.(private.RecommendedEntry).DownloadURL,
			})
		}
		fmt.Println("🔒 Private")
		utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
		return
	}

	for _, entry := range all {
		if jdkFilter != 0 && entry.MajorValue() != jdkFilter {
			continue
		}
		if ltsOnly && !entry.LtsValue() {
			continue
		}
		if majorOnly && entry.MinorValue() != 0 {
			continue
		}

		data = append(data, []string{
			entry.(private.RecommendedEntry).Version,
			entry.(private.RecommendedEntry).OS,
			entry.(private.RecommendedEntry).Arch,
			entry.(private.RecommendedEntry).LTS,
			entry.(private.RecommendedEntry).DownloadURL,
		})
	}

	fmt.Println("🔒 Private")
	utils.PrintTable(data, []string{"Versione", "OS", "Arch", "LTS", "Download"})
}
