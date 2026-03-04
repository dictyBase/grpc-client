package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/client"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/wait"
	"github.com/urfave/cli/v3"
)

func main() { //nolint:funlen
	app := &cli.Command{
		Name:  "goldenbraid-list",
		Usage: "List GoldenBraid plasmids from stock API",
		Commands: []*cli.Command{
			{
				Name:  "list-all",
				Usage: "List all plasmids without filter, fetched 30 at a time",
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
				},
				Action: client.ListAllPlasmids,
			},
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
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Limit for the number of results",
						Value: client.DefaultLookupLimit,
					},
				},
				Action: client.LookupPlasmidByName,
			},
			{
				Name:  "wait-job",
				Usage: "Wait for a Kubernetes job to complete, detecting stuck pods early",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Job name to wait for",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "namespace",
						Usage: "Kubernetes namespace",
						Value: "dev",
					},
					&cli.StringFlag{
						Name:  "timeout",
						Usage: "Maximum wait duration (e.g. 60s, 5m)",
						Value: "60s",
					},
					&cli.StringFlag{
						Name:    "kubeconfig",
						Usage:   "Path to kubeconfig file",
						Sources: cli.EnvVars("KUBECONFIG"),
					},
				},
				Action: wait.JobAction,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
