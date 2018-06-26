// Command rewrite-history
package main

import (
	"fmt"
	"os"

	"github.com/benhinchley/cmd"
)

func main() {
	p, err := cmd.NewProgram("rewrite-history", "", nil, []cmd.Command{
		&moveCmd{},
	})
	if err != nil {
		exitWithCode(err, 1)
	}

	if err := p.Run(os.Args, setup); err != nil {
		switch err.(type) {
		case *cmd.ErrNoSuchCommand:
			exitWithCode(err, 127)
		case *cmd.ErrNoDefaultCommand:
			exitWithCode(err, 126)
		default:
			exitWithCode(err, 1)
		}
	}
}

func exitWithCode(err error, code int) {
	fmt.Fprintf(os.Stderr, "%v", err)
	os.Exit(code)
}

func setup(env *cmd.Environment, c cmd.Command, args []string) error {
	ctx := &Context{
		Context: env.GetDefaultContext(),
	}

	return c.Run(ctx, args)
}

type Context struct {
	cmd.Context
}
