package overlay

import (
	"context"

	"gitlab.com/mnm/bud/internal/dsync"
	"gitlab.com/mnm/bud/internal/fscache"
	"gitlab.com/mnm/bud/internal/pubsub"

	"io/fs"

	"gitlab.com/mnm/bud/internal/dag"
	"gitlab.com/mnm/bud/package/merged"

	"gitlab.com/mnm/bud/package/conjure"
	"gitlab.com/mnm/bud/package/gomod"
	"gitlab.com/mnm/bud/package/pluginfs"
)

// Load the overlay filesystem
func Load(module *gomod.Module) (*FileSystem, error) {
	cache := fscache.New()
	pluginFS, err := pluginfs.Load(module)
	if err != nil {
		return nil, err
	}
	cfs := conjure.New()
	merged := merged.Merge(cache.Wrap("cfs", cfs), cache.Wrap("pluginfs", pluginFS))
	dag := dag.New()
	ps := pubsub.New()
	return &FileSystem{cache, cfs, dag, cache.Wrap("merged", merged), module, ps}, nil
}

// Serve is just load without the cache
// TODO: consolidate
func Serve(module *gomod.Module) (*Server, error) {
	pluginFS, err := pluginfs.Load(module)
	if err != nil {
		return nil, err
	}
	cfs := conjure.New()
	merged := merged.Merge(cfs, pluginFS)
	dag := dag.New()
	ps := pubsub.New()
	return &FileSystem{fscache.New(), cfs, dag, merged, module, ps}, nil
}

type Server = FileSystem

type F interface {
	fs.FS
	Link(from, to string)
}

type FileSystem struct {
	cache  *fscache.Cache
	cfs    *conjure.FileSystem
	dag    *dag.Graph
	fsys   fs.FS
	module *gomod.Module
	ps     pubsub.Client
}

func (f *FileSystem) Link(from, to string) {
}

func (f *FileSystem) Open(name string) (fs.File, error) {
	// fmt.Println("overlay opening", name)
	return f.fsys.Open(name)
}

var _ fs.FS = (*FileSystem)(nil)

type GenerateFile func(ctx context.Context, fsys F, file *File) error

func (fn GenerateFile) GenerateFile(ctx context.Context, fsys F, file *File) error {
	return fn(ctx, fsys, file)
}

func (f *FileSystem) GenerateFile(path string, fn func(ctx context.Context, fsys F, file *File) error) {
	f.cfs.GenerateFile(path, func(file *conjure.File) error {
		return fn(context.TODO(), f, &File{File: file})
	})
}

func (f *FileSystem) FileGenerator(path string, generator FileGenerator) {
	f.GenerateFile(path, generator.GenerateFile)
}

type GenerateDir func(ctx context.Context, fsys F, dir *Dir) error

func (fn GenerateDir) GenerateDir(ctx context.Context, fsys F, dir *Dir) error {
	return fn(ctx, fsys, dir)
}

func (f *FileSystem) GenerateDir(path string, fn func(ctx context.Context, fsys F, dir *Dir) error) {
	f.cfs.GenerateDir(path, func(dir *conjure.Dir) error {
		return fn(context.TODO(), f, &Dir{f, dir})
	})
}

func (f *FileSystem) DirGenerator(path string, generator DirGenerator) {
	f.GenerateDir(path, generator.GenerateDir)
}

func (f *FileSystem) ServeFile(path string, fn func(ctx context.Context, fsys F, file *File) error) {
	f.cfs.ServeFile(path, func(file *conjure.File) error {
		return fn(context.TODO(), f, &File{file})
	})
}

func (f *FileSystem) FileServer(path string, server FileServer) {
	f.ServeFile(path, server.ServeFile)
}

// Sync the overlay to the filesystem
func (f *FileSystem) Sync(dir string) error {
	// Clear the filesystem cache before syncing again
	f.cache.Clear()
	return dsync.Dir(f.fsys, dir, f.module.DirFS(dir), ".")
}
