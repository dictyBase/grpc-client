package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestLookupSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("STOCK_API_SERVICE_HOST", "stock-api.dev.svc")
	t.Setenv("STOCK_API_SERVICE_PORT", "9345")

	var gotHost, gotPort string
	app := &cli.Command{
		Name: "plasmid",
		Commands: []*cli.Command{
			{
				Name: "plasmid",
				Commands: []*cli.Command{
					{
						Name: "lookup",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "host",
								Sources: cli.EnvVars("STOCK_API_SERVICE_HOST"),
							},
							&cli.StringFlag{
								Name:    "port",
								Sources: cli.EnvVars("STOCK_API_SERVICE_PORT"),
							},
							&cli.StringFlag{
								Name:     "name",
								Required: true,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"app", "plasmid", "lookup", "--name", "pDGB_A1"})
	require.NoError(t, err)
	require.Equal(t, "stock-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
}

func TestListSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("STOCK_API_SERVICE_HOST", "stock-api.dev.svc")
	t.Setenv("STOCK_API_SERVICE_PORT", "9345")

	var gotHost, gotPort string
	app := &cli.Command{
		Name: "plasmid",
		Commands: []*cli.Command{
			{
				Name: "plasmid",
				Commands: []*cli.Command{
					{
						Name: "list",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "host",
								Sources: cli.EnvVars("STOCK_API_SERVICE_HOST"),
							},
							&cli.StringFlag{
								Name:    "port",
								Sources: cli.EnvVars("STOCK_API_SERVICE_PORT"),
							},
							&cli.StringFlag{
								Name:  "filter",
								Value: "summary=~GoldenBraid",
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"app", "plasmid", "list"})
	require.NoError(t, err)
	require.Equal(t, "stock-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
}

func TestFetchSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("STOCK_API_SERVICE_HOST", "stock-api.dev.svc")
	t.Setenv("STOCK_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotID string
	app := &cli.Command{
		Name: "goldenbraid-list",
		Commands: []*cli.Command{
			{
				Name: "plasmid",
				Commands: []*cli.Command{
					{
						Name: "fetch",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "host",
								Sources: cli.EnvVars("STOCK_API_SERVICE_HOST"),
							},
							&cli.StringFlag{
								Name:    "port",
								Sources: cli.EnvVars("STOCK_API_SERVICE_PORT"),
							},
							&cli.StringFlag{
								Name:     "identifier",
								Aliases:  []string{"i"},
								Required: true,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							gotID = cmd.String("identifier")
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(
		context.Background(),
		[]string{"app", "plasmid", "fetch", "--identifier", "DBP0000001"},
	)
	require.NoError(t, err)
	require.Equal(t, "stock-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "DBP0000001", gotID)
}
