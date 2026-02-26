package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/client"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "goldenbraid-list",
		Usage: "List GoldenBraid plasmids from stock API",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List GoldenBraid plasmids matching a filter",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "host",
						Usage:   "gRPC server host address",
						Sources: cli.EnvVars("STOCK_API_SERVICE_HOST"),
					},
					&cli.StringFlag{
						Name:    "port",
						Usage:   "gRPC server port",
						Sources: cli.EnvVars("STOCK_API_SERVICE_PORT"),
					},
					&cli.StringFlag{
						Name:    "filter",
						Usage:   "Filter string for the stock API",
						Value:   "summary=~GoldenBraid",
						Sources: cli.EnvVars("STOCK_API_FILTER"),
					},
				},
				Action: client.ListPlasmids,
			},
			{
				Name:  "lookup",
				Usage: "Look up a GoldenBraid plasmid by exact name",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "host",
						Usage:   "gRPC server host address",
						Sources: cli.EnvVars("STOCK_API_SERVICE_HOST"),
					},
					&cli.StringFlag{
						Name:    "port",
						Usage:   "gRPC server port",
						Sources: cli.EnvVars("STOCK_API_SERVICE_PORT"),
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Exact plasmid name to look up (e.g. pDGB3alpha1)",
						Required: true,
						Sources:  cli.EnvVars("PLASMID_NAME"),
					},
				},
				Action: client.LookupPlasmidByName,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
