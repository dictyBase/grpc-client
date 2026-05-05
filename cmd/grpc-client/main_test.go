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
		Name: "search",
		Commands: []*cli.Command{
			{
				Name: "search",
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

	err := app.Run(context.Background(), []string{"app", "search", "lookup", "--name", "pDGB_A1"})
	require.NoError(t, err)
	require.Equal(t, "stock-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
}

func TestListSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("STOCK_API_SERVICE_HOST", "stock-api.dev.svc")
	t.Setenv("STOCK_API_SERVICE_PORT", "9345")

	var gotHost, gotPort string
	app := &cli.Command{
		Name: "search",
		Commands: []*cli.Command{
			{
				Name: "search",
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

	err := app.Run(context.Background(), []string{"app", "search", "list"})
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
				Name: "search",
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
		[]string{"app", "search", "fetch", "--identifier", "DBP0000001"},
	)
	require.NoError(t, err)
	require.Equal(t, "stock-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "DBP0000001", gotID)
}

func TestStrainFetchSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("STOCK_API_SERVICE_HOST", "stock-api.dev.svc")
	t.Setenv("STOCK_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotID string
	app := &cli.Command{
		Name: "grpc-client",
		Commands: []*cli.Command{
			{
				Name: "strain",
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
		[]string{"app", "strain", "fetch", "--identifier", "DBS0000001"},
	)
	require.NoError(t, err)
	require.Equal(t, "stock-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "DBS0000001", gotID)
}

func TestStrainFilterSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("STOCK_API_SERVICE_HOST", "stock-api.dev.svc")
	t.Setenv("STOCK_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotType string
	var gotLimit, gotCursor int64
	app := &cli.Command{
		Name: "grpc-client",
		Commands: []*cli.Command{
			{
				Name: "strain",
				Commands: []*cli.Command{
					{
						Name: "filter",
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
								Name:    "strain-type",
								Aliases: []string{"st"},
								Value:   "all",
							},
							&cli.IntFlag{
								Name:    "limit",
								Aliases: []string{"l"},
								Value:   10,
							},
							&cli.IntFlag{
								Name:  "cursor",
								Value: 0,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							gotType = cmd.String("strain-type")
							gotLimit = int64(cmd.Int("limit"))
							gotCursor = int64(cmd.Int("cursor"))
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(
		context.Background(),
		[]string{
			"app",
			"strain",
			"filter",
			"--strain-type",
			"general strain",
			"--limit",
			"5",
			"--cursor",
			"10",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "stock-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "general strain", gotType)
	require.Equal(t, int64(5), gotLimit)
	require.Equal(t, int64(10), gotCursor)
}

func TestAnnotationFindSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("ANNOTATION_API_SERVICE_HOST", "annotation-api.dev.svc")
	t.Setenv("ANNOTATION_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotFilter string
	var gotLimit, gotCursor int64
	app := &cli.Command{
		Name: "grpc-client",
		Commands: []*cli.Command{
			{
				Name: "annotation",
				Commands: []*cli.Command{
					{
						Name: "find",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "host",
								Sources: cli.EnvVars("ANNOTATION_API_SERVICE_HOST"),
							},
							&cli.StringFlag{
								Name:    "port",
								Sources: cli.EnvVars("ANNOTATION_API_SERVICE_PORT"),
							},
							&cli.StringFlag{
								Name:  "filter",
								Value: "",
							},
							&cli.IntFlag{
								Name:    "limit",
								Aliases: []string{"l"},
								Value:   10,
							},
							&cli.IntFlag{
								Name:  "cursor",
								Value: 0,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							gotFilter = cmd.String("filter")
							gotLimit = int64(cmd.Int("limit"))
							gotCursor = int64(cmd.Int("cursor"))
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(
		context.Background(),
		[]string{
			"app",
			"annotation",
			"find",
			"--filter",
			"ontology===cellular_component",
			"--limit",
			"20",
			"--cursor",
			"5",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "annotation-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "ontology===cellular_component", gotFilter)
	require.Equal(t, int64(20), gotLimit)
	require.Equal(t, int64(5), gotCursor)
}
