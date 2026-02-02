package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dictyBase/learn-golang/grpc/plasmid/goldenbraid/internal/client"
	"github.com/urfave/cli/v3"
)

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
		Action: client.RunListPlasmids,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
