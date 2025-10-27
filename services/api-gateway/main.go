package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/AuraReaper/voom/services/api-gateway/internal/handlers"
	"github.com/AuraReaper/voom/shared/env"
	"github.com/AuraReaper/voom/shared/messaging"
	"github.com/AuraReaper/voom/shared/tracing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	log.Println("Starting API Gateway")
	e := echo.New()

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:  "api-gateway",
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

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	// middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// routes
	e.GET("/health", handlers.HealthCheck)
	e.POST("/trip/preview", tracing.WrapHandler(handlers.HandleTripPreview))
	e.GET("/ws/drivers", tracing.WrapHandler(func(c echo.Context) error {
		return handlers.HandleDriversWebSocket(c, rabbitmq)
	}))
	e.GET("/ws/riders", tracing.WrapHandler(func(c echo.Context) error {
		return handlers.HandleRidersWebSocket(c, rabbitmq)
	}))
	e.POST("/trip/start", tracing.WrapHandler(handlers.HandleCreateTrip))
	e.POST("/webhook/stripe", tracing.WrapHandler(func(c echo.Context) error {
		return handlers.HandleStripeWebHook(c, rabbitmq)
	}))

	// graceful shutdown handling and starting server
	serverErrors := make(chan error, 1)

	go func() {
		e.Logger.Info("Server listening on port %s", httpAddr)
		serverErrors <- e.Start(httpAddr)
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		e.Logger.Error("error starting the server: %v", err)
	case sig := <-shutdown:
		e.Logger.Info("Server is shutting down due to %v signal", sig)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Info("could not shut down the server gracefully: %v", err)
		e.Close()
	}
}
