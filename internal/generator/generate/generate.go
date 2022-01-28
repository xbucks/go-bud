package generate

import (
	_ "embed"

	"gitlab.com/mnm/bud/internal/di"
	"gitlab.com/mnm/bud/mod"

	"gitlab.com/mnm/bud/budfs"
	"gitlab.com/mnm/bud/gen"
	"gitlab.com/mnm/bud/internal/gotemplate"
	"gitlab.com/mnm/bud/internal/imports"
)

//go:embed generate.gotext
var template string

var generator = gotemplate.MustParse("generate", template)

type Generator struct {
	BFS      budfs.FS
	Injector *di.Injector
	Module   *mod.Module
	Embed    bool
	Hot      bool
	Minify   bool
}

type State struct {
	Imports  []*imports.Import
	Provider *di.Provider
	Embed    bool
	Hot      bool
	Minify   bool
}

func (g *Generator) GenerateFile(_ gen.F, file *gen.File) error {
	// Don't create a generate file if custom user-generators don't exist
	if err := gen.SkipUnless(g.BFS, "bud/generator/generator.go"); err != nil {
		return err
	}
	imports := imports.New()
	imports.AddStd("os")
	// imports.AddStd("fmt")
	imports.AddNamed("console", "gitlab.com/mnm/bud/log/console")
	imports.AddNamed("mod", "gitlab.com/mnm/bud/mod")
	imports.AddNamed("budfs", "gitlab.com/mnm/bud/budfs")
	imports.AddNamed("generate", "gitlab.com/mnm/bud/generate")
	provider, err := g.Injector.Wire(&di.Function{
		Name:   "Load",
		Target: g.Module.Import("bud/generate"),
		Params: []di.Dependency{
			&di.Type{
				Import: "gitlab.com/mnm/bud/budfs",
				Type:   "FS",
			},
			&di.Type{
				Import: "gitlab.com/mnm/bud/mod",
				Type:   "*Module",
			},
		},
		Results: []di.Dependency{
			&di.Type{
				Import: g.Module.Import("bud/generator"),
				Type:   "Generators",
			},
			&di.Error{},
		},
	})
	if err != nil {
		return err
	}
	for _, imp := range provider.Imports {
		imports.AddNamed(imp.Name, imp.Path)
	}
	code, err := generator.Generate(&State{
		Imports:  imports.List(),
		Provider: provider,
		Embed:    g.Embed,
		Minify:   g.Minify,
		Hot:      g.Hot,
	})
	if err != nil {
		return err
	}
	file.Write(code)
	return nil
}
