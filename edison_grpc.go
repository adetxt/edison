package edison

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (ed *Edison) EnableGRPC(opts ...grpc.ServerOption) {
	ed.grpcServer = grpc.NewServer(opts...)
	ed.grpcEnabled = true
	ed.gwmux = runtime.NewServeMux()
}

func (ed *Edison) RegisterGRPCGateway(f GRPCGatewayHandler) {
	ed.grpcGateways = append(ed.grpcGateways, f)
}

func (ed *Edison) GRPCServer() *grpc.Server {
	return ed.grpcServer
}

func (ed *Edison) StartGRPCServer(grpcPort string, restPort string) {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Serve gRPC Server
	log.Printf("Serving gRPC Server on 0.0.0.0:%s\n", grpcPort)
	go func() {
		log.Fatalln(ed.grpcServer.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("0.0.0.0:%s", grpcPort),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	for _, f := range ed.grpcGateways {
		f(context.Background(), ed.gwmux, conn)
	}

	ed.ec.Use(
		echo.WrapMiddleware(func(h http.Handler) http.Handler {
			return ed.gwmux
		}),
	)

	ed.RestRouter("GET", "/__health", func(ctx context.Context, c RestContext) error {
		return c.EchoContext.String(http.StatusOK, "ok")
	})

	log.Printf("Serving REST Server on http://0.0.0.0:%s\n", restPort)
	if err := ed.ec.Start(fmt.Sprintf(":%s", restPort)); err != nil && err != http.ErrServerClosed {
		ed.ec.Logger.Fatal("shutting down the server")
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ed.ec.Shutdown(ctx); err != nil {
		ed.ec.Logger.Fatal(err)
	}
}
