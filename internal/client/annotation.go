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
	annotationpb "github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/aggregation"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/domain"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const DefaultAnnotationLimit = 10

// AnnotationConfig holds configuration for fetching annotations from the Annotation API
type AnnotationConfig struct {
	ServerAddr string
	Port       string
	Filter     string
	Limit      int64
	Cursor     int64
}

// annotationWithConnection enriches AnnotationConfig with a gRPC connection
type annotationWithConnection struct {
	AnnotationConfig
	Connection *grpc.ClientConn
}

// createAnnotationConnection creates a gRPC connection for the annotation API
func createAnnotationConnection(
	cfg AnnotationConfig,
) IOE.IOEither[error, *grpc.ClientConn] {
	return IOE.TryCatchError(func() (*grpc.ClientConn, error) {
		return grpc.NewClient(
			fmt.Sprintf("%s:%s", cfg.ServerAddr, cfg.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	})
}

// createAnnotationWithConnection enriches annotation config with a gRPC connection.
func createAnnotationWithConnection(
	cfg AnnotationConfig,
) IOE.IOEither[error, annotationWithConnection] {
	return F.Pipe1(
		createAnnotationConnection(cfg),
		IOE.Map[error](func(conn *grpc.ClientConn) annotationWithConnection {
			return annotationWithConnection{
				AnnotationConfig: cfg,
				Connection:       conn,
			}
		}),
	)
}

// callListAnnotations executes gRPC ListAnnotations call using enriched context
func callListAnnotations(
	ctx annotationWithConnection,
) IOE.IOEither[error, *annotationpb.TaggedAnnotationCollection] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*annotationpb.TaggedAnnotationCollection, error) {
			defer ctx.Connection.Close()
			return annotationpb.NewTaggedAnnotationServiceClient(ctx.Connection).
				ListAnnotations(context.Background(),
					&annotationpb.ListParameters{
						Limit:  ctx.Limit,
						Filter: ctx.Filter,
						Cursor: ctx.Cursor,
					})
		}),
		IOE.MapLeft[*annotationpb.TaggedAnnotationCollection](
			fperrors.OnError("failed to list annotations"),
		),
	)
}

// ToAnnotationResults converts protobuf collection to domain results
func ToAnnotationResults(
	collection *annotationpb.TaggedAnnotationCollection,
) []domain.AnnotationResult {
	return F.Pipe1(
		collection.Data,
		A.Map(func(d *annotationpb.TaggedAnnotationCollection_Data) domain.AnnotationResult {
			return domain.AnnotationResult{
				ID:        d.Id,
				EntryID:   d.Attributes.GetEntryId(),
				Tag:       d.Attributes.GetTag(),
				Ontology:  d.Attributes.GetOntology(),
				Value:     d.Attributes.GetValue(),
				CreatedBy: d.Attributes.GetCreatedBy(),
				Version:   d.Attributes.GetVersion(),
			}
		}),
	)
}

// printAnnotationResults prints the annotation results to stdout.
func printAnnotationResults(
	results []domain.AnnotationResult,
	nextCursor int64,
) {
	fmt.Printf("total annotations fetched %d\n", len(results))
	for _, a := range results {
		fmt.Println(aggregation.FormatAnnotationRecord(a))
	}
	fmt.Printf("next-cursor:%d\n", nextCursor)
}

// runFindAnnotation executes the full pipeline for finding annotations by filter.
func runFindAnnotation(cfg AnnotationConfig) error {
	result := F.Pipe6(
		IOE.Of[error](cfg),
		IOE.ChainFirstIOK[error](
			IO.Logf[AnnotationConfig](
				"Finding annotations: %+v",
			),
		),
		IOE.Chain(createAnnotationWithConnection),
		IOE.MapLeft[annotationWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListAnnotations),
		fputil.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, *annotationpb.TaggedAnnotationCollection] {
				return T.MakeTuple2(err, (*annotationpb.TaggedAnnotationCollection)(nil))
			},
			func(coll *annotationpb.TaggedAnnotationCollection) T.Tuple2[error, *annotationpb.TaggedAnnotationCollection] {
				return T.MakeTuple2[error](nil, coll)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	results := ToAnnotationResults(result.F2)
	nextCursor := int64(0)
	if result.F2.Meta != nil {
		nextCursor = result.F2.Meta.NextCursor
	}
	printAnnotationResults(results, nextCursor)

	return nil
}
