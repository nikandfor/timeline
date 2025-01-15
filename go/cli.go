//go:build ignore

package main

import (
	"io"
	"os"
	"runtime/pprof"

	"nikand.dev/go/cli"
	"nikand.dev/go/cli/flag"
	"tlog.app/go/errors"
	"tlog.app/go/tlog"
	"tlog.app/go/tlog/ext/tlflag"

	"nikand.dev/go/timeline/go/timeline"
)

func main() {
	app := &cli.Command{
		Name:   "heatmap",
		Action: run,
		Args:   cli.Args{},
		Flags: []*cli.Flag{
			cli.NewFlag("log", "stderr?console=dm", "log output file (or stderr)"),
			cli.NewFlag("verbosity,v", "", "logger verbosity topics"),
			cli.NewFlag("debug", "", "debug address", flag.Hidden),
			cli.NewFlag("profile", "", "save cpu profile", flag.Hidden),
			cli.FlagfileFlag,
			cli.HelpFlag,
		},
	}

	cli.RunAndExit(app, os.Args, os.Environ())
}

func before(c *cli.Command) error {
	w, err := tlflag.OpenWriter(c.String("log"))
	if err != nil {
		return errors.Wrap(err, "open log file")
	}

	tlog.DefaultLogger = tlog.New(w)

	tlog.SetVerbosity(c.String("verbosity"))

	return nil
}

func run(c *cli.Command) (err error) {
	if q := c.String("profile"); q != "" {
		f, err := os.Create(q)
		if err != nil {
			return errors.Wrap(err, "open profile file")
		}

		defer closer(f, &err, "close profile file")

		err = pprof.StartCPUProfile(f)
		if err != nil {
			return errors.Wrap(err, "start cpu profiling")
		}

		defer pprof.StopCPUProfile()
	}

	data, err := os.ReadFile(c.Args.First())
	if err != nil {
		return errors.Wrap(err, "read file")
	}

	pts, err := timeline.Parse(data, nil)
	if err != nil {
		return errors.Wrap(err, "process")
	}

	_ = pts

	return nil
}

func closer(c io.Closer, errp *error, msg string) {
	err := c.Close()
	if *errp == nil && err != nil {
		*errp = errors.WrapDepth(err, 1, msg)
	}
}
