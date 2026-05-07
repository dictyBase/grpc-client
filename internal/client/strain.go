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

// buildStrainFilter builds the filter string for listing strains by type.
func buildStrainFilter(stype string) string {
	filter := "ontology==dicty_strain_property"
	if stype == "all" {
		return fmt.Sprintf("%s;tag==%s,tag==%s,tag==%s",
			filter,
			domain.StrainFilterAllowed[0],
			domain.StrainFilterAllowed[1],
			domain.StrainFilterAllowed[2],
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
func ToStrainResults(collection *stockpb.StrainCollection) []domain.StrainResult {
	return F.Pipe1(
		collection.GetData(),
		A.Map(func(s *stockpb.StrainCollection_Data) domain.StrainResult {
			return domain.StrainResult{
				ID:                  s.GetId(),
				Label:               s.GetAttributes().GetLabel(),
				CreatedBy:           s.GetAttributes().GetCreatedBy(),
				Species:             s.GetAttributes().GetSpecies(),
				DictyStrainProperty: s.GetAttributes().GetDictyStrainProperty(),
			}
		}),
	)
}

// printStrainResults prints the strain results to stdout.
func printStrainResults(results []domain.StrainResult, nextCursor int64) {
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
		domain.StrainFilterAllowed,
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
