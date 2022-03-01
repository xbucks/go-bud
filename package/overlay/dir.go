package overlay

import (
	"context"

	"gitlab.com/mnm/bud/package/conjure"
)

type DirGenerator interface {
	GenerateDir(ctx context.Context, fsys F, dir *Dir) error
}

type Dir struct {
	fsys F
	*conjure.Dir
}

func (d *Dir) GenerateFile(path string, fn func(ctx context.Context, fsys F, file *File) error) {
	d.Dir.GenerateFile(path, func(ctx context.Context, file *conjure.File) error {
		return fn(ctx, d.fsys, &File{file})
	})
}

func (d *Dir) FileGenerator(path string, generator FileGenerator) {
	d.GenerateFile(path, generator.GenerateFile)
}

func (d *Dir) GenerateDir(path string, fn func(ctx context.Context, fsys F, dir *Dir) error) {
	d.Dir.GenerateDir(path, func(ctx context.Context, dir *conjure.Dir) error {
		return fn(ctx, d.fsys, &Dir{d.fsys, dir})
	})
}

func (d *Dir) DirGenerator(path string, generator DirGenerator) {
	d.GenerateDir(path, generator.GenerateDir)
}
