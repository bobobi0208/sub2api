package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProviderPricingRoutes(r *gin.Engine, h *handler.ProviderPricingHandler) {
	r.GET("/api/provider/pricing", h.Get)
}
