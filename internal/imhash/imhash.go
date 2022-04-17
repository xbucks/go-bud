package imhash

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/cespare/xxhash"
	"github.com/livebud/bud/package/parser"
	"golang.org/x/sync/errgroup"

	"github.com/livebud/bud/package/gomod"
)

func find(module *gomod.Module, mainDir string) (*fileSet, error) {
	parser := parser.New(module, module)
	fset := newFileSet()
	// Add the following if they exist
	if err := addIfExist(module, fset, "go.mod", "package.json", "package-lock.json"); err != nil {
		return nil, err
	}
	if err := findDeps(fset, module, parser, mainDir); err != nil {
		return nil, err
	}
	return fset, nil
}

func Hash(module *gomod.Module, mainDir string) (string, error) {
	fset, err := find(module, mainDir)
	if err != nil {
		return "", err
	}
	return fset.Hash(module)
}

func Debug(module *gomod.Module, mainDir string, w io.Writer) error {
	fset, err := find(module, mainDir)
	if err != nil {
		return err
	}
	return fset.Debug(module, w)
}

func exists(fsys fs.FS, path string) (bool, error) {
	if _, err := fs.Stat(fsys, path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func addIfExist(module *gomod.Module, fset *fileSet, paths ...string) error {
	eg := new(errgroup.Group)
	for _, path := range paths {
		path := path
		eg.Go(func() error {
			// Add go.mod if it exists
			if exists, err := exists(module, path); err != nil {
				return err
			} else if exists {
				fset.Add(path)
			}
			return nil
		})
	}
	return eg.Wait()
}

func hashFile(fsys fs.FS, filePath string) ([]byte, error) {
	f, err := fsys.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := xxhash.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func newFileSet() *fileSet {
	return &fileSet{
		m: map[string]struct{}{},
	}
}

type fileSet struct {
	mu sync.RWMutex
	m  map[string]struct{}
}

func (s *fileSet) Add(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[path]; !ok {
		s.m[path] = struct{}{}
	}
}

func (s *fileSet) List() (list []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for path := range s.m {
		list = append(list, path)
	}
	sort.Strings(list)
	return list
}

func (s *fileSet) Hash(fsys fs.FS) (string, error) {
	h := xxhash.New()
	for _, file := range s.List() {
		hash, err := hashFile(fsys, file)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%x %s\n", hash, file)
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)), nil
}

func (s *fileSet) Debug(fsys fs.FS, w io.Writer) error {
	for _, file := range s.List() {
		hash, err := hashFile(fsys, file)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%x %s\n", hash, file)
	}
	return nil
}

func shouldWalk(module *gomod.Module, importPath string) bool {
	return module.IsLocal(importPath) ||
		// TODO: consider removing in release mode and are able to turn enable this
		// for development
		strings.HasPrefix(importPath, "github.com/livebud/bud")
}

func findDeps(fset *fileSet, module *gomod.Module, parser *parser.Parser, dir string) (err error) {
	imported, err := parser.Import(dir)
	if err != nil {
		return err
	}
	// Add all the Go files
	for _, path := range imported.GoFiles {
		fset.Add(filepath.Join(dir, path))
	}
	// Add all the embeds
	// TODO: resolve patterns
	for _, path := range imported.EmbedPatterns {
		fset.Add(filepath.Join(dir, path))
	}
	// Traverse imports and compute a hash
	eg := new(errgroup.Group)
	for _, importPath := range imported.Imports {
		importPath := importPath
		eg.Go(func() error {
			if !shouldWalk(module, importPath) {
				return nil
			}
			dir, err := module.ResolveDirectory(importPath)
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(module.Directory(), dir)
			if err != nil {
				return err
			}
			if err := findDeps(fset, module, parser, relPath); err != nil {
				return err
			}
			return nil
		})
	}
	return eg.Wait()
}
