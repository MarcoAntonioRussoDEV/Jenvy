package adoptium

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Available struct {
    AvailableReleases []int `json:"available_releases"`
}

func GetAvailableVersions() ([]string, error) {
    resp, err := http.Get("https://api.adoptium.net/v3/info/available_releases")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var info Available
    if err := json.Unmarshal(body, &info); err != nil {
        return nil, err
    }

    var versions []string
    for _, v := range info.AvailableReleases {
        versions = append(versions, fmt.Sprintf("%d", v))
    }
    return versions, nil
}
