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
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/types"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// createConnection creates a gRPC connection
func createConnection(
	cfg types.ListPlasmidsConfig,
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
	cfg types.ListPlasmidsConfig,
) IOE.IOEither[error, types.WithConnection] {
	return F.Pipe1(
		createConnection(cfg),
		IOE.Map[error](func(conn *grpc.ClientConn) types.WithConnection {
			return types.WithConnection{
				ListPlasmidsConfig: cfg,
				Connection:         conn,
			}
		}),
	)
}

// callListPlasmids executes gRPC ListPlasmids call using enriched context
func callListPlasmids(
	ctx types.WithConnection,
) IOE.IOEither[error, *stockpb.PlasmidCollection] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*stockpb.PlasmidCollection, error) {
			defer ctx.Connection.Close()
			return stockpb.NewStockServiceClient(ctx.Connection).
				ListPlasmids(context.Background(),
					&stockpb.StockParameters{
						Limit:  1000,
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
) []types.PlasmidResult {
	return F.Pipe1(
		collection.Data,
		A.Map(func(p *stockpb.PlasmidCollection_Data) types.PlasmidResult {
			return types.PlasmidResult{
				ID:      p.Id,
				Name:    p.Attributes.GetName(),
				Summary: p.Attributes.GetSummary(),
			}
		}),
	)
}

// ListPlasmids implements the main pipeline for listing plasmids
// It serves as the CLI Action runner
func ListPlasmids(ctx context.Context, cmd *cli.Command) error {
	result := F.Pipe7(
		IOE.Of[error](types.ListPlasmidsConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter:     "summary=~GoldenBraid",
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[types.ListPlasmidsConfig](
				"Starting plasmid listing: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[types.WithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListPlasmids),
		IOE.Map[error](ToPlasmidResults),
		fputil.ToEither[error, []types.PlasmidResult],
		E.Fold(
			func(err error) T.Tuple2[error, []types.PlasmidResult] {
				return T.MakeTuple2(err, []types.PlasmidResult(nil))
			},
			func(data []types.PlasmidResult) T.Tuple2[error, []types.PlasmidResult] {
				return T.MakeTuple2[error](nil, data)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	for _, p := range result.F2 {
		fmt.Println(aggregation.FormatPlasmidRecord(p))
	}

	return nil
}
