package http

import (
	"hrms/internal/feature/city/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CityHandler struct {
	svc service.CityService
}

func NewCityHandler(svc service.CityService) *CityHandler {
	return &CityHandler{svc: svc}
}

// ListCities godoc
// @Summary      List cities
// @Description  Returns all available cities in Kazakhstan
// @Tags         Cities
// @Produce      json
// @Success      200  {array}   repository.City
// @Failure      500  {object}  map[string]string
// @Router       /cities [get]
func (h *CityHandler) ListCities(c *gin.Context) {
	cities, err := h.svc.ListCities(c.Request.Context())
	if err != nil {
		log.Printf("[Handler] ListCities error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, cities)
}
