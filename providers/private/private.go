package private

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"jvm/utils"
)

type PrivateRelease struct {
	Version     string `json:"version"`
	DownloadURL string `json:"download"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	LTS         bool   `json:"lts"`
}

type RecommendedEntry struct {
	Version     string
	DownloadURL string
	OS          string
	Arch        string
	LTS         string
	Major       int
	Minor       int
	Patch       int
}

// ✔️ Fetch remoto da endpoint privato con token opzionale
func GetPrivateJDKs() ([]PrivateRelease, error) {
	cfg, err := utils.LoadConfig()
	if err != nil || cfg.PrivateEndpoint == "" {
		return nil, errors.New("⚠️ Private endpoint not configured. Check ~/.jvm/config.json")
	}

	endpoint := cfg.PrivateEndpoint
	token := cfg.PrivateToken

	if endpoint == "" {
		return nil, errors.New("⚠️ JVM_PRIVATE_ENDPOINT environment variable not set")
	}

	req, _ := http.NewRequest("GET", endpoint, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("❌ Server responded with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var list []PrivateRelease
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, fmt.Errorf("JSON parsing error: %v", err)
	}
	return list, nil
}

// ✔️ Raccomanda una versione per major: LTS > patch > minor
func GetRecommendedJDKs(list []PrivateRelease) []RecommendedEntry {
	group := make(map[int][]RecommendedEntry)

	for _, j := range list {
		major, minor, patch := utils.ParseGenericVersion(j.Version)

		entry := RecommendedEntry{
			Version:     j.Version,
			DownloadURL: j.DownloadURL,
			OS:          strings.ToLower(j.OS),
			Arch:        strings.ToLower(j.Arch),
			LTS:         utils.IfBool(j.LTS),
			Major:       major,
			Minor:       minor,
			Patch:       patch,
		}
		group[major] = append(group[major], entry)
	}

	var result []RecommendedEntry
	for _, entries := range group {
		utils.SortRecommended(entries)
		result = append(result, entries[0])
	}

	utils.SortRecommended(result)
	return result
}

// ✨ Per stampare in tabella con intestazioni coerenti
func (r RecommendedEntry) LtsValue() bool {
	return r.LTS == utils.IfBool(true)
}
func (r RecommendedEntry) PatchValue() int {
	return r.Patch
}
func (r RecommendedEntry) MinorValue() int {
	return r.Minor
}

func ConvertToRecommended(list []PrivateRelease) []RecommendedEntry {
	var result []RecommendedEntry
	for _, j := range list {
		major, minor, patch := utils.ParseGenericVersion(j.Version)

		entry := RecommendedEntry{
			Version:     j.Version,
			DownloadURL: j.DownloadURL,
			OS:          strings.ToLower(j.OS),
			Arch:        strings.ToLower(j.Arch),
			LTS:         utils.IfBool(j.LTS),
			Major:       major,
			Minor:       minor,
			Patch:       patch,
		}
		result = append(result, entry)
	}
	return result
}

func (r RecommendedEntry) MajorValue() int { return r.Major }
