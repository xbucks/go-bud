package main

import (
	"context"
	"os"
	"strings"

	"gitlab.com/mnm/bud/internal/bud"
	"gitlab.com/mnm/bud/internal/bud/build"
	"gitlab.com/mnm/bud/internal/bud/run"
	"gitlab.com/mnm/bud/internal/bud/tool/di"
	v8 "gitlab.com/mnm/bud/internal/bud/tool/v8"
	v8client "gitlab.com/mnm/bud/internal/bud/tool/v8/client"

	"gitlab.com/mnm/bud/package/commander"

	"gitlab.com/mnm/bud/pkg/log/console"
)

func main() {
	if err := do(); err != nil {
		if !isExitStatus(err) {
			console.Error(err.Error())
		}
		os.Exit(1)
	}
}

func do() error {
	// $ bud
	bud := new(bud.Command)
	cli := commander.New("bud")
	cli.Flag("chdir", "Change the working directory").Short('C').String(&bud.Dir).Default(".")
	cli.Flag("trace", "Enable tracing").Short('t').Bool(&bud.Trace).Default(false)
	cli.Args("args").Strings(&bud.Args)
	cli.Run(bud.Run)

	{ // $ bud run
		cmd := &run.Command{Bud: bud}
		cli := cli.Command("run", "run the development server")
		cli.Flag("embed", "embed the assets").Bool(&bud.Embed).Default(false)
		cli.Flag("hot", "hot reload the frontend").Bool(&bud.Hot).Default(true)
		cli.Flag("minify", "minify the assets").Bool(&bud.Minify).Default(false)
		cli.Flag("port", "port").String(&cmd.Port).Default("3000")
		cli.Run(cmd.Run)
	}

	{ // $ bud build
		cmd := &build.Command{Bud: bud}
		cli := cli.Command("build", "build the production server")
		cli.Flag("embed", "embed the assets").Bool(&bud.Embed).Default(true)
		cli.Flag("hot", "hot reload the frontend").Bool(&bud.Hot).Default(false)
		cli.Flag("minify", "minify the assets").Bool(&bud.Minify).Default(true)
		cli.Run(cmd.Run)
	}

	{ // $ bud tool
		cli := cli.Command("tool", "extra tools")

		{ // $ bud tool di
			cmd := &di.Command{Bud: bud}
			cli := cli.Command("di", "dependency injection generator")
			cli.Flag("dependency", "generate dependency provider").Short('d').Strings(&cmd.Dependencies)
			cli.Flag("external", "mark dependency as external").Short('e').Strings(&cmd.Externals).Optional()
			cli.Flag("map", "map interface types to concrete types").Short('m').StringMap(&cmd.Map).Optional()
			cli.Flag("target", "target import path").Short('t').String(&cmd.Target)
			cli.Flag("hoist", "hoist dependencies that depend on externals").Bool(&cmd.Hoist).Default(false)
			cli.Flag("verbose", "verbose logging").Short('v').Bool(&cmd.Verbose).Default(false)
			cli.Run(cmd.Run)
		}

		{ // $ bud tool v8
			cmd := &v8.Command{Bud: bud}
			cli := cli.Command("v8", "Execute Javascript with V8 from stdin")
			cli.Run(cmd.Run)

			{ // $ bud tool v8 client
				cmd := &v8client.Command{Bud: bud}
				cli := cli.Command("client", "V8 client used during development")
				cli.Run(cmd.Run)
			}
		}
	}
	ctx := context.Background()
	return cli.Parse(ctx, os.Args[1:])
}

func isExitStatus(err error) bool {
	return strings.Contains(err.Error(), "exit status ")
}
