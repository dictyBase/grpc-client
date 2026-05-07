package client

import (
	"context"
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	stockpb "github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/grpc-client/internal/domain"
)

// callListPlasmids executes gRPC ListPlasmids call using enriched context
func callListPlasmids(
	ctx StockWithConnection,
) IOE.IOEither[error, *stockpb.PlasmidCollection] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*stockpb.PlasmidCollection, error) {
			defer ctx.Connection.Close()

			return stockpb.NewStockServiceClient(ctx.Connection).
				ListPlasmids(context.Background(),
					&stockpb.StockParameters{
						Limit:  ctx.Limit,
						Filter: ctx.Filter,
					})
		}),
		IOE.MapLeft[*stockpb.PlasmidCollection](
			fperrors.OnError("failed to list plasmids"),
		),
	)
}

// wrapFetchPlasmidError returns an error mapping function that
// produces a user-friendly message for NotFound errors and wraps others.
func wrapFetchPlasmidError(plasmidID string) func(error) error {
	return func(err error) error {
		return F.Pipe2(
			err,
			O.FromPredicate(isNotFoundError),
			O.Fold(
				F.Constant(
					F.Pipe1(
						err,
						fperrors.OnError("error fetching plasmid"),
					),
				),
				F.Constant1[error](
					fmt.Errorf(
						"plasmid with identifier %s not found",
						plasmidID,
					),
				),
			),
		)
	}
}

// callGetPlasmid executes gRPC GetPlasmid call using enriched context
func callGetPlasmid(
	ctx StockWithConnection,
) IOE.IOEither[error, *stockpb.Plasmid] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*stockpb.Plasmid, error) {
			defer ctx.Connection.Close()

			return stockpb.NewStockServiceClient(ctx.Connection).
				GetPlasmid(context.Background(),
					&stockpb.StockId{Id: ctx.PlasmidID})
		}),
		IOE.MapLeft[*stockpb.Plasmid](
			wrapFetchPlasmidError(ctx.PlasmidID),
		),
	)
}

// ToPlasmidResults converts protobuf collection to domain results
func ToPlasmidResults(
	collection *stockpb.PlasmidCollection,
) []domain.PlasmidResult {
	return F.Pipe1(
		collection.GetData(),
		A.Map(func(pdata *stockpb.PlasmidCollection_Data) domain.PlasmidResult {
			return domain.PlasmidResult{
				ID:      pdata.GetId(),
				Name:    pdata.GetAttributes().GetName(),
				Summary: pdata.GetAttributes().GetSummary(),
			}
		}),
	)
}

// printPlasmidResults prints the plasmid results to stdout.
func printPlasmidResults(results []domain.PlasmidResult) {
	lines := F.Pipe1(results, A.Map(domain.FormatPlasmidRecord))
	fmt.Printf(">>> total %d records <<<\n", len(results))

	for _, line := range lines {
		fmt.Println(line)
	}
}

// callListPlasmidsLoop executes gRPC ListPlasmids calls in a loop using enriched context.
func callListPlasmidsLoop(
	ctx StockWithConnection,
) IOE.IOEither[error, []domain.PlasmidResult] {
	return IOE.TryCatchError(func() ([]domain.PlasmidResult, error) {
		defer ctx.Connection.Close()

		var allResults []domain.PlasmidResult

		cursor := int64(0)
		client := stockpb.NewStockServiceClient(ctx.Connection)

		for {
			coll, err := client.ListPlasmids(
				context.Background(),
				&stockpb.StockParameters{
					Limit:  ctx.Limit,
					Cursor: cursor,
				},
			)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to list plasmids batch: %w",
					err,
				)
			}

			allResults = append(allResults, ToPlasmidResults(coll)...)
			if coll.GetMeta() == nil || coll.GetMeta().GetNextCursor() == 0 {
				break
			}

			cursor = coll.GetMeta().GetNextCursor()
		}

		return allResults, nil
	})
}

// printTop10PlasmidResults prints up to the top 10 plasmid results and shows total count.
func printTop10PlasmidResults(results []domain.PlasmidResult) {
	fmt.Printf(">>> total %d records retrieved <<<\n", len(results))
	top := results

	if len(top) > TopRecordsLimit {
		top = top[:TopRecordsLimit]
	}

	lines := F.Pipe1(top, A.Map(domain.FormatPlasmidRecord))
	for _, line := range lines {
		fmt.Println(line)
	}
}

// callListPlasmidsLoop executes gRPC ListPlasmids calls in a loop using enriched context.
