package client

import (
	"google.golang.org/grpc"
)

// StockConfig holds configuration for stock service operations.
type StockConfig struct {
	ServerAddr string
	Port       string
	Filter     string
	Limit      int64
	Cursor     int64
	PlasmidID  string
	StrainID   string
	StrainType string
}

// StockWithConnection enriches StockConfig with a gRPC connection.
type StockWithConnection struct {
	StockConfig
	Connection *grpc.ClientConn
}
