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
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
)

// wrapFetchStrainError returns an error mapping function that
// produces a user-friendly message for NotFound errors and wraps others.
func wrapFetchStrainError(strainID string) func(error) error {
	return func(err error) error {
		return F.Pipe2(
			err,
			O.FromPredicate(isNotFoundError),
			O.Fold(
				F.Constant(
					F.Pipe1(
						err,
						fperrors.OnError("error fetching strain"),
					),
				),
				F.Constant1[error](
					fmt.Errorf(
						"strain with identifier %s not found",
						strainID,
					),
				),
			),
		)
	}
}

// callGetStrain executes gRPC GetStrain call using enriched context
func callGetStrain(
	ctx StockWithConnection,
) IOE.IOEither[error, *stockpb.Strain] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*stockpb.Strain, error) {
			defer ctx.Connection.Close()
			return stockpb.NewStockServiceClient(ctx.Connection).
				GetStrain(context.Background(),
					&stockpb.StockId{Id: ctx.StrainID})
		}),
		IOE.MapLeft[*stockpb.Strain](
			wrapFetchStrainError(ctx.StrainID),
		),
	)
}

// runFetchStrain executes the full pipeline for fetching a single strain by ID.
func runFetchStrain(cfg StockConfig) error {
	result := F.Pipe7(
		IOE.Of[error](cfg),
		IOE.ChainFirstIOK[error](
			IO.Logf[StockConfig](
				"Fetching strain by ID: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[StockWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callGetStrain),
		IOE.Map[error](func(s *stockpb.Strain) string {
			return fmt.Sprintf(
				"%s %s %s %s %s %s",
				s.Data.Id,
				s.Data.Attributes.Label,
				s.Data.Attributes.CreatedBy,
				s.Data.Attributes.Publications,
				s.Data.Attributes.Species,
				s.Data.Attributes.Genes,
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

// buildStrainFilter builds the filter string for listing strains by type.
func buildStrainFilter(stype string) string {
	filter := "ontology==dicty_strain_property"
	if stype == "all" {
		return fmt.Sprintf("%s;tag==%s,tag==%s,tag==%s",
			filter,
			aggregation.StrainFilterAllowed[0],
			aggregation.StrainFilterAllowed[1],
			aggregation.StrainFilterAllowed[2],
		)
	}
	return fmt.Sprintf("%s;tag==%s", filter, stype)
}

// callListStrains executes gRPC ListStrains call using enriched context
func callListStrains(
	ctx StockWithConnection,
) IOE.IOEither[error, *stockpb.StrainCollection] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*stockpb.StrainCollection, error) {
			defer ctx.Connection.Close()
			return stockpb.NewStockServiceClient(ctx.Connection).
				ListStrains(context.Background(),
					&stockpb.StockParameters{
						Limit:  ctx.Limit,
						Filter: ctx.Filter,
						Cursor: ctx.Cursor,
					})
		}),
		IOE.MapLeft[*stockpb.StrainCollection](
			fperrors.OnError("failed to list strains"),
		),
	)
}

// ToStrainResults converts protobuf collection to domain results
func ToStrainResults(collection *stockpb.StrainCollection) []aggregation.StrainResult {
	return F.Pipe1(
		collection.Data,
		A.Map(func(s *stockpb.StrainCollection_Data) aggregation.StrainResult {
			return aggregation.StrainResult{
				ID:                  s.Id,
				Label:               s.Attributes.GetLabel(),
				CreatedBy:           s.Attributes.GetCreatedBy(),
				Species:             s.Attributes.GetSpecies(),
				DictyStrainProperty: s.Attributes.GetDictyStrainProperty(),
			}
		}),
	)
}

// printStrainResults prints the strain results to stdout.
func printStrainResults(results []aggregation.StrainResult, nextCursor int64) {
	fmt.Printf("total strain fetched %d\n", len(results))
	for _, s := range results {
		fmt.Printf(
			"%s %s %s %s %s\n",
			s.ID,
			s.Label,
			s.CreatedBy,
			s.Species,
			s.DictyStrainProperty,
		)
	}
	fmt.Printf("next-cursor:%d\n", nextCursor)
}

func isAllowedStrainType(cfg StockConfig) bool {
	return F.Pipe1(
		aggregation.StrainFilterAllowed,
		A.Any(strEq(cfg.StrainType)),
	)
}

func strainTypeValidation(cfg StockConfig) IOE.IOEither[error, StockConfig] {
	return F.Pipe2(
		cfg,
		O.FromPredicate(isAllowedStrainType),
		IOE.FromOption[StockConfig](func() error {
			return fmt.Errorf(
				"strain type %s is not allowed",
				cfg.StrainType,
			)
		}),
	)
}

// runFilterStrain executes the full pipeline for filtering strains by type.
func runFilterStrain(cfg StockConfig) error {
	result := F.Pipe7(
		IOE.Of[error](cfg),
		IOE.Chain(strainTypeValidation),
		IOE.ChainFirstIOK[error](
			IO.Logf[StockConfig](
				"Starting strain filtering: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[StockWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListStrains),
		fputil.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, *stockpb.StrainCollection] {
				return T.MakeTuple2(err, (*stockpb.StrainCollection)(nil))
			},
			func(coll *stockpb.StrainCollection) T.Tuple2[error, *stockpb.StrainCollection] {
				return T.MakeTuple2[error](nil, coll)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	if len(result.F2.Data) == 0 {
		return fmt.Errorf("no strain found with filter %s", cfg.Filter)
	}

	results := ToStrainResults(result.F2)
	nextCursor := int64(0)
	if result.F2.Meta != nil {
		nextCursor = result.F2.Meta.NextCursor
	}
	printStrainResults(results, nextCursor)

	return nil
}
