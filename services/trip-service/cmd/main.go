package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/AuraReaper/voom/services/trip-service/internal/infrastructure/events"
	"github.com/AuraReaper/voom/services/trip-service/internal/infrastructure/grpc"
	"github.com/AuraReaper/voom/services/trip-service/internal/infrastructure/repository"
	"github.com/AuraReaper/voom/services/trip-service/internal/service"
	"github.com/AuraReaper/voom/shared/env"
	"github.com/AuraReaper/voom/shared/messaging"
	"github.com/AuraReaper/voom/shared/tracing"
	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:  "trip-service",
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

	repo := repository.NewInmemRepository()
	svc := service.NewTripService(repo)

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMq Connection")

	publisher := events.NewTripEventPublisher(rabbitmq)

	// Start driver consumer
	driverConsumer := events.NewDriverConsumer(rabbitmq, svc)
	go driverConsumer.Listen()

	// Start payment consumer
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	go paymentConsumer.Listen()

	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpc.NewGRPCHandler(grpcServer, svc, publisher)
	log.Printf("Starting gRPC server Trip Service on port: %s", lis.Addr().String())

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
