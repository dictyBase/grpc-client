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
)

const DefaultAnnotationLimit = 10

// callListAnnotations executes gRPC ListAnnotations call using enriched context
func callListAnnotations(
	ctx domain.WithConnection,
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
func runFindAnnotation(cfg domain.ListPlasmidsConfig) error {
	result := F.Pipe6(
		IOE.Of[error](cfg),
		IOE.ChainFirstIOK[error](
			IO.Logf[domain.ListPlasmidsConfig](
				"Finding annotations: %+v",
			),
		),
		IOE.Chain(createWithConnection),
		IOE.MapLeft[domain.WithConnection](
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
