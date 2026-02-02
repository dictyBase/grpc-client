package main

import (
	"context"
	"log"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/client"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/fputil"
	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/types"
	"github.com/urfave/cli/v3"
)

// RunListPlasmids implements the main pipeline for listing plasmids
func RunListPlasmids(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		IOE.Of[error](types.ListPlasmidsConfig{
			ServerAddr: cmd.String("host"),
			Port:       cmd.String("port"),
			Filter:     "summary=~GoldenBraid",
		}),
		IOE.ChainFirstIOK[error](
			IO.Logf[types.ListPlasmidsConfig](
				"Starting plasmid listing: %+v",
			),
		),
		IOE.Chain(client.ListPlasmids),
		IOE.Map[error](client.ToPlasmidResults),
		fputil.ToEither[error, []string],
		E.Fold(
			func(err error) error { return err },  // Left: return error
			func(_ []string) error { return nil }, // Right: success
		),
	)
}

func main() {
	app := &cli.Command{
		Name:  "goldenbraid-list",
		Usage: "List GoldenBraid plasmids from stock API",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "gRPC server host address",
				Value: "localhost",
			},
			&cli.StringFlag{
				Name:  "port",
				Usage: "gRPC server port",
				Value: "9560",
			},
		},
		Action: RunListPlasmids,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
