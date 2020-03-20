package freezer

type FileSystem struct {
	tempDir func(string, string) (string, error)
}

func NewFileSystem(tempDir func(string, string) (string, error)) FileSystem {
	return FileSystem{
		tempDir: tempDir,
	}
}

func (fs FileSystem) TempDir(tempDir, tempName string) (string, error) {
	return fs.tempDir(tempDir, tempName)
}
