package liberica

import (
	"encoding/json"
	"io"
	"net/http"
)

type LibericaRelease struct {
    Version     string `json:"version"`
    DownloadURL string `json:"downloadUrl"`
    OS          string `json:"os"`
    Arch        string `json:"architecture"`
    Bitness     int    `json:"bitness"`
}

func GetLibericaJDKs() ([]LibericaRelease, error) {
    url := "https://api.bell-sw.com/v1/liberica/releases?bitness=64&os=windows&arch=x86&package-type=zip&bundle-type=jdk"

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var data []LibericaRelease
    if err := json.Unmarshal(body, &data); err != nil {
        return nil, err
    }

    return data, nil
}
