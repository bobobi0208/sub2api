//go:build unit

package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRegisterProviderPricingRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterProviderPricingRoutes(router, &handler.ProviderPricingHandler{})

	registered := router.Routes()
	require.Len(t, registered, 1)
	require.Equal(t, http.MethodGet, registered[0].Method)
	require.Equal(t, "/api/provider/pricing", registered[0].Path)

	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/api/provider/pricing", nil),
		httptest.NewRequest(http.MethodGet, "/api/v1/provider/pricing", nil),
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusNotFound, response.Code)
	}
}
