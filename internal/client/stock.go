// Package client implements the gRPC client logic for interacting with the
// stock service.
package client

import (
	"context"
	"fmt"

	eq "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/domain"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	DefaultPlasmidLimit      = 1000
	DefaultLookupLimit       = 3
	TopRecordsLimit          = 10
	BatchFetchLimit          = 30
	DefaultStrainFilterLimit = 10
)

var strEq = eq.Equals(eq.FromStrictEquals[string]())

// createConnection creates a gRPC connection
func createConnection(
	cfg domain.ListPlasmidsConfig,
) IOE.IOEither[error, *grpc.ClientConn] {
	return IOE.TryCatchError(func() (*grpc.ClientConn, error) {
		return grpc.NewClient(
			fmt.Sprintf("%s:%s", cfg.ServerAddr, cfg.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	})
}

// createWithConnection enriches config with a gRPC connection.
func createWithConnection(
	cfg domain.ListPlasmidsConfig,
) IOE.IOEither[error, domain.WithConnection] {
	return F.Pipe1(
		createConnection(cfg),
		IOE.Map[error](func(conn *grpc.ClientConn) domain.WithConnection {
			return domain.WithConnection{
				ListPlasmidsConfig: cfg,
				Connection:         conn,
			}
		}),
	)
}

// isNotFoundError checks if the given error is a gRPC NotFound error.
func isNotFoundError(err error) bool {
	return status.Code(err) == codes.NotFound
}

// ListPlasmids implements the main pipeline for listing plasmids
// It serves as the CLI Action runner
func ListPlasmids(_ context.Context, cmd *cli.Command) error {
	return runPlasmidList(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		Filter:     cmd.String("filter"),
		Limit:      DefaultPlasmidLimit,
	})
}

// LookupPlasmidByName looks up a plasmid by exact name using the plasmid_name filter.
func LookupPlasmidByName(_ context.Context, cmd *cli.Command) error {
	return runPlasmidList(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		Filter:     fmt.Sprintf("plasmid_name===%s", cmd.String("name")),
		Limit:      int64(cmd.Int("limit")),
	})
}

// FetchPlasmid connects to the gRPC stock service and fetches a single plasmid by ID.
func FetchPlasmid(_ context.Context, cmd *cli.Command) error {
	return runFetchPlasmid(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		PlasmidID:  cmd.String("identifier"),
	})
}

// FetchStrain connects to the gRPC stock service and fetches a single strain by ID.
func FetchStrain(_ context.Context, cmd *cli.Command) error {
	return runFetchStrain(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		StrainID:   cmd.String("identifier"),
	})
}

// FilterStrain filters strains by type and prints the results.
func FilterStrain(_ context.Context, cmd *cli.Command) error {
	stype := cmd.String("strain-type")
	return runFilterStrain(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		Filter:     buildStrainFilter(stype),
		StrainType: stype,
		Limit:      int64(cmd.Int("limit")),
		Cursor:     int64(cmd.Int("cursor")),
	})
}

// ListAllPlasmids implements the main pipeline for listing all plasmids paginated
// It serves as the CLI Action runner
func ListAllPlasmids(_ context.Context, cmd *cli.Command) error {
	return runAllPlasmidList(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		Filter:     "",
		Limit:      BatchFetchLimit,
	})
}

// FindAnnotation lists annotations matching a filter and prints the results.
func FindAnnotation(_ context.Context, cmd *cli.Command) error {
	return runFindAnnotation(AnnotationConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		Filter:     cmd.String("filter"),
		Limit:      int64(cmd.Int("limit")),
		Cursor:     int64(cmd.Int("cursor")),
	})
}
