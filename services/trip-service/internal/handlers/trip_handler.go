package handlers

import (
	"net/http"

	"github.com/AuraReaper/voom/services/trip-service/internal/domain"
	"github.com/AuraReaper/voom/services/trip-service/internal/service"
	"github.com/AuraReaper/voom/shared/types"
	"github.com/labstack/echo/v4"
)

type TripHandler struct {
	svc *service.TripService
}

func NewTripHandler(svc *service.TripService) *TripHandler {
	return &TripHandler{
		svc: svc,
	}
}

func (h *TripHandler) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "Service is Healthy.")
}

func (h *TripHandler) CreateTrip(c echo.Context) error {
	var fare domain.RideFareModel
	if err := c.Bind(&fare); err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
	}

	trip, err := h.svc.CreateTrip(c.Request().Context(), &fare)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, trip)
}

func (h *TripHandler) GetRoute(c echo.Context) error {
	var request struct {
		Pickup      *types.Coordinate `json:"pickup"`
		Destination *types.Coordinate `json:"destination"`
	}
	
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	if request.Pickup == nil || request.Destination == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "pickup and destination coordinates are required",
		})
	}

	route, err := h.svc.GetRoute(c.Request().Context(), request.Pickup, request.Destination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, route)
}
