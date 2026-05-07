// Package client implements the gRPC client logic for interacting with the
// stock service.
package client

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	eq "github.com/IBM/fp-go/v2/eq"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/predicate"
	T "github.com/IBM/fp-go/v2/tuple"
	stockpb "github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/grpc-client/internal/domain"
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

var (
	strEq = eq.Equals(eq.FromStrictEquals[string]())

	isNotFoundError = F.Pipe1(
		P.IsStrictEqual[codes.Code]()(codes.NotFound),
		P.ContraMap(func(err error) codes.Code {
			return status.Code(err)
		}),
	)
)

// createConnection creates a gRPC connection
func createConnection(
	cfg StockConfig,
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
	cfg StockConfig,
) IOE.IOEither[error, StockWithConnection] {
	return F.Pipe1(
		createConnection(cfg),
		IOE.Map[error](func(conn *grpc.ClientConn) StockWithConnection {
			return StockWithConnection{
				StockConfig: cfg,
				Connection:  conn,
			}
		}),
	)
}

// ListPlasmids implements the main pipeline for listing plasmids
// It serves as the CLI Action runner
func ListPlasmids(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		IOE.Of[error](StockConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter:     cmd.String("filter"),
			Limit:      DefaultPlasmidLimit,
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[StockConfig](
				"Starting plasmid listing: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[StockWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListPlasmids),
		domain.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, *stockpb.PlasmidCollection] {
				return T.MakeTuple2(err, (*stockpb.PlasmidCollection)(nil))
			},
			func(coll *stockpb.PlasmidCollection) T.Tuple2[error, *stockpb.PlasmidCollection] {
				return T.MakeTuple2[error](nil, coll)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	printPlasmidResults(ToPlasmidResults(result.F2))

	return nil
}

// LookupPlasmidByName looks up a plasmid by exact name using the plasmid_name filter.
func LookupPlasmidByName(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		IOE.Of[error](StockConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter:     fmt.Sprintf("plasmid_name===%s", cmd.String("name")),
			Limit:      int64(cmd.Int("limit")),
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[StockConfig](
				"Looking up plasmid by name: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[StockWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListPlasmids),
		domain.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, *stockpb.PlasmidCollection] {
				return T.MakeTuple2(err, (*stockpb.PlasmidCollection)(nil))
			},
			func(coll *stockpb.PlasmidCollection) T.Tuple2[error, *stockpb.PlasmidCollection] {
				return T.MakeTuple2[error](nil, coll)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	printPlasmidResults(ToPlasmidResults(result.F2))

	return nil
}

// FetchPlasmid connects to the gRPC stock service and fetches a single plasmid by ID.
func FetchPlasmid(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe7(
		IOE.Of[error](StockConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			PlasmidID:  cmd.String("identifier"),
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[StockConfig](
				"Fetching plasmid by ID: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[StockWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callGetPlasmid),
		IOE.Map[error](func(pdata *stockpb.Plasmid) string {
			return fmt.Sprintf(
				"%s %s %s %s",
				pdata.GetData().GetId(),
				pdata.GetData().GetAttributes().GetName(),
				pdata.GetData().GetAttributes().GetCreatedBy(),
				pdata.GetData().GetAttributes().GetSummary(),
			)
		}),
		domain.ToEither,
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
	result := F.Pipe7(
		IOE.Of[error](StockConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			StrainID:   cmd.String("identifier"),
		}),
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
				s.GetData().GetId(),
				s.GetData().GetAttributes().GetLabel(),
				s.GetData().GetAttributes().GetCreatedBy(),
				s.GetData().GetAttributes().GetPublications(),
				s.GetData().GetAttributes().GetSpecies(),
				s.GetData().GetAttributes().GetGenes(),
			)
		}),
		domain.ToEither,
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

// FilterStrain filters strains by type and prints the results.
func FilterStrain(_ context.Context, cmd *cli.Command) error {
	stype := cmd.String("strain-type")
	result := F.Pipe7(
		IOE.Of[error](StockConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter:     buildStrainFilter(stype),
			StrainType: stype,
			Limit:      int64(cmd.Int("limit")),
			Cursor:     int64(cmd.Int("cursor")),
		}),
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
		domain.ToEither,
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

	if len(result.F2.GetData()) == 0 {
		return fmt.Errorf("no strain found with filter %s", buildStrainFilter(stype))
	}

	results := ToStrainResults(result.F2)
	nextCursor := int64(0)

	if result.F2.GetMeta() != nil {
		nextCursor = result.F2.GetMeta().GetNextCursor()
	}

	printStrainResults(results, nextCursor)

	return nil
}

// ListAllPlasmids implements the main pipeline for listing all plasmids paginated
// It serves as the CLI Action runner
func ListAllPlasmids(_ context.Context, cmd *cli.Command) error {
	either := F.Pipe5(
		IOE.Of[error](StockConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter:     "",
			Limit:      BatchFetchLimit,
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[StockConfig](
				"Starting paginated plasmid listing: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[StockWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListPlasmidsLoop),
		domain.ToEither,
	)
	result := E.Fold(
		func(err error) T.Tuple2[error, []domain.PlasmidResult] {
			return T.MakeTuple2(err, []domain.PlasmidResult(nil))
		},
		func(data []domain.PlasmidResult) T.Tuple2[error, []domain.PlasmidResult] {
			return T.MakeTuple2[error](nil, data)
		},
	)(either)

	if result.F1 != nil {
		return result.F1
	}

	printTop10PlasmidResults(result.F2)

	return nil
}
