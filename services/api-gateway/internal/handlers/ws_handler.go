package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/AuraReaper/voom/services/api-gateway/grpc_clients"
	"github.com/AuraReaper/voom/shared/contracts"
	"github.com/AuraReaper/voom/shared/messaging"
	"github.com/AuraReaper/voom/shared/proto/driver"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	connManager = messaging.NewConnectionManager()
)

func HandleRidersWebSocket(c echo.Context, rabbitmq *messaging.RabbitMQ) error {
	conn, err := connManager.Upgrade(c.Response(), c.Request())
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
	}

	defer conn.Close()

	userID := c.QueryParam("userID")
	if userID == "" {
		c.Logger().Warn("userID not present.")
		return echo.NewHTTPError(http.StatusBadRequest, "userID is required")
	}

	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	queues := []string{
		messaging.NotifyNoDriverFoundQueue,
		messaging.NotifyDriverAssignQueue,
		messaging.NotifyPaymentSessionCreatedQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rabbitmq, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("failed to start comnsumer for queue: %s: err: %v", q, err)
		}
	}

	// Read loop
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Logger().Errorf("WebSocket error: %v", err)
			}
			break
		}
		c.Logger().Infof("Received message: %s", string(msg))
	}

	return nil
}

func HandleDriversWebSocket(c echo.Context, rabbitmq *messaging.RabbitMQ) error {
	// Validate parameters before upgrading

	query := c.Request().URL.Query()
	userID := query.Get("userID")
	if userID == "" {
		c.Logger().Warn("userID not present.")
		return echo.NewHTTPError(http.StatusBadRequest, "userID is required")
	}

	packageSlug := query.Get("packageSlug")
	if packageSlug == "" {
		c.Logger().Warn("packageSlug not present.")
		return echo.NewHTTPError(http.StatusBadRequest, "packageSlug is required")
	}

	conn, err := connManager.Upgrade(c.Response(), c.Request())
	if err != nil {
		c.Logger().Errorf("Failed to upgrade to websocket: %v", err)
		return err
	}

	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	defer conn.Close()

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		c.Logger().Errorf("failed to create driver service client: %v", err)
		return err
	}

	defer driverService.Close()

	c.Logger().Info("Registering driver...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	driver, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		c.Logger().Errorf("error registering driver: %v", err)
		return err
	}
	c.Logger().Info("Successfully registered driver")

	// Write a message
	if err := connManager.SendMessage(userID, contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driver,
	}); err != nil {
		c.Logger().Errorf("Error sending message: %v", err)
		return err
	}
	c.Logger().Info("message sent successfully.")

	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rabbitmq, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("failed to start comnsumer for queue: %s: err: %v", q, err)
		}
	}

	// Read loop
	for {
		_, msg, err := conn.ReadMessage()
		c.Logger().Infof("Received message from driver: %s", string(msg))
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Logger().Errorf("WebSocket error: %v", err)
			}
			break
		}

		type driverMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var driverMsg driverMessage
		if err := json.Unmarshal(msg, &driverMsg); err != nil {
			log.Printf("Error unmarsahlling driver message: %v", err)
			continue
		}

		// handle the diffrent message type
		switch driverMsg.Type {
		case contracts.DriverCmdLocation:
			// handle location here
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			// forward msg to rabbitmq
			if err := rabbitmq.PublishMessage(ctx, driverMsg.Type, contracts.AmqpMessage{
				OwnerID: userID,
				Data:    driverMsg.Data,
			}); err != nil {
				log.Printf("error publishing message to rabbitmq: %v", err)
			}
		default:
			log.Printf("unknown message type: %s", driverMsg.Type)
		}
	}

	return nil
}
