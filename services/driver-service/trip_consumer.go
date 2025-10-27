package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/AuraReaper/voom/shared/contracts"
	"github.com/AuraReaper/voom/shared/messaging"
	"github.com/rabbitmq/amqp091-go"
)

type tripEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripEventComsumer(rabbitmq *messaging.RabbitMQ, service *Service) *tripEventConsumer {
	return &tripEventConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *tripEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Fatalf("failed to unmrashal the message: %v", err)
			return err
		}

		var payload messaging.TripEventData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			log.Printf("failed to unmarshall message: %v", err)
			return err
		}

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.handleFindAndNotifyDrivers(ctx, payload)
		}

		log.Printf("unknown trip event: %+v", payload)

		log.Printf("driver received message: %+v", msg)
		return nil
	})
}

func (c *tripEventConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	suitableIDs, err := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)
	if err != nil {
		log.Printf("failed to find available drivers: %v", err)
		return err
	}

	log.Printf("Found suitable drivers %v", len(suitableIDs))

	if len(suitableIDs) == 0 {
		// Notify the driver that no drivers are available
		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payload.Trip.UserID,
		}); err != nil {
			log.Printf("Failed to publish message to exchange: %v", err)
			return err
		}

		return nil
	}

	for _, suitableDriverID := range suitableIDs {
		marshalledEvent, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal payload for driver %s: %v", suitableDriverID, err)
			continue // Skip to the next driver
		}

		// Notify the driver about a potential trip
		if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
			OwnerID: suitableDriverID,
			Data:    marshalledEvent,
		}); err != nil {
			log.Printf("Failed to publish message to exchange for driver %s: %v", suitableDriverID, err)
			// Decide if you want to stop or continue. For now, we'll log and continue.
		}
	}

	return nil
}
