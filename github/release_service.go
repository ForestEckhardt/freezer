package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ReleaseService struct {
	config Config
}

type ReleaseAsset struct {
	URL  string `json:"url"`
	Name string `json:"name"`
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

	if rs.config.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", rs.config.Token))
	}

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

func (rs ReleaseService) GetReleaseAsset(asset ReleaseAsset) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", asset.URL, nil)
	if err != nil {
		return nil, err
	}

	if rs.config.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", rs.config.Token))
	}

	req.Header.Add("Accept", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return resp.Body, nil
}

func (rs ReleaseService) GetReleaseTarball(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if rs.config.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", rs.config.Token))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return resp.Body, nil
}
