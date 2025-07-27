package adoptium

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AdoptiumResponse struct {
    Binaries []struct {
        OS      string `json:"os"`
        Arch    string `json:"architecture"`
        Package struct {
            Link string `json:"link"`
        } `json:"package"`
    } `json:"binaries"`

    VersionData struct {
        OpenJDKVersion string `json:"openjdk_version"`
    } `json:"version_data"`
}


func GetJDKList() ([]AdoptiumResponse, error) {
    url := "https://api.adoptium.net/v3/assets/feature_releases/21/ga?architecture=x64&os=windows&image_type=jdk"

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var data []AdoptiumResponse
    if err := json.Unmarshal(body, &data); err != nil {
        return nil, err
    }

    return data, nil
}
func GetAllJDKs() ([]AdoptiumResponse, error) {
    versions, err := GetAvailableVersions()
    if err != nil {
        return nil, err
    }

    var all []AdoptiumResponse
    for _, v := range versions {
        url := fmt.Sprintf("https://api.adoptium.net/v3/assets/feature_releases/%s/ga?architecture=x64&os=windows&image_type=jdk", v)
        resp, err := http.Get(url)
        if err != nil {
            continue
        }
        body, _ := io.ReadAll(resp.Body)
        resp.Body.Close()

        var data []AdoptiumResponse
        if err := json.Unmarshal(body, &data); err == nil {
            all = append(all, data...)
        }
    }
    return all, nil
}

