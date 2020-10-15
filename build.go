package makeaddon

import (
	"archive/zip"
	"github.com/Masterminds/vcs"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Builder manages the build process for an addon.
type Builder struct {
	dir    string
	data   *MetaData
	cache  *Cache
	writer *zip.Writer
	mapper FolderMap
}

// NewBuilder creates a new builder for an addon in the given directory, which will be written as a zip file to
// the given output stream. Returns an error if the addon metadata could not be found. The actual build process
// is invoked by calling the Build() function.
func NewBuilder(dir string, out io.Writer) (*Builder, error) {
	data, err := MetaDataFromDirectory(dir)
	if err != nil {
		return nil, err
	}

	return &Builder{
		dir:    dir,
		data:   data,
		cache:  NewCache(),
		writer: zip.NewWriter(out),
		mapper: NewFolderMap(data),
	}, nil
}

// Build creates an addon, checking out dependencies and copying over source files.
func (b *Builder) Build() error {
	for dest := range b.data.Externals {
		if err := b.checkout(b.data.Externals[dest], dest); err != nil {
			return err
		}
	}

	if err := b.copyFiles(b.dir, "", ""); err != nil {
		return err
	}

	return b.writer.Close()
}

// checkout clones the remote repository to a temporary dir then adds it to the addon zip at the target location.
func (b *Builder) checkout(config External, target string) error {
	log.Printf("Checking out dependency %s", config.Url)
	dir, fresh := b.cache.Dir(config.Url, config.Tag)

	var repo vcs.Repo
	var err error
	if strings.HasPrefix(config.Url, "https://repos.wowace.com/wow/") {
		repo, err = vcs.NewSvnRepo(config.Url, dir)
	} else {
		repo, err = vcs.NewRepo(config.Url, dir)
	}
	if err != nil {
		return err
	}

	if fresh {
		if err = repo.Get(); err != nil {
			return err
		}
		if len(config.Tag) > 0 {
			if err = repo.UpdateVersion(config.Tag); err != nil {
				return err
			}
		}
	}

	if config.Tag == "latest" {
		// TODO: Find latest tag and use that version instead
	} else if config.Tag == "" {
		if err = repo.Update(); err != nil {
			return err
		}
	}

	return b.copyFiles(dir, "", target)
}

// copyFiles recursively copies all files from the basePath+subDir into the outDir folder of the addon zip.
func (b *Builder) copyFiles(basePath, subDir, outDir string) error {
	files, err := ioutil.ReadDir(filepath.Join(basePath, subDir))
	if err != nil {
		return err
	}

	for i := range files {
		file := filepath.Join(subDir, files[i].Name())
		resolved, ok := b.mapper.Resolve(filepath.Join(outDir, file))
		if !ok {
			continue
		}

		if files[i].IsDir() {
			if err := b.copyFiles(basePath, file, outDir); err != nil {
				return err
			}
		} else if err := b.copyFile(resolved, filepath.Join(basePath, file)); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) copyFile(target, file string) error {
	w, err := b.writer.Create(target)
	if err != nil {
		return err
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

// FolderMap holds a map of on-disk locations to target locations within the addon zip file. A special value of "-"
// indicates that the file or folder should be ignored.
type FolderMap map[string]string

// Resolve maps the given relative path based on the mappings defined in the package metadata and ignore lists.
// The second return value indicates whether the file/folder should be copied; if it is false then the returned
// path should be ignored.
func (f FolderMap) Resolve(p string) (string, bool) {
	// Ignore any file that starts with a period
	if strings.HasPrefix(path.Base(p), ".") {
		return "", false
	}

	// Find the longest matching folder in our map
	sanitised := strings.ToLower(strings.ReplaceAll(p, "\\", "/"))
	for {
		if match, ok := f[sanitised]; ok {
			if match == "-" {
				return "", false
			}

			return path.Join(match, strings.TrimLeft(p[len(sanitised):], "/")), true
		}

		index := strings.LastIndex(sanitised, "/")
		if index == -1 {
			if len(sanitised) == 0 {
				return "", false
			}
			sanitised = ""
		} else {
			sanitised = sanitised[0:index]
		}
	}
}

// NewFolderMap creates a FolderMap based on the metadata provided.
func NewFolderMap(data *MetaData) FolderMap {
	folders := FolderMap{
		"": data.PackageAs,

		// Make sure we don't try to include our own output...
		"addon.zip": "-",
	}

	for src := range data.MoveFolders {
		dst := data.MoveFolders[src]
		folders[strings.ToLower(strings.TrimLeft(strings.TrimPrefix(src, data.PackageAs), "/"))] = dst
	}

	for i := range data.Ignore {
		folders[data.Ignore[i]] = "-"
	}

	return folders
}
