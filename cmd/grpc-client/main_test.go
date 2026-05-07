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

func TestAnnotationFindByTagSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("ANNOTATION_API_SERVICE_HOST", "annotation-api.dev.svc")
	t.Setenv("ANNOTATION_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotOntology, gotTag string
	var gotLimit, gotCursor int64
	app := &cli.Command{
		Name: "grpc-client",
		Commands: []*cli.Command{
			{
				Name: "annotation",
				Commands: []*cli.Command{
					{
						Name: "findbytag",
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
								Name:  "ontology",
								Value: "",
							},
							&cli.StringFlag{
								Name:  "tag",
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
							gotOntology = cmd.String("ontology")
							gotTag = cmd.String("tag")
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
			"findbytag",
			"--ontology",
			"cellular_component",
			"--tag",
			"GO:0005634",
			"--limit",
			"15",
			"--cursor",
			"3",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "annotation-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "cellular_component", gotOntology)
	require.Equal(t, "GO:0005634", gotTag)
	require.Equal(t, int64(15), gotLimit)
	require.Equal(t, int64(3), gotCursor)
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

func TestAnnotationGroupFindSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("ANNOTATION_API_SERVICE_HOST", "annotation-api.dev.svc")
	t.Setenv("ANNOTATION_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotIdentifier, gotTag, gotOntology string
	var gotLimit, gotCursor int64
	app := &cli.Command{
		Name: "grpc-client",
		Commands: []*cli.Command{
			{
				Name: "annotation",
				Commands: []*cli.Command{
					{
						Name: "groupfind",
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
								Name:     "identifier",
								Aliases:  []string{"i"},
								Required: true,
							},
							&cli.StringFlag{
								Name:  "ontology",
								Value: "",
							},
							&cli.StringFlag{
								Name:  "tag",
								Value: "",
							},
							&cli.IntFlag{
								Name:    "limit",
								Aliases: []string{"l"},
								Value:   100,
							},
							&cli.IntFlag{
								Name:  "cursor",
								Value: 0,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							gotIdentifier = cmd.String("identifier")
							gotTag = cmd.String("tag")
							gotOntology = cmd.String("ontology")
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
			"app", "annotation", "groupfind",
			"--identifier", "DDB_G123",
			"--tag", "GO:0005634",
			"--ontology", "cellular_component",
			"--limit", "50",
			"--cursor", "7",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "annotation-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "DDB_G123", gotIdentifier)
	require.Equal(t, "GO:0005634", gotTag)
	require.Equal(t, "cellular_component", gotOntology)
	require.Equal(t, int64(50), gotLimit)
	require.Equal(t, int64(7), gotCursor)
}

func TestAnnotationRemoveSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("ANNOTATION_API_SERVICE_HOST", "annotation-api.dev.svc")
	t.Setenv("ANNOTATION_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotTag, gotIdentifier, gotOntology string
	app := &cli.Command{
		Name: "grpc-client",
		Commands: []*cli.Command{
			{
				Name: "annotation",
				Commands: []*cli.Command{
					{
						Name: "remove",
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
								Name:     "tag",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "identifier",
								Aliases:  []string{"i"},
								Required: true,
							},
							&cli.StringFlag{
								Name:     "ontology",
								Required: true,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							gotTag = cmd.String("tag")
							gotIdentifier = cmd.String("identifier")
							gotOntology = cmd.String("ontology")
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
			"app", "annotation", "remove", "--tag", "GO:0005634",
			"--identifier", "DDB_G123", "--ontology", "cellular_component",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "annotation-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "GO:0005634", gotTag)
	require.Equal(t, "DDB_G123", gotIdentifier)
	require.Equal(t, "cellular_component", gotOntology)
}

func TestAnnoFeatCreateSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("ANNO_FEAT_API_SERVICE_HOST", "annofeat-api.dev.svc")
	t.Setenv("ANNO_FEAT_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotID, gotName, gotCreatedBy, gotSynonyms, gotProperties string
	app := &cli.Command{
		Name: "grpc-client",
		Commands: []*cli.Command{
			{
				Name: "annofeat",
				Commands: []*cli.Command{
					{
						Name: "create",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "host",
								Sources: cli.EnvVars("ANNO_FEAT_API_SERVICE_HOST"),
							},
							&cli.StringFlag{
								Name:    "port",
								Sources: cli.EnvVars("ANNO_FEAT_API_SERVICE_PORT"),
							},
							&cli.StringFlag{
								Name:     "id",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "name",
								Required: true,
							},
							&cli.StringFlag{
								Name:    "created-by",
								Sources: cli.EnvVars("ANNO_FEAT_CREATED_BY"),
							},
							&cli.StringFlag{
								Name: "synonyms",
							},
							&cli.StringFlag{
								Name: "properties",
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							gotID = cmd.String("id")
							gotName = cmd.String("name")
							gotCreatedBy = cmd.String("created-by")
							gotSynonyms = cmd.String("synonyms")
							gotProperties = cmd.String("properties")
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
			"app", "annofeat", "create",
			"--id", "DDB_G0285425",
			"--name", "Test Feature",
			"--synonyms", "test1,test2",
			"--properties", "description=Test description,note=Test note",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "annofeat-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "DDB_G0285425", gotID)
	require.Equal(t, "Test Feature", gotName)
	require.Equal(t, "", gotCreatedBy)
	require.Equal(t, "test1,test2", gotSynonyms)
	require.Equal(t, "description=Test description,note=Test note", gotProperties)
}

func TestAnnoFeatGetSubcommandPicksUpGRPCEnvVars(t *testing.T) {
	t.Setenv("ANNO_FEAT_API_SERVICE_HOST", "annofeat-api.dev.svc")
	t.Setenv("ANNO_FEAT_API_SERVICE_PORT", "9345")

	var gotHost, gotPort, gotID string
	app := &cli.Command{
		Name: "grpc-client",
		Commands: []*cli.Command{
			{
				Name: "annofeat",
				Commands: []*cli.Command{
					{
						Name: "get",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "host",
								Sources: cli.EnvVars("ANNO_FEAT_API_SERVICE_HOST"),
							},
							&cli.StringFlag{
								Name:    "port",
								Sources: cli.EnvVars("ANNO_FEAT_API_SERVICE_PORT"),
							},
							&cli.StringFlag{
								Name:     "id",
								Required: true,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							gotHost = cmd.String("host")
							gotPort = cmd.String("port")
							gotID = cmd.String("id")
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(
		context.Background(),
		[]string{"app", "annofeat", "get", "--id", "DDB_G0285425"},
	)
	require.NoError(t, err)
	require.Equal(t, "annofeat-api.dev.svc", gotHost)
	require.Equal(t, "9345", gotPort)
	require.Equal(t, "DDB_G0285425", gotID)
}
