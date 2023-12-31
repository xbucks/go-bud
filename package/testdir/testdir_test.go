package testdir_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/livebud/bud/internal/is"
	"github.com/livebud/bud/internal/versions"
	"github.com/livebud/bud/package/testdir"
)

func TestDir(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	td, err := testdir.Load()
	is.NoErr(err)
	td.Modules["github.com/livebud/bud-test-plugin"] = "v0.0.2"
	td.Files["controller/controller.go"] = `package controller`
	td.Bytes["public/favicon.ico"] = []byte{0x00}
	td.NodeModules["svelte"] = versions.Svelte
	td.NodeModules["livebud"] = "*"
	is.NoErr(td.Write(ctx))
	is.NoErr(td.Exists(
		"controller/controller.go",
		"public/favicon.ico",
		"node_modules/svelte/package.json",
		"node_modules/livebud/package.json",
		"package.json",
		"go.mod",
	))
	// Ensure livebud doesn't leak into dir
	is.NoErr(td.NotExists(
		"qs",
		"url",
		"tsconfig.json",
	))
}

func TestRefresh(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	td, err := testdir.Load()
	is.NoErr(err)
	td.Modules["github.com/livebud/bud-test-plugin"] = "v0.0.2"
	td.Files["controller/controller.go"] = `package controller`
	td.Bytes["public/favicon.ico"] = []byte{0x00}
	td.NodeModules["svelte"] = versions.Svelte
	td.NodeModules["livebud"] = "*"
	is.NoErr(td.Write(ctx))
	is.NoErr(td.Exists(
		"controller/controller.go",
		"public/favicon.ico",
		"node_modules/svelte/package.json",
		"node_modules/livebud/package.json",
		"package.json",
		"go.mod",
	))
	favicon := []byte{0x01}
	td.Bytes["public/favicon.ico"] = favicon
	td.NodeModules["uid"] = "2.0.0"
	is.NoErr(td.Write(ctx))
	is.NoErr(td.Exists(
		"controller/controller.go",
		"public/favicon.ico",
		"node_modules/livebud/package.json",
		"node_modules/svelte/package.json",
		"node_modules/uid/package.json",
		"package.json",
		"go.mod",
	))
	fav, err := fs.ReadFile(td, "public/favicon.ico")
	is.NoErr(err)
	is.Equal(favicon, fav)
}

func TestOverwrite(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()
	td, err := testdir.Load()
	is.NoErr(err)
	td.Files["controller/controller.go"] = `
		package controller
		type Controller struct {}
		func (c *Controller) Index() string { return "" }
	`
	td.Files["view/index.svelte"] = `<h1>hello</h1>`
	is.NoErr(td.Write(ctx))
	is.NoErr(td.Exists("controller/controller.go"))
	is.NoErr(td.Exists("view/index.svelte"))
	controller1, err := fs.ReadFile(td, "controller/controller.go")
	is.NoErr(err)
	view1, err := fs.ReadFile(td, "view/index.svelte")
	is.NoErr(err)
	is.Equal(string(view1), `<h1>hello</h1>`)
	td.Files["view/index.svelte"] = `<h1>hi</h1>`
	is.NoErr(td.Write(ctx))
	is.NoErr(td.Exists("controller/controller.go"))
	is.NoErr(td.Exists("view/index.svelte"))
	controller2, err := fs.ReadFile(td, "controller/controller.go")
	is.NoErr(err)
	view2, err := fs.ReadFile(td, "view/index.svelte")
	is.NoErr(err)
	is.Equal(string(controller1), string(controller2))
	is.Equal(string(view2), `<h1>hi</h1>`)

	// Remove
	is.NoErr(td.RemoveAll("view/index.svelte"))
	view3, err := fs.ReadFile(td, "view/index.svelte")
	is.True(err != nil)
	is.True(errors.Is(err, fs.ErrNotExist))
	is.Equal(view3, nil)
}
