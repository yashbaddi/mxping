package main

import (
	"context"
	"os"

	"github.com/yashbaddi/mxping/internal/cli"
)

func main() {
	app := cli.NewApp()
	code := app.Run(context.Background(), os.Args[1:])
	os.Exit(code)
}
