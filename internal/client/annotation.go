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
	"github.com/dictyBase/grpc-client/internal/domain"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DefaultAnnotationLimit      = 10
	DefaultAnnotationGroupLimit = 100
)

// annotationCollection is a type alias for the tagged annotation collection tuple
type annotationCollection = *annotationpb.TaggedAnnotationCollection

// annotationGroupCollection is a type alias for the tagged annotation group collection tuple
type annotationGroupCollection = *annotationpb.TaggedAnnotationGroupCollection

// AnnotationConfig holds configuration for fetching annotations from the Annotation API
type AnnotationConfig struct {
	ServerAddr string
	Port       string
	Filter     string
	Limit      int64
	Cursor     int64
	Tag        string
	Identifier string
	Ontology   string
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
		collection.GetData(),
		A.Map(func(d *annotationpb.TaggedAnnotationCollection_Data) domain.AnnotationResult {
			return domain.AnnotationResult{
				ID:        d.GetId(),
				EntryID:   d.GetAttributes().GetEntryId(),
				Tag:       d.GetAttributes().GetTag(),
				Ontology:  d.GetAttributes().GetOntology(),
				Value:     d.GetAttributes().GetValue(),
				CreatedBy: d.GetAttributes().GetCreatedBy(),
				Version:   d.GetAttributes().GetVersion(),
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
		fmt.Println(domain.FormatAnnotationRecord(a))
	}

	fmt.Printf("next-cursor:%d\n", nextCursor)
}

// callListAnnotationGroups executes gRPC ListAnnotationGroups call using enriched context
func callListAnnotationGroups(
	ctx annotationWithConnection,
) IOE.IOEither[error, *annotationpb.TaggedAnnotationGroupCollection] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*annotationpb.TaggedAnnotationGroupCollection, error) {
			defer ctx.Connection.Close()

			return annotationpb.NewTaggedAnnotationServiceClient(ctx.Connection).
				ListAnnotationGroups(context.Background(),
					&annotationpb.ListGroupParameters{
						Limit:  ctx.Limit,
						Filter: ctx.Filter,
						Cursor: ctx.Cursor,
					})
		}),
		IOE.MapLeft[*annotationpb.TaggedAnnotationGroupCollection](
			fperrors.OnError("failed to list annotation groups"),
		),
	)
}

// ToAnnotationGroupResults converts protobuf group collection to domain results
func ToAnnotationGroupResults(
	collection *annotationpb.TaggedAnnotationGroupCollection,
) []domain.AnnotationGroupResult {
	return F.Pipe1(
		collection.GetData(),
		A.Map(
			func(d *annotationpb.TaggedAnnotationGroupCollection_Data) domain.AnnotationGroupResult {
				return domain.AnnotationGroupResult{
					GroupID: d.GetGroup().GetGroupId(),
					Annotations: F.Pipe1(
						d.GetGroup().GetData(),
						A.Map(func(
							a *annotationpb.TaggedAnnotationGroup_Data,
						) domain.AnnotationResult {
							return domain.AnnotationResult{
								ID:        a.GetId(),
								EntryID:   a.GetAttributes().GetEntryId(),
								Tag:       a.GetAttributes().GetTag(),
								Ontology:  a.GetAttributes().GetOntology(),
								Value:     a.GetAttributes().GetValue(),
								CreatedBy: a.GetAttributes().GetCreatedBy(),
								Version:   a.GetAttributes().GetVersion(),
							}
						}),
					),
				}
			},
		),
	)
}

// printAnnotationGroupResults prints the annotation group results to stdout.
func printAnnotationGroupResults(
	results []domain.AnnotationGroupResult,
	nextCursor int64,
) {
	fmt.Printf("total groups %d\n", len(results))

	for _, g := range results {
		fmt.Print(domain.FormatAnnotationGroupRecord(g))
	}

	fmt.Printf("next-cursor:%d\n", nextCursor)
}

// buildAnnotationGroupFilter constructs the filter string for ListAnnotationGroups.
func buildAnnotationGroupFilter(identifier, tag, ontology string) string {
	filter := fmt.Sprintf("entry_id===%s", identifier)

	if tag != "" {
		filter += fmt.Sprintf(";tag===%s", tag)
	}

	if ontology != "" {
		filter += fmt.Sprintf(";ontology===%s", ontology)
	}

	return filter
}

// FindAnnotation lists annotations matching a filter and prints the results.
func FindAnnotation(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		IOE.Of[error](AnnotationConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter:     cmd.String("filter"),
			Limit:      int64(cmd.Int("limit")),
			Cursor:     int64(cmd.Int("cursor")),
		}),
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
		domain.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, annotationCollection] {
				return T.MakeTuple2(err, annotationCollection(nil))
			},
			func(coll annotationCollection) T.Tuple2[error, annotationCollection] {
				return T.MakeTuple2[error](nil, coll)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	results := ToAnnotationResults(result.F2)
	nextCursor := int64(0)

	if result.F2.GetMeta() != nil {
		nextCursor = result.F2.GetMeta().GetNextCursor()
	}

	printAnnotationResults(results, nextCursor)

	return nil
}

// FindByTag lists annotations filtered by tag and ontology and prints the results.
func FindByTag(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		IOE.Of[error](AnnotationConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter: fmt.Sprintf(
				"tag===%s;ontology===%s",
				cmd.String("tag"),
				cmd.String("ontology"),
			),
			Limit:  int64(cmd.Int("limit")),
			Cursor: int64(cmd.Int("cursor")),
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[AnnotationConfig](
				"Finding annotations by tag: %+v",
			),
		),
		IOE.Chain(createAnnotationWithConnection),
		IOE.MapLeft[annotationWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListAnnotations),
		domain.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, *annotationpb.TaggedAnnotationCollection] {
				return T.MakeTuple2(
					err,
					(*annotationpb.TaggedAnnotationCollection)(nil),
				)
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

	if result.F2.GetMeta() != nil {
		nextCursor = result.F2.GetMeta().GetNextCursor()
	}

	printAnnotationResults(results, nextCursor)

	return nil
}

// FindAnnotationGroup retrieves annotation groups filtered by identifier, tag, and ontology.
func FindAnnotationGroup(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		IOE.Of[error](AnnotationConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter: buildAnnotationGroupFilter(
				cmd.String("identifier"),
				cmd.String("tag"),
				cmd.String("ontology"),
			),
			Limit:  int64(cmd.Int("limit")),
			Cursor: int64(cmd.Int("cursor")),
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[AnnotationConfig](
				"Finding annotation groups: %+v",
			),
		),
		IOE.Chain(createAnnotationWithConnection),
		IOE.MapLeft[annotationWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callListAnnotationGroups),
		domain.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, annotationGroupCollection] {
				return T.MakeTuple2(err, annotationGroupCollection(nil))
			},
			func(coll annotationGroupCollection) T.Tuple2[error, annotationGroupCollection] {
				return T.MakeTuple2[error](nil, coll)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	results := ToAnnotationGroupResults(result.F2)
	nextCursor := int64(0)

	if result.F2.GetMeta() != nil {
		nextCursor = result.F2.GetMeta().GetNextCursor()
	}

	printAnnotationGroupResults(results, nextCursor)

	return nil
}

// callRemoveAnnotation looks up an annotation by tag/identifier/ontology, then deletes it.
func callRemoveAnnotation(
	ctx annotationWithConnection,
) IOE.IOEither[error, *annotationpb.TaggedAnnotation] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*annotationpb.TaggedAnnotation, error) {
			defer ctx.Connection.Close()

			client := annotationpb.NewTaggedAnnotationServiceClient(ctx.Connection)

			ta, err := client.GetEntryAnnotation(context.Background(),
				&annotationpb.EntryAnnotationRequest{
					Tag:      ctx.Tag,
					EntryId:  ctx.Identifier,
					Ontology: ctx.Ontology,
				})
			if err != nil {
				return nil, fperrors.OnError("failed to get entry annotation")(err)
			}

			_, err = client.DeleteAnnotation(context.Background(),
				&annotationpb.DeleteAnnotationRequest{
					Id:    ta.GetData().GetId(),
					Purge: true,
				})
			if err != nil {
				return nil, fperrors.OnError("failed to delete annotation")(err)
			}

			return ta, nil
		}),
		IOE.MapLeft[*annotationpb.TaggedAnnotation](
			fperrors.OnError("error in removing annotation"),
		),
	)
}

// printRemovedAnnotation prints the details of a deleted annotation.
func printRemovedAnnotation(ta *annotationpb.TaggedAnnotation) {
	fmt.Printf(
		"deleted tag=> %s ontology=> %s entry=> %s value=> %s rank=> %d\n",
		ta.GetData().GetAttributes().GetTag(),
		ta.GetData().GetAttributes().GetOntology(),
		ta.GetData().GetAttributes().GetEntryId(),
		ta.GetData().GetAttributes().GetValue(),
		ta.GetData().GetAttributes().GetRank(),
	)
}

// RemoveAnnotation looks up and deletes an annotation by tag, identifier, and ontology.
func RemoveAnnotation(_ context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		IOE.Of[error](AnnotationConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Tag:        cmd.String("tag"),
			Identifier: cmd.String("identifier"),
			Ontology:   cmd.String("ontology"),
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[AnnotationConfig](
				"Removing annotation: %+v",
			),
		),
		IOE.Chain(createAnnotationWithConnection),
		IOE.MapLeft[annotationWithConnection](
			fperrors.OnError("failed to create connection"),
		),
		IOE.Chain(callRemoveAnnotation),
		domain.ToEither,
		E.Fold(
			func(err error) T.Tuple2[error, *annotationpb.TaggedAnnotation] {
				return T.MakeTuple2(err, (*annotationpb.TaggedAnnotation)(nil))
			},
			func(ta *annotationpb.TaggedAnnotation) T.Tuple2[error, *annotationpb.TaggedAnnotation] {
				return T.MakeTuple2[error](nil, ta)
			},
		),
	)

	if result.F1 != nil {
		return result.F1
	}

	printRemovedAnnotation(result.F2)

	return nil
}
