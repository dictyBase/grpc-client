package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	Eq "github.com/IBM/fp-go/v2/eq"
	fperrors "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	PR "github.com/IBM/fp-go/v2/pair"
	P "github.com/IBM/fp-go/v2/predicate"
	R "github.com/IBM/fp-go/v2/record"
	T "github.com/IBM/fp-go/v2/tuple"
	feature "github.com/dictyBase/go-genproto/dictybaseapis/feature_annotation"
	"github.com/dictyBase/grpc-client/internal/domain"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const keyValueParts = 2

var (
	splitByComma = F.Bind2of2(strings.Split)(",")

	splitOnEq = F.Bind23of3(strings.SplitN)("=", keyValueParts)

	sliceLen         = func(kv []string) int { return len(kv) } // ← wraps built-in
	hasKeyValueParts = F.Pipe2(
		keyValueParts,
		Eq.Equals(Eq.FromStrictEquals[int]()),
		P.ContraMap(sliceLen),
	)

	parseProperty = F.Flow2(
		F.Flow3(
			strings.TrimSpace,
			splitOnEq,
			O.FromPredicate(hasKeyValueParts),
		),
		O.Map(func(kv []string) PR.Pair[string, string] {
			return PR.MakePair(
				strings.TrimSpace(kv[0]),
				strings.TrimSpace(kv[1]),
			)
		}),
	)
)

// AnnoFeatConfig holds configuration for feature annotation service operations.
type AnnoFeatConfig struct {
	ServerAddr string
	Port       string
	ID         string
	Name       string
	Synonyms   []string
	Properties map[string]string
	CreatedBy  string
}

// annoFeatWithConnection enriches AnnoFeatConfig with a gRPC connection.
type annoFeatWithConnection struct {
	AnnoFeatConfig
	Connection *grpc.ClientConn
}

// createAnnoFeatConnection creates a gRPC connection for the feature annotation API.
func createAnnoFeatConnection(
	cfg AnnoFeatConfig,
) IOE.IOEither[error, *grpc.ClientConn] {
	return IOE.TryCatchError(func() (*grpc.ClientConn, error) {
		return grpc.NewClient(
			fmt.Sprintf("%s:%s", cfg.ServerAddr, cfg.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	})
}

// createAnnoFeatWithConnection enriches anno feat config with a gRPC connection.
func createAnnoFeatWithConnection(
	cfg AnnoFeatConfig,
) IOE.IOEither[error, annoFeatWithConnection] {
	return F.Pipe1(
		createAnnoFeatConnection(cfg),
		IOE.Map[error](func(conn *grpc.ClientConn) annoFeatWithConnection {
			return annoFeatWithConnection{
				AnnoFeatConfig: cfg,
				Connection:     conn,
			}
		}),
	)
}

// callCreateFeatureAnnotation executes gRPC CreateFeatureAnnotation call.
func callCreateFeatureAnnotation(
	ctx annoFeatWithConnection,
) IOE.IOEither[error, *feature.FeatureAnnotation] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*feature.FeatureAnnotation, error) {
			defer ctx.Connection.Close()

			return feature.NewFeatureAnnotationServiceClient(ctx.Connection).
				CreateFeatureAnnotation(context.Background(),
					buildNewFeatureAnnotation(ctx.AnnoFeatConfig),
				)
		}),
		IOE.MapLeft[*feature.FeatureAnnotation](
			fperrors.OnError("failed to create feature annotation"),
		),
	)
}

// callGetFeatureAnnotation executes gRPC GetFeatureAnnotation call.
func callGetFeatureAnnotation(
	ctx annoFeatWithConnection,
) IOE.IOEither[error, *feature.FeatureAnnotation] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*feature.FeatureAnnotation, error) {
			defer ctx.Connection.Close()

			return feature.NewFeatureAnnotationServiceClient(ctx.Connection).
				GetFeatureAnnotation(context.Background(),
					&feature.FeatureAnnotationId{Id: ctx.ID},
				)
		}),
		IOE.MapLeft[*feature.FeatureAnnotation](
			fperrors.OnError("failed to get feature annotation"),
		),
	)
}

// buildNewFeatureAnnotation constructs a NewFeatureAnnotation from config.
func buildNewFeatureAnnotation(cfg AnnoFeatConfig) *feature.NewFeatureAnnotation {
	fa := &feature.NewFeatureAnnotation{
		Id:        cfg.ID,
		CreatedBy: cfg.CreatedBy,
		CreatedAt: timestamppb.Now(),
		Attributes: &feature.FeatureAnnotationAttributes{
			Name:     cfg.Name,
			Synonyms: cfg.Synonyms,
		},
	}
	for tag, value := range cfg.Properties {
		fa.Attributes.Properties = append(fa.Attributes.Properties, &feature.TagProperty{
			Tag:       tag,
			Value:     value,
			CreatedBy: cfg.CreatedBy,
		})
	}

	return fa
}

// toAnnoFeatResult converts a FeatureAnnotation proto to a domain result.
func toAnnoFeatResult(fa *feature.FeatureAnnotation) domain.AnnoFeatResult {
	result := domain.AnnoFeatResult{
		ID:        fa.GetId(),
		Name:      fa.GetAttributes().GetName(),
		CreatedBy: fa.GetCreatedBy(),
		CreatedAt: fa.GetCreatedAt().AsTime().Format(time.RFC3339),
		Synonyms:  fa.GetAttributes().GetSynonyms(),
	}
	for _, prop := range fa.GetAttributes().GetProperties() {
		result.Properties[prop.GetTag()] = prop.GetValue()
	}

	return result
}

// printFeatureAnnotation prints a feature annotation result to stdout.
func printFeatureAnnotation(fa *feature.FeatureAnnotation) {
	fmt.Println(domain.FormatAnnoFeatRecord(toAnnoFeatResult(fa)))
}

// CreateFeatAnno creates a feature annotation via the feature annotation gRPC service.
func CreateFeatAnno(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		IOE.Of[error](AnnoFeatConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			ID:         cmd.String("id"),
			Name:       cmd.String("name"),
			Synonyms:   splitByComma(cmd.String("synonyms")),
			Properties: parseProperties(cmd.String("properties")),
			CreatedBy:  cmd.String("created-by"),
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[AnnoFeatConfig](
				"Creating feature annotation: %+v",
			),
		),
		IOE.Chain(createAnnoFeatWithConnection),
		IOE.MapLeft[annoFeatWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callCreateFeatureAnnotation),
		domain.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, *feature.FeatureAnnotation] {
				return T.MakeTuple2(err, (*feature.FeatureAnnotation)(nil))
			},
			func(fa *feature.FeatureAnnotation) T.Tuple2[error, *feature.FeatureAnnotation] {
				return T.MakeTuple2[error](nil, fa)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	printFeatureAnnotation(result.F2)

	return nil
}

// GetFeatAnno retrieves a feature annotation by ID via the feature annotation gRPC service.
func GetFeatAnno(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		IOE.Of[error](AnnoFeatConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			ID:         cmd.String("id"),
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[AnnoFeatConfig](
				"Getting feature annotation: %+v",
			),
		),
		IOE.Chain(createAnnoFeatWithConnection),
		IOE.MapLeft[annoFeatWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callGetFeatureAnnotation),
		domain.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, *feature.FeatureAnnotation] {
				return T.MakeTuple2(err, (*feature.FeatureAnnotation)(nil))
			},
			func(fa *feature.FeatureAnnotation) T.Tuple2[error, *feature.FeatureAnnotation] {
				return T.MakeTuple2[error](nil, fa)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	printFeatureAnnotation(result.F2)

	return nil
}

// parseProperties parses a comma-separated "key=value" string into a map.
func parseProperties(raw string) R.Record[string, string] {
	return F.Pipe3(
		raw,
		splitByComma,
		A.FilterMap(parseProperty),
		R.FromEntries[string, string],
	)
}
