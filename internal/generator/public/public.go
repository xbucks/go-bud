package public

import (
	_ "embed"
	"io/fs"

	"gitlab.com/mnm/bud/go/mod"
	"gitlab.com/mnm/bud/vfs"

	"gitlab.com/mnm/bud/gen"
	"gitlab.com/mnm/bud/internal/gotemplate"
)

//go:embed public.gotext
var template string

var generator = gotemplate.MustParse("public", template)

type Generator struct {
	Module *mod.Module
	Embed  bool
	Minify bool
}

type State struct {
	Embed bool
	Files []*File
}

type File struct {
	Path string
}

func (g *Generator) GenerateFile(_ gen.F, file *gen.File) error {
	exist := vfs.SomeExist(g.Module, "public")
	if len(exist) == 0 {
		return fs.ErrNotExist
	}
	code, err := generator.Generate(State{})
	if err != nil {
		return err
	}
	file.Write(code)
	return nil
}
