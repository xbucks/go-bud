package command

import (
	"context"
	_ "embed"
	"fmt"
	"io/fs"

	"gitlab.com/mnm/bud/internal/gotemplate"
	"gitlab.com/mnm/bud/internal/imports"
	"gitlab.com/mnm/bud/package/overlay"
	"gitlab.com/mnm/bud/pkg/gomod"
	goparse "gitlab.com/mnm/bud/pkg/parser"
)

//go:embed command.gotext
var template string

var generator = gotemplate.MustParse("command.gotext", template)

func Generate(state *State) ([]byte, error) {
	if state.Command == nil {
		return nil, fmt.Errorf("command: generator must have a root command")
	}
	return generator.Generate(state)
}

func New(fs fs.FS, module *gomod.Module, parser *goparse.Parser) *Command {
	return &Command{fs, module, parser}
}

type Command struct {
	fs     fs.FS
	module *gomod.Module
	parser *goparse.Parser
}

func (c *Command) Parse(ctx context.Context) (*State, error) {
	return (&parser{
		fs:      c.fs,
		module:  c.module,
		parser:  c.parser,
		imports: imports.New(),
	}).Parse(ctx)
}

func (c *Command) Compile(ctx context.Context) ([]byte, error) {
	// Parse project commands into state
	state, err := c.Parse(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: Add in the core commands or a default command

	// Generate code from the state
	return Generate(state)
}

func (c *Command) GenerateFile(ctx context.Context, fsys overlay.F, file *overlay.File) error {
	code, err := c.Compile(ctx)
	if err != nil {
		return err
	}
	file.Data = code
	return nil
}