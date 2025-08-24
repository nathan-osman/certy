package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nathan-osman/certy/server"
	"github.com/nathan-osman/certy/storage"
	"github.com/nathan-osman/gosvc"
	"github.com/urfave/cli/v2"
)

var (
	a = &gosvc.Application{
		Name:        "certy",
		Description: "Manage X.509 certs. & keys",
	}
)

func main() {
	app := &cli.App{
		Name:  "certy",
		Usage: "Simple web-based certificate authority",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "data-dir",
				EnvVars: []string{"DATA_DIR"},
				Usage:   "path to data directory",
			},
			&cli.StringFlag{
				Name:    "server-addr",
				Value:   ":80",
				EnvVars: []string{"SERVER_ADDR"},
				Usage:   "HTTP address to listen on",
			},
		},
		Commands: gosvc.Commands(a.Platform()),
		Action: func(c *cli.Context) error {

			// Create the storage instance
			st, err := storage.New(&storage.Config{
				DataDir: c.String("data-dir"),
			})
			if err != nil {
				return err
			}

			// Start the server
			s, err := server.New(&server.Config{
				Addr:    c.String("server-addr"),
				Storage: st,
			})
			if err != nil {
				return err
			}
			defer s.Close()

			// Wait for SIGINT or SIGTERM
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s\n", err.Error())
		os.Exit(1)
	}
}
