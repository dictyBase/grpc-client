package client

import (
	"context"
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	T "github.com/IBM/fp-go/v2/tuple"
	stockpb "github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/aggregation"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/domain"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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

const DefaultPlasmidLimit = 1000

// callListPlasmids executes gRPC ListPlasmids call using enriched context
func callListPlasmids(
	ctx domain.WithConnection,
) IOE.IOEither[error, *stockpb.PlasmidCollection] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*stockpb.PlasmidCollection, error) {
			defer ctx.Connection.Close()
			return stockpb.NewStockServiceClient(ctx.Connection).
				ListPlasmids(context.Background(),
					&stockpb.StockParameters{
						Limit:  DefaultPlasmidLimit,
						Filter: ctx.Filter,
					})
		}),
		IOE.MapLeft[*stockpb.PlasmidCollection](
			fperrors.OnError("failed to list plasmids"),
		),
	)
}

// ToPlasmidResults converts protobuf collection to domain results
func ToPlasmidResults(
	collection *stockpb.PlasmidCollection,
) []domain.PlasmidResult {
	return F.Pipe1(
		collection.Data,
		A.Map(func(p *stockpb.PlasmidCollection_Data) domain.PlasmidResult {
			return domain.PlasmidResult{
				ID:      p.Id,
				Name:    p.Attributes.GetName(),
				Summary: p.Attributes.GetSummary(),
			}
		}),
	)
}

// printPlasmidResults prints the plasmid results to stdout.
func printPlasmidResults(results []domain.PlasmidResult) {
	lines := F.Pipe1(results, A.Map(aggregation.FormatPlasmidRecord))
	fmt.Printf(">>> total %d records <<<\n", len(results))
	for _, line := range lines {
		fmt.Println(line)
	}
}

// runPlasmidList executes the full pipeline for a given config and prints results.
func runPlasmidList(cfg domain.ListPlasmidsConfig) error {
	result := F.Pipe7(
		IOE.Of[error](cfg),
		IOE.ChainFirstIOK[error](
			IO.Logf[domain.ListPlasmidsConfig](
				"Starting plasmid listing: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[domain.WithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListPlasmids),
		IOE.Map[error](ToPlasmidResults),
		fputil.ToEither[error, []domain.PlasmidResult],
		E.Fold(
			func(err error) T.Tuple2[error, []domain.PlasmidResult] {
				return T.MakeTuple2(err, []domain.PlasmidResult(nil))
			},
			func(data []domain.PlasmidResult) T.Tuple2[error, []domain.PlasmidResult] {
				return T.MakeTuple2[error](nil, data)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	printPlasmidResults(result.F2)

	return nil
}

// ListPlasmids implements the main pipeline for listing plasmids
// It serves as the CLI Action runner
func ListPlasmids(_ context.Context, cmd *cli.Command) error {
	return runPlasmidList(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		Filter:     cmd.String("filter"),
	})
}

// LookupPlasmidByName looks up a plasmid by exact name using the plasmid_name filter.
func LookupPlasmidByName(_ context.Context, cmd *cli.Command) error {
	return runPlasmidList(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		Filter:     fmt.Sprintf("plasmid_name===%s", cmd.String("name")),
	})
}
