package http

import (
	"hrms/internal/feature/ai/service"
	"hrms/internal/infrastructure/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AIHandler struct {
	service service.AIService
}

func NewAIHandler(svc service.AIService) *AIHandler {
	return &AIHandler{service: svc}
}

// ReviewTaskReport godoc
// @Summary      AI review of task report
// @Description  Admin requests an AI analysis of a submitted task report before approving or rejecting
// @Tags         AI
// @Produce      json
// @Param        id  path      string  true  "Task ID (UUID)"
// @Success      200  {object}  service.TaskReviewResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks/{id}/ai-review [post]
func (h *AIHandler) ReviewTaskReport(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	resp, err := h.service.ReviewTaskReport(c.Request.Context(), h.callerUserID(c), taskID)
	if err != nil {
		log.Printf("[AIHandler] ReviewTaskReport: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetAnalyticsSummary godoc
// @Summary      AI analytics summary
// @Description  Generate a natural language executive summary of the current HR analytics
// @Tags         AI
// @Produce      json
// @Param        date        query  string  false  "Attendance date YYYY-MM-DD"
// @Param        start_date  query  string  false  "Reporting period start YYYY-MM-DD"
// @Param        end_date    query  string  false  "Reporting period end YYYY-MM-DD"
// @Success      200  {object}  service.AnalyticsSummaryResponse
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/ai-summary [get]
func (h *AIHandler) GetAnalyticsSummary(c *gin.Context) {
	resp, err := h.service.GetAnalyticsSummary(c.Request.Context(), h.callerUserID(c), service.AnalyticsSummaryRequest{
		Date:      c.Query("date"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	})
	if err != nil {
		log.Printf("[AIHandler] GetAnalyticsSummary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AIHandler) callerUserID(c *gin.Context) uuid.UUID {
	if value, exists := c.Get(middleware.UserIDKey); exists {
		if id, ok := value.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}
