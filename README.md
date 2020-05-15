# Freezer: A Library to Help Keep Your Buildpacks on Ice

## Usage
```go
func TestIntegration(t *testing.T) {
	fetcher := freezer.NewFetcher()
	Expect(fetcher.Open()).To(Succeed())
	defer fetcher.Close()

	localBuildpack, err = fetcher.Get("path/to/buildpack", freezer.Uncached)
	Expect(err).NotTo(HaveOccurred())

	localBuildpackCached, err = fetcher.Get("path/to/buildpack", freezer.Cached)
	Expect(err).NotTo(HaveOccurred())

	remoteBuildpack, err = fetcher.Get("github.com/remote/buildpack", freezer.Uncached)
	Expect(err).NotTo(HaveOccurred())

	remoteBuildpackCached, err = fetcher.Get("github.com/remote/buildpack", freezer.Cached)
	Expect(err).NotTo(HaveOccurred())
}
```

## Cleaning Up Cache Corruption
If there is any cache corruption you can go to `$HOME/.freezer-cache` and either delete all of the contents or find the offending file and delete that. Local buildpacks are under their name and if you have a cached version it will be in a sub directory named `cached`, if you are dealing with a remote buildpack it will be under in a directory that is the org you pulled it from then in a directory that is the name of the repo and if you have a cached version it will be in a sub directory named `cached`.  If you delete any of these files they will be rebuilt or fetched on your next run.   
