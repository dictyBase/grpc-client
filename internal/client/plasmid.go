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
	O "github.com/IBM/fp-go/v2/option"
	T "github.com/IBM/fp-go/v2/tuple"
	stockpb "github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/aggregation"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/domain"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
)

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
	ctx domain.WithConnection,
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
		collection.Data,
		A.Map(func(pdata *stockpb.PlasmidCollection_Data) domain.PlasmidResult {
			return domain.PlasmidResult{
				ID:      pdata.Id,
				Name:    pdata.Attributes.GetName(),
				Summary: pdata.Attributes.GetSummary(),
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
		fputil.ToEither,
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

// callListPlasmidsLoop executes gRPC ListPlasmids calls in a loop using enriched context.
func callListPlasmidsLoop(
	ctx domain.WithConnection,
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
			if coll.Meta == nil || coll.Meta.NextCursor == 0 {
				break
			}
			cursor = coll.Meta.NextCursor
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
	lines := F.Pipe1(top, A.Map(aggregation.FormatPlasmidRecord))
	for _, line := range lines {
		fmt.Println(line)
	}
}

// runAllPlasmidList executes the full pipeline for listing all plasmids paginated.
func runAllPlasmidList(cfg domain.ListPlasmidsConfig) error {
	result := F.Pipe6(
		IOE.Of[error](cfg),
		IOE.ChainFirstIOK[error](
			IO.Logf[domain.ListPlasmidsConfig](
				"Starting paginated plasmid listing: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[domain.WithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListPlasmidsLoop),
		fputil.ToEither,
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

	printTop10PlasmidResults(result.F2)

	return nil
}

// runFetchPlasmid executes the full pipeline for fetching a single plasmid by ID.
func runFetchPlasmid(cfg domain.ListPlasmidsConfig) error {
	result := F.Pipe7(
		IOE.Of[error](cfg),
		IOE.ChainFirstIOK[error](
			IO.Logf[domain.ListPlasmidsConfig](
				"Fetching plasmid by ID: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[domain.WithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callGetPlasmid),
		IOE.Map[error](func(pdata *stockpb.Plasmid) string {
			return fmt.Sprintf(
				"%s %s %s %s",
				pdata.Data.Id,
				pdata.Data.Attributes.Name,
				pdata.Data.Attributes.CreatedBy,
				pdata.Data.Attributes.Summary,
			)
		}),
		fputil.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, string] {
				return T.MakeTuple2(err, "")
			},
			func(data string) T.Tuple2[error, string] {
				return T.MakeTuple2[error](nil, data)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	fmt.Println(result.F2)

	return nil
}
