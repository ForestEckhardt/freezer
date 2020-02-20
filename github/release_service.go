package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type ReleaseService struct {
	config Config
}

type ReleaseAsset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName    string         `json:"tag_name"`
	Assets     []ReleaseAsset `json:"assets"`
	TarballURL string         `json:"tarball_url"`
}

func NewReleaseService(config Config) ReleaseService {
	return ReleaseService{
		config: config,
	}
}

func (rs ReleaseService) Get(org, repo string) (Release, error) {
	uri, err := url.Parse(rs.config.Endpoint)
	if err != nil {
		return Release{}, err
	}

	uri.Path = fmt.Sprintf("/repos/%s/%s/releases/latest", org, repo)

	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", rs.config.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Release{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return Release{}, err
	}

	return release, nil
}
