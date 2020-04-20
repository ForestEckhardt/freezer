package commands_test

import (
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/commands"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testStock(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		transport cargo.Transport
		packager  freezer.PackingTools

		// githubAPI *httptest.Server

		command commands.Stock
	)

	it.Before(func() {
		transport = cargo.NewTransport()
		packager = freezer.NewPackingTools()

		// 	githubAPI = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 		dump, _ := httputil.DumpRequest(req, true)
		// 		if req.Header.Get("Authorization") != "token some-github-token" {
		// 			w.WriteHeader(http.StatusForbidden)
		// 			return
		// 		}
		//
		// 		switch {
		// 		case req.URL.Path == "/repos/some-org/some-repo/releases/latest":
		// 			w.Write([]byte(`{
		// "tag_name": "some-tag",
		// "assets": [
		//   {
		//     "browser_download_url": "some-browser-download-url"
		//   }
		// ],
		// "tarball_url": "some-tarball-url"
		// 				}`))
		// 		default:
		// 			Fail(fmt.Sprintf("unexpected request:\n%s", dump))
		// 		}
		// 	}))

		command = commands.NewStock(transport, packager)
	})

	context("Execute", func() {
		it("runs the fetcher", func() {
			err := command.Execute([]string{
				"--org", "some-org",
				"--repo", "some-repo",
				"--github-token", "some-github-token",
			})

			Expect(err).NotTo(HaveOccurred())
		})
	})

	context("failure cases", func() {
		context("when given an unknown flag", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{"--unknown"})
				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined: -unknown")))
			})
		})

		context("when the --org flag is empty", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{
					"--repo", "some-repo",
					"--github-token", "some-token",
				})
				Expect(err).To(MatchError("missing required flag --org"))
			})
		})

		context("when the --repo flag is empty", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{
					"--org", "some-org",
					"--github-token", "some-token",
				})
				Expect(err).To(MatchError("missing required flag --repo"))
			})
		})

		context("when the --github-token flag is empty", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{
					"--org", "some-org",
					"--repo", "some-repo",
				})
				Expect(err).To(MatchError("missing required flag --github-token"))
			})
		})

	})
}
