package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/AuraReaper/voom/services/api-gateway/grpc_clients"
	"github.com/AuraReaper/voom/services/api-gateway/pkg/types"
	"github.com/AuraReaper/voom/shared/contracts"
	"github.com/AuraReaper/voom/shared/env"
	"github.com/AuraReaper/voom/shared/messaging"
	"github.com/AuraReaper/voom/shared/tracing"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

var tracer = tracing.GetTracer("api-gateway")

func HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "Service is healthy.")
}

func HandleTripPreview(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "handleTripStart")
	defer span.End()

	var reqBody types.PreviewTripRequest
	if err := c.Bind(&reqBody); err != nil {
		return c.String(http.StatusBadRequest, "invalid request body")
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		c.Logger().Fatal(err)
	}

	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(ctx, reqBody.ToProto())
	if err != nil {
		c.Logger().Infof("failed to preview a trip: %v", err)
		return c.String(http.StatusInternalServerError, "failed to preview a trip")
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "request valid",
		"data":    tripPreview,
	})
}

func HandleCreateTrip(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "handleTripPreview")
	defer span.End()
	var req types.CreateTripRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "invalid request body")
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		c.Logger().Fatal(err)
	}

	defer tripService.Close()

	createTrip, err := tripService.Client.CreateTrip(ctx, req.ToProto())
	if err != nil {
		c.Logger().Infof("failed to create a trip: %v", err)
		return c.String(http.StatusInternalServerError, "failed to create a trip")
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "request valid",
		"data":    createTrip,
	})
}

func HandleStripeWebHook(c echo.Context, rb *messaging.RabbitMQ) error {
	ctx, span := tracer.Start(c.Request().Context(), "handleStripeWebhook")
	defer span.End()

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to read request body")
	}
	defer c.Request().Body.Close()

	webHookKey := env.GetString("STRIPE_WEBHOOK_KEY", "")
	if webHookKey == "" {
		return fmt.Errorf("webhook key is required")
	}

	sigHeader := c.Request().Header.Get("Stripe-Signature")

	event, err := webhook.ConstructEventWithOptions(
		body,
		sigHeader,
		webHookKey,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		return c.String(http.StatusBadRequest, "Invalid signature")
	}

	log.Printf("Received Stripe event: %s", event.Type)

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook session: %v", err)
			return c.String(http.StatusBadRequest, "Invalid Payload")
		}

		payload := messaging.PaymentStatusUpdateData{
			TripID:   session.Metadata["trip_id"],
			UserID:   session.Metadata["user_id"],
			DriverID: session.Metadata["driver_id"],
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshalling payload: %v", err)
			return c.String(http.StatusInternalServerError, "failed to marshall Payload")
		}

		message := contracts.AmqpMessage{
			OwnerID: session.Metadata["user_id"],
			Data:    payloadBytes,
		}

		if err := rb.PublishMessage(
			ctx,
			contracts.PaymentEventSuccess,
			message,
		); err != nil {
			log.Printf("Error publishing paymeny event: %v", err)
			return c.String(http.StatusInternalServerError, "failed to publish payment event")
		}
	}

	return nil
}
