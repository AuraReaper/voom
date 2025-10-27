package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/AuraReaper/voom/services/driver-service/internal/repository"
	"github.com/AuraReaper/voom/shared/env"
	"github.com/AuraReaper/voom/shared/messaging"
	"github.com/AuraReaper/voom/shared/tracing"
	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:  "driver-service",
		Environment:  env.GetString("ENVIRONMENT", "development"),
		OTLPEndpoint: env.GetString("OTLP_ENDPOINT", "http://jaeger:4318/v1/traces"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("Failed to initialize the tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	repo := repository.NewInmemDriverRepository()
	svc := NewService(repo)

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Printf("failed to connect to rabbitmq: %v", err)
		return
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMq Connection")

	consumer := NewTripEventComsumer(rabbitmq, svc)
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("failed to listen to the message: %v", err)
		}
	}()

	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	NewGrpcHandler(grpcServer, svc)
	log.Printf("Starting gRPC server Driver Service on port: %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
			cancel()
		}
	}()

	// wait for the shutdown signal
	<-ctx.Done()
	log.Printf("shutting down the server...")
	grpcServer.GracefulStop()
}
