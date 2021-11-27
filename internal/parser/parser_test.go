package parser_test

import (
	"path/filepath"
	"testing"

	"gitlab.com/mnm/bud/internal/modcache"

	"github.com/matryer/is"
	"gitlab.com/mnm/bud/go/mod"
	"gitlab.com/mnm/bud/internal/parser"
	"gitlab.com/mnm/bud/internal/txtar"
	"gitlab.com/mnm/bud/vfs"
)

func TestStructLookup(t *testing.T) {
	is := is.New(t)
	testfile, err := txtar.ParseFile("testdata/struct-lookup.txt")
	is.NoErr(err)
	dir := t.TempDir()
	err = vfs.Write(dir, testfile)
	is.NoErr(err)
	modCache := modcache.New(filepath.Join(dir, "mod"))
	p := parser.New(mod.New(modCache))
	pkg, err := p.Parse(filepath.Join(dir, "app", "hello"))
	is.NoErr(err)
	is.Equal(pkg.Name(), "hello")
	stct := pkg.Struct("A")
	is.True(stct != nil)
	field := stct.Field("S")
	is.True(field != nil)
	def, err := field.Definition()
	is.NoErr(err)
	is.Equal(def.Name(), "Struct")
	pkg = def.Package()
	is.Equal(pkg.Name(), "two")
	modFile, err := pkg.Modfile()
	is.NoErr(err)
	is.Equal(modFile.ModulePath(), "mod.test/two")
	stct = pkg.Struct("Struct")
	is.True(stct != nil)
	field = stct.Field("Dep")
	is.True(field != nil)
	def, err = field.Definition()
	is.NoErr(err)
	is.Equal(def.Name(), "Dep")
	pkg = def.Package()
	is.Equal(pkg.Name(), "inner")
	stct = pkg.Struct("Dep")
	is.True(stct != nil)
	modFile, err = pkg.Modfile()
	is.NoErr(err)
	is.Equal(modFile.ModulePath(), "mod.test/three")
}

func TestInterfaceLookup(t *testing.T) {
	is := is.New(t)
	testfile, err := txtar.ParseFile("testdata/interface-lookup.txt")
	is.NoErr(err)
	dir := t.TempDir()
	err = vfs.Write(dir, testfile)
	is.NoErr(err)
	modCache := modcache.New(filepath.Join(dir, "mod"))
	p := parser.New(mod.New(modCache))
	pkg, err := p.Parse(filepath.Join(dir, "app", "hello"))
	is.NoErr(err)
	is.Equal(pkg.Name(), "hello")
	stct := pkg.Struct("A")
	is.True(stct != nil)
	field := stct.Field("S")
	is.True(field != nil)
	def, err := field.Definition()
	is.NoErr(err)
	is.Equal(def.Name(), "Interface")
	pkg = def.Package()
	is.Equal(pkg.Name(), "two")
	modFile, err := pkg.Modfile()
	is.NoErr(err)
	is.Equal(modFile.ModulePath(), "mod.test/two")
	iface := pkg.Interface("Interface")
	is.True(iface != nil)
	is.Equal(iface.Name(), "Interface")
	method := iface.Method("Test")
	is.True(method != nil)
	results := method.Results()
	is.Equal(len(results), 1)
	def, err = results[0].Definition()
	is.NoErr(err)
	is.Equal(def.Name(), "Interface")
	pkg = def.Package()
	is.Equal(pkg.Name(), "inner")
	iface = pkg.Interface("Interface")
	is.True(iface != nil)
	method = iface.Method("String")
	is.True(method != nil)
	is.Equal(method.Name(), "String")
	modFile, err = pkg.Modfile()
	is.NoErr(err)
	is.Equal(modFile.ModulePath(), "mod.test/three")
}
func TestAliasLookup(t *testing.T) {
	is := is.New(t)
	testfile, err := txtar.ParseFile("testdata/alias-lookup.txt")
	is.NoErr(err)
	dir := t.TempDir()
	err = vfs.Write(dir, testfile)
	is.NoErr(err)
	modCache := modcache.New(filepath.Join(dir, "mod"))
	p := parser.New(mod.New(modCache))
	pkg, err := p.Parse(filepath.Join(dir, "app"))
	is.NoErr(err)
	is.Equal(pkg.Name(), "main")
	alias := pkg.Alias("Middleware")
	is.True(alias != nil)
	is.Equal(alias.Name(), "Middleware")
	def, err := alias.Definition()
	is.NoErr(err)
	is.Equal(def.Name(), "Middleware")
	pkg = def.Package()
	is.Equal(pkg.Name(), "public")
	alias = pkg.Alias("Middleware")
	is.True(alias != nil)
	def, err = alias.Definition()
	is.NoErr(err)
	is.Equal(def.Name(), "Interface")
	pkg = def.Package()
	is.Equal(pkg.Name(), "public")
	middleware := pkg.Interface("Interface")
	is.True(middleware != nil)
	method := middleware.Method("Middleware")
	is.True(method != nil)
	is.Equal(method.Name(), "Middleware")
}
