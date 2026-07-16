package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const providerPricingSchemaVersion = "1.1"

type providerPricingReader interface {
	Get(context.Context) (*service.ProviderPricingData, error)
}

type ProviderPricingHandler struct {
	pricingService providerPricingReader
	authSecret     string
	now            func() time.Time
}

func NewProviderPricingHandler(pricingService *service.ProviderPricingService, cfg *config.Config) *ProviderPricingHandler {
	authSecret := ""
	if cfg != nil {
		authSecret = cfg.ProviderPricing.AuthSecret
	}
	return newProviderPricingHandler(pricingService, authSecret, time.Now)
}

func newProviderPricingHandler(
	pricingService providerPricingReader,
	authSecret string,
	now func() time.Time,
) *ProviderPricingHandler {
	return &ProviderPricingHandler{
		pricingService: pricingService,
		authSecret:     authSecret,
		now:            now,
	}
}

func (h *ProviderPricingHandler) Get(c *gin.Context) {
	if !h.authorized(c) {
		h.writeEnvelope(c, http.StatusUnauthorized, false, "unauthorized", nil)
		return
	}

	data, err := h.pricingService.Get(c.Request.Context())
	if err != nil {
		slog.Error("provider pricing feed unavailable", "error", err)
		h.writeEnvelope(c, http.StatusInternalServerError, false, "provider pricing unavailable", nil)
		return
	}

	siteDomain := data.SiteDomain
	if siteDomain == "" {
		siteDomain = providerPricingRequestHost(c.Request.Host)
	}
	responseData := providerPricingResponseData{
		Currency:   data.Currency,
		PriceUnit:  data.PriceUnit,
		SiteName:   data.SiteName,
		SiteDomain: siteDomain,
		UpdatedAt:  data.UpdatedAt.UTC().Format(time.RFC3339Nano),
		Models:     providerPricingResponseModels(data.Models),
	}
	h.writeEnvelope(c, http.StatusOK, true, "", responseData)
}

func (h *ProviderPricingHandler) authorized(c *gin.Context) bool {
	if h.authSecret == "" {
		return true
	}

	timestampHeader := c.GetHeader("X-Hvoy-Ts")
	signatureHeader := c.GetHeader("X-Hvoy-Sign")
	if timestampHeader == "" || signatureHeader == "" {
		return false
	}
	timestamp, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return false
	}
	now := h.now().Unix()
	if timestamp < now-60 || timestamp > now+60 {
		return false
	}

	signature, err := hex.DecodeString(signatureHeader)
	if err != nil || len(signature) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.authSecret))
	_, _ = mac.Write([]byte(timestampHeader))
	return hmac.Equal(signature, mac.Sum(nil))
}

func (h *ProviderPricingHandler) writeEnvelope(c *gin.Context, status int, success bool, message string, data any) {
	c.JSON(status, providerPricingEnvelope{
		SchemaVersion: providerPricingSchemaVersion,
		Success:       success,
		Message:       message,
		Data:          data,
	})
}

type providerPricingEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Data          any    `json:"data"`
}

type providerPricingResponseData struct {
	Currency   string           `json:"currency"`
	PriceUnit  string           `json:"price_unit"`
	SiteName   string           `json:"site_name,omitempty"`
	SiteDomain string           `json:"site_domain,omitempty"`
	UpdatedAt  string           `json:"updated_at"`
	Models     []map[string]any `json:"models"`
}

func providerPricingResponseModels(models []service.ProviderPricingModel) []map[string]any {
	rows := make([]map[string]any, 0, len(models))
	for _, model := range models {
		row := map[string]any{
			"model_name": model.ModelName,
			"group_name": model.GroupName,
			"enabled":    model.Enabled,
			"note":       model.Note,
		}
		if model.PriceUnit == "per_call" {
			row["price_unit"] = model.PriceUnit
			row["unit_price"] = model.UnitPrice
		} else {
			row["input_price"] = model.InputPrice
			row["output_price"] = model.OutputPrice
			row["cache_input_price"] = model.CacheInputPrice
			row["cache_create_price"] = model.CacheCreatePrice
			row["cache_create_price_1h"] = model.CacheCreatePrice1H
		}
		rows = append(rows, row)
	}
	return rows
}

func providerPricingRequestHost(host string) string {
	parsed := &url.URL{Host: strings.TrimSpace(host)}
	return strings.ToLower(parsed.Hostname())
}
