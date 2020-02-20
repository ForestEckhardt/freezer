package github_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/ForestEckhardt/freezer/github"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testReleaseService(t *testing.T, context spec.G, it spec.S) {
	var (
		service github.ReleaseService
		api     *httptest.Server
	)

	context("Get", func() {
		it.Before(func() {
			api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				dump, _ := httputil.DumpRequest(req, true)

				if req.Header.Get("Authorization") != "token some-github-token" {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				switch req.URL.Path {
				case "/repos/some-org/some-repo/releases/latest":
					w.Write([]byte(`{
  "tag_name": "some-tag",
  "assets": [
    {
      "browser_download_url": "some-browser-download-url"
    }
  ],
  "tarball_url": "some-tarball-url"
					}`))
				case "/repos/some-org/missing-repo/releases/latest":
					w.WriteHeader(http.StatusNotFound)
				case "/repos/some-org/malformed-repo/releases/latest":
					w.Write([]byte("%%%"))
				default:
					Fail(fmt.Sprintf("unexpected request:\n%s", dump))
				}
			}))

			service = github.NewReleaseService(github.Config{
				Endpoint: api.URL,
				Token:    "some-github-token",
			})
		})

		it("fetches the latest release", func() {
			release, err := service.Get("some-org", "some-repo")
			Expect(err).ToNot(HaveOccurred())
			Expect(release).To(Equal(github.Release{
				TagName: "some-tag",
				Assets: []github.ReleaseAsset{
					{
						BrowserDownloadURL: "some-browser-download-url",
					},
				},
				TarballURL: "some-tarball-url",
			}))
		})

		context("failure cases", func() {
			context("when the request url is malformed", func() {
				it.Before(func() {
					service = github.NewReleaseService(github.Config{
						Endpoint: "%%%",
					})
				})

				it("returns an error", func() {
					_, err := service.Get("some-org", "some-repo")
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape \"%%%\"")))
				})
			})

			context("when the response status is not 200 OK", func() {
				it("returns an error", func() {
					_, err := service.Get("some-org", "missing-repo")
					Expect(err).To(MatchError("unexpected response status: 404 Not Found"))
				})
			})

			context("when the response JSON is malformed", func() {
				it("returns an error", func() {
					_, err := service.Get("some-org", "malformed-repo")
					Expect(err).To(MatchError(ContainSubstring("invalid character '%'")))
				})
			})
		})
	})
}
