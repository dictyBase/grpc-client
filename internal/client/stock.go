package client

import (
	"context"
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	stockpb "github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// createConnection creates a gRPC connection
// Pure IOEither - no imperative error handling
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

// ListPlasmids fetches plasmids using IOE.Map to enrich context
func ListPlasmids(cfg types.ListPlasmidsConfig) types.CollectionIOE {
	return F.Pipe3(
		createConnection(cfg),
		IOE.MapLeft[*grpc.ClientConn](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Map[error](func(conn *grpc.ClientConn) types.WithConnection {
			return types.WithConnection{
				ListPlasmidsConfig: cfg,
				Connection:         conn,
			}
		}),
		IOE.Chain(callListPlasmids),
	)
}

// ToPlasmidResults converts protobuf collection to domain results
// Uses A.Map for array transformation (fp-go pattern)
func ToPlasmidResults(
	collection *stockpb.PlasmidCollection,
) []string {
	return F.Pipe2(
		collection.Data,
		A.Map(func(p *stockpb.PlasmidCollection_Data) types.PlasmidResult {
			return types.PlasmidResult{
				ID:      p.Id,
				Name:    p.Attributes.GetName(),
				Summary: p.Attributes.GetSummary(),
			}
		}),
		A.Map(FormatPlasmidRecord),
	)
}

// FormatPlasmidRecord formats a single plasmid result as a display string.
// Pure function - no side effects.
func FormatPlasmidRecord(p types.PlasmidResult) string {
	summary := fputil.TruncateWords(p.Summary, 30)
	return fmt.Sprintf("ID: %s | Name: %s | Summary: %s", p.ID, p.Name, summary)
}
