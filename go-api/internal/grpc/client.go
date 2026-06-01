// Package grpc provides the gRPC client wrapper for the Rust Engine.
package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/brvm/go-api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// EngineClient wraps the gRPC connection to the Rust Engine.
type EngineClient struct {
	conn   *grpc.ClientConn
	client pb.SignalServiceClient
}

// NewEngineClient creates a new gRPC client connected to the Rust Engine at addr.
// addr should be "host:port", e.g. "localhost:50051".
func NewEngineClient(addr string) (*EngineClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to engine at %s: %w", addr, err)
	}

	return &EngineClient{
		conn:   conn,
		client: pb.NewSignalServiceClient(conn),
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *EngineClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GenerateSignal calls the Rust Engine to compute a signal for a single ticker.
func (c *EngineClient) GenerateSignal(
	ctx context.Context,
	ticker string,
	prices []float64,
	volumes []int64,
	highs []float64,
	lows []float64,
	currentPrice float64,
	fundamental *pb.FundamentalData,
	macroData *pb.MacroData,
) (*pb.SignalResponse, error) {
	req := &pb.SignalRequest{
		Ticker:       ticker,
		Prices:       prices,
		Volumes:      volumes,
		Highs:        highs,
		Lows:         lows,
		CurrentPrice: currentPrice,
		Fundamental:  fundamental,
		Macro:        macroData,
	}

	resp, err := c.client.GenerateSignal(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GenerateSignal RPC failed: %w", err)
	}

	return resp, nil
}

// ScanAllTickers calls the Rust Engine to analyze all provided tickers in parallel.
func (c *EngineClient) ScanAllTickers(
	ctx context.Context,
	tickers []*pb.TickerData,
) ([]*pb.SignalResponse, error) {
	req := &pb.ScanRequest{
		Tickers: tickers,
	}

	resp, err := c.client.ScanAllTickers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ScanAllTickers RPC failed: %w", err)
	}

	return resp.Signals, nil
}

// HealthCheck verifies the Rust Engine is running and reachable.
func (c *EngineClient) HealthCheck(ctx context.Context) (*pb.HealthResponse, error) {
	resp, err := c.client.HealthCheck(ctx, &pb.HealthRequest{})
	if err != nil {
		return nil, fmt.Errorf("HealthCheck RPC failed: %w", err)
	}

	return resp, nil
}
