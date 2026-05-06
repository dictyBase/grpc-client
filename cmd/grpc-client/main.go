package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dictyBase/grpc-client/internal/client"
	"github.com/dictyBase/grpc-client/internal/wait"
	"github.com/urfave/cli/v3"
)

func main() {
	app := buildCommandTree()
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func buildCommandTree() *cli.Command {
	return &cli.Command{
		Name:  "grpc-client",
		Usage: "CLI for gRPC stock service operations",
		Commands: []*cli.Command{
			buildSearchCommands(),
			buildStrainCommands(),
			buildAnnotationCommands(),
			buildWaitJobCommand(),
		},
	}
}

func buildSearchCommands() *cli.Command {
	return &cli.Command{
		Name:  "search",
		Usage: "Search for GoldenBraid plasmids in the stock API",
		Commands: []*cli.Command{
			buildListAllCommand(),
			buildListCommand(),
			buildLookupCommand(),
			buildPlasmidFetchCommand(),
		},
	}
}

func buildListAllCommand() *cli.Command {
	return &cli.Command{
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
	}
}

func buildListCommand() *cli.Command {
	return &cli.Command{
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
	}
}

func buildLookupCommand() *cli.Command {
	return &cli.Command{
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
	}
}

func buildPlasmidFetchCommand() *cli.Command {
	return &cli.Command{
		Name:  "fetch",
		Usage: "Fetch a single plasmid by its identifier",
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
				Name:     "identifier",
				Aliases:  []string{"i"},
				Usage:    "Identifier of the plasmid to fetch",
				Required: true,
			},
		},
		Action: client.FetchPlasmid,
	}
}

func buildStrainCommands() *cli.Command {
	return &cli.Command{
		Name:  "strain",
		Usage: "Strain-related operations on the stock API",
		Commands: []*cli.Command{
			buildStrainFetchCommand(),
			buildStrainFilterCommand(),
		},
	}
}

func buildStrainFilterCommand() *cli.Command {
	return &cli.Command{
		Name:  "filter",
		Usage: "List strains by type filter (REMI-seq, general strain, bacterial strain, or all)",
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
				Name:    "strain-type",
				Aliases: []string{"st"},
				Usage:   "Type of strain to filter for (REMI-seq, general strain, bacterial strain, or all)",
				Value:   "all",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Number of strains to fetch",
				Value:   client.DefaultStrainFilterLimit,
			},
			&cli.IntFlag{
				Name:  "cursor",
				Usage: "Offset for fetching list of strains",
				Value: 0,
			},
		},
		Action: client.FilterStrain,
	}
}

func buildStrainFetchCommand() *cli.Command {
	return &cli.Command{
		Name:  "fetch",
		Usage: "Fetch a single strain by its identifier",
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
				Name:     "identifier",
				Aliases:  []string{"i"},
				Usage:    "Identifier of the strain to fetch",
				Required: true,
			},
		},
		Action: client.FetchStrain,
	}
}

func buildWaitJobCommand() *cli.Command {
	return &cli.Command{
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
	}
}

func buildAnnotationCommands() *cli.Command {
	return &cli.Command{
		Name:  "annotation",
		Usage: "Annotation-related operations on the stock API",
		Commands: []*cli.Command{
			buildAnnotationFindCommand(),
			buildAnnotationFindByTagCommand(),
			buildAnnotationGroupFindCommand(),
			buildAnnotationRemoveCommand(),
		},
	}
}

func buildAnnotationFindCommand() *cli.Command {
	return &cli.Command{
		Name:  "find",
		Usage: "Find annotations matching a filter",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Usage:   "gRPC server host address",
				Sources: cli.EnvVars("ANNOTATION_API_SERVICE_HOST"),
			},
			&cli.StringFlag{
				Name:    "port",
				Usage:   "gRPC server port",
				Sources: cli.EnvVars("ANNOTATION_API_SERVICE_PORT"),
			},
			&cli.StringFlag{
				Name:  "filter",
				Usage: "Filter string for annotations (e.g. entry_id===DDB_G123;ontology===cellular_component)",
				Value: "",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Number of annotations to fetch",
				Value:   client.DefaultAnnotationLimit,
			},
			&cli.IntFlag{
				Name:  "cursor",
				Usage: "Offset for fetching list of annotations",
				Value: 0,
			},
		},
		Action: client.FindAnnotation,
	}
}

func buildAnnotationFindByTagCommand() *cli.Command {
	return &cli.Command{
		Name:  "findbytag",
		Usage: "Find annotations filtered by tag and ontology",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Usage:   "gRPC server host address",
				Sources: cli.EnvVars("ANNOTATION_API_SERVICE_HOST"),
			},
			&cli.StringFlag{
				Name:    "port",
				Usage:   "gRPC server port",
				Sources: cli.EnvVars("ANNOTATION_API_SERVICE_PORT"),
			},
			&cli.StringFlag{
				Name:  "ontology",
				Usage: "Ontology name (e.g. cellular_component, biological_process)",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "tag",
				Usage: "Tag or term name in the ontology (e.g. GO:0005634)",
				Value: "",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Number of annotations to fetch",
				Value:   client.DefaultAnnotationLimit,
			},
			&cli.IntFlag{
				Name:  "cursor",
				Usage: "Offset for fetching list of annotations",
				Value: 0,
			},
		},
		Action: client.FindByTag,
	}
}

func buildAnnotationGroupFindCommand() *cli.Command {
	return &cli.Command{
		Name:  "groupfind",
		Usage: "Retrieve annotation groups by identifier, optionally filtered by tag and ontology",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Usage:   "gRPC server host address",
				Sources: cli.EnvVars("ANNOTATION_API_SERVICE_HOST"),
			},
			&cli.StringFlag{
				Name:    "port",
				Usage:   "gRPC server port",
				Sources: cli.EnvVars("ANNOTATION_API_SERVICE_PORT"),
			},
			&cli.StringFlag{
				Name:     "identifier",
				Aliases:  []string{"i"},
				Usage:    "Identifier that will be searched for annotation groups",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "ontology",
				Usage: "Ontology name (e.g. cellular_component, biological_process)",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "tag",
				Usage: "Tag or term name in the ontology (e.g. GO:0005634)",
				Value: "",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Number of annotation groups to fetch",
				Value:   client.DefaultAnnotationGroupLimit,
			},
			&cli.IntFlag{
				Name:  "cursor",
				Usage: "Offset for fetching list of annotation groups",
				Value: 0,
			},
		},
		Action: client.FindAnnotationGroup,
	}
}

func buildAnnotationRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Delete an annotation by tag, identifier, and ontology",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Usage:   "gRPC server host address",
				Sources: cli.EnvVars("ANNOTATION_API_SERVICE_HOST"),
			},
			&cli.StringFlag{
				Name:    "port",
				Usage:   "gRPC server port",
				Sources: cli.EnvVars("ANNOTATION_API_SERVICE_PORT"),
			},
			&cli.StringFlag{
				Name:     "tag",
				Usage:    "Tag or term name in the ontology (e.g. GO:0005634)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "identifier",
				Aliases:  []string{"i"},
				Usage:    "Identifier that will be searched for annotation",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "ontology",
				Usage:    "Ontology name (e.g. cellular_component, biological_process)",
				Required: true,
			},
		},
		Action: client.RemoveAnnotation,
	}
}
