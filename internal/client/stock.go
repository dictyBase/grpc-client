// Package client implements the gRPC client logic for interacting with the
// stock service.
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
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
						Limit:  ctx.Limit,
						Filter: ctx.Filter,
					})
		}),
		IOE.MapLeft[*stockpb.PlasmidCollection](
			fperrors.OnError("failed to list plasmids"),
		),
	)
}

// isNotFoundError checks if the given error is a gRPC NotFound error.
func isNotFoundError(err error) bool {
	return status.Code(err) == codes.NotFound
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

const (
	DefaultLookupLimit       = 3
	TopRecordsLimit          = 10
	BatchFetchLimit          = 30
	DefaultStrainFilterLimit = 10
)

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
		IOE.Map[error](func(p *stockpb.Plasmid) string {
			return fmt.Sprintf(
				"%s %s %s %s",
				p.Data.Id,
				p.Data.Attributes.Name,
				p.Data.Attributes.CreatedBy,
				p.Data.Attributes.Summary,
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

// FetchPlasmid connects to the gRPC stock service and fetches a single plasmid by ID.
func FetchPlasmid(_ context.Context, cmd *cli.Command) error {
	return runFetchPlasmid(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		PlasmidID:  cmd.String("identifier"),
	})
}

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
	ctx domain.WithConnection,
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
func runFetchStrain(cfg domain.ListPlasmidsConfig) error {
	result := F.Pipe7(
		IOE.Of[error](cfg),
		IOE.ChainFirstIOK[error](
			IO.Logf[domain.ListPlasmidsConfig](
				"Fetching strain by ID: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[domain.WithConnection](
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

// FetchStrain connects to the gRPC stock service and fetches a single strain by ID.
func FetchStrain(_ context.Context, cmd *cli.Command) error {
	return runFetchStrain(domain.ListPlasmidsConfig{
		ServerAddr: cmd.String("host"),
		Port:       cmd.String("port"),
		StrainID:   cmd.String("identifier"),
	})
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
	ctx domain.WithConnection,
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
		collection.Data,
		A.Map(func(s *stockpb.StrainCollection_Data) domain.StrainResult {
			return domain.StrainResult{
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

// validateStrainType returns an Option containing the config when the strain
// type is found in the allowed list, or None otherwise.
func validateStrainType(cfg domain.ListPlasmidsConfig) O.Option[domain.ListPlasmidsConfig] {
	return F.Pipe1(
		A.Head(
			A.Filter(
				func(s string) bool { return s == cfg.StrainType },
			)(
				domain.StrainFilterAllowed,
			),
		),
		O.Map(F.Constant1[string](cfg)),
	)
}

// runFilterStrain executes the full pipeline for filtering strains by type.
func runFilterStrain(cfg domain.ListPlasmidsConfig) error {
	result := F.Pipe7(
		IOE.Of[error](cfg),
		IOE.Chain(
			func(c domain.ListPlasmidsConfig) IOE.IOEither[error, domain.ListPlasmidsConfig] {
				return F.Pipe1(
					validateStrainType(c),
					IOE.FromOption[domain.ListPlasmidsConfig](
						func() error {
							return fmt.Errorf("strain type %s is not allowed", c.StrainType)
						},
					),
				)
			},
		),
		IOE.ChainFirstIOK[error](
			IO.Logf[domain.ListPlasmidsConfig](
				"Starting strain filtering: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[domain.WithConnection](
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
