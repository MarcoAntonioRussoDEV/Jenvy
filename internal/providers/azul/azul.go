package azul

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type AzulPackage struct {
    Name        string   `json:"name"`
    JavaVersion []int    `json:"java_version"`
    DownloadURL string   `json:"download_url"`
    Latest      bool     `json:"latest"`
}


func GetAzulJDKs() ([]AzulPackage, error) {
    url := "https://api.azul.com/metadata/v1/zulu/packages?java_package_type=jdk&os=windows&arch=x86_64&availability_types=CA&release_status=ga&page_size=100"

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var data []AzulPackage
    if err := json.Unmarshal(body, &data); err != nil {
        return nil, err
    }

    return data, nil
}

func FormatVersion(v []int) string {
    var parts []string
    for _, n := range v {
        parts = append(parts, fmt.Sprintf("%d", n))
    }
    return strings.Join(parts, ".")
}

func InferPlatform(name string) (string, string) {
    name = strings.ToLower(name)
    switch {
    case strings.Contains(name, "win_x64"):
        return "windows", "x64"
    case strings.Contains(name, "linux_x64"):
        return "linux", "x64"
    case strings.Contains(name, "macos_x64"):
        return "macos", "x64"
    default:
        return "?", "?"
    }
}

