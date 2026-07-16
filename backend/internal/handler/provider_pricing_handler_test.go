//go:build unit

package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type stubProviderPricingReader struct {
	data  *service.ProviderPricingData
	err   error
	calls *int
}

func (s stubProviderPricingReader) Get(context.Context) (*service.ProviderPricingData, error) {
	if s.calls != nil {
		*s.calls = *s.calls + 1
	}
	return s.data, s.err
}

func TestProviderPricingHandler_PublicSuccess(t *testing.T) {
	inputPrice := 3.0
	unitPrice := 0.2
	updatedAt := time.Date(2026, time.July, 16, 4, 0, 0, 0, time.UTC)
	handler := newProviderPricingHandler(
		stubProviderPricingReader{data: &service.ProviderPricingData{
			Currency:   "CNY",
			PriceUnit:  "per_1m_tokens",
			SiteName:   "Example Site",
			SiteDomain: "",
			UpdatedAt:  updatedAt,
			Models: []service.ProviderPricingModel{
				{
					ModelName:  "gpt-5.5",
					GroupName:  "standard",
					InputPrice: &inputPrice,
					Enabled:    true,
					Note:       "",
				},
				{
					ModelName: "gpt-image-2",
					GroupName: "sora",
					PriceUnit: "per_call",
					UnitPrice: &unitPrice,
					Enabled:   true,
					Note:      "",
				},
			},
		}},
		"",
		func() time.Time { return updatedAt },
	)

	response := performProviderPricingRequest(t, handler, "fallback.example.com:8443", map[string]string{
		"X-Hvoy-Ts":   "ignored",
		"X-Hvoy-Sign": "ignored",
	})
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "application/json; charset=utf-8", response.Header().Get("Content-Type"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, "1.1", body["schema_version"])
	require.Equal(t, true, body["success"])
	require.Equal(t, "", body["message"])

	data := body["data"].(map[string]any)
	require.Equal(t, "CNY", data["currency"])
	require.Equal(t, "per_1m_tokens", data["price_unit"])
	require.Equal(t, "Example Site", data["site_name"])
	require.Equal(t, "fallback.example.com", data["site_domain"])
	require.Equal(t, "2026-07-16T04:00:00Z", data["updated_at"])

	models := data["models"].([]any)
	require.Len(t, models, 2)
	token := models[0].(map[string]any)
	require.Equal(t, "gpt-5.5", token["model_name"])
	require.Equal(t, float64(3), token["input_price"])
	for _, key := range []string{"output_price", "cache_input_price", "cache_create_price", "cache_create_price_1h"} {
		value, exists := token[key]
		require.Truef(t, exists, "token row must include %s", key)
		require.Nil(t, value)
	}
	require.NotContains(t, token, "price_unit")
	require.NotContains(t, token, "unit_price")

	perCall := models[1].(map[string]any)
	require.Equal(t, map[string]any{
		"model_name": "gpt-image-2",
		"group_name": "sora",
		"price_unit": "per_call",
		"unit_price": float64(0.2),
		"enabled":    true,
		"note":       "",
	}, perCall)
}

func TestProviderPricingHandler_ServiceError(t *testing.T) {
	handler := newProviderPricingHandler(
		stubProviderPricingReader{err: errors.New("database unavailable")},
		"",
		time.Now,
	)
	response := performProviderPricingRequest(t, handler, "example.com", nil)

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{
		"schema_version":"1.1",
		"success":false,
		"message":"provider pricing unavailable",
		"data":null
	}`, response.Body.String())
}

func TestProviderPricingHandler_HMAC(t *testing.T) {
	const secret = "provider-secret"
	fixedNow := time.Date(2026, time.July, 16, 4, 0, 0, 0, time.UTC)
	nowUnix := fixedNow.Unix()

	tests := []struct {
		name       string
		timestamp  string
		signature  func(string) string
		wantStatus int
	}{
		{
			name: "valid current timestamp", timestamp: unixString(nowUnix),
			signature:  func(ts string) string { return providerPricingTestSignature(secret, ts) },
			wantStatus: http.StatusOK,
		},
		{
			name: "valid past boundary", timestamp: unixString(nowUnix - 60),
			signature:  func(ts string) string { return providerPricingTestSignature(secret, ts) },
			wantStatus: http.StatusOK,
		},
		{
			name: "valid future boundary", timestamp: unixString(nowUnix + 60),
			signature:  func(ts string) string { return providerPricingTestSignature(secret, ts) },
			wantStatus: http.StatusOK,
		},
		{name: "missing headers", wantStatus: http.StatusUnauthorized},
		{
			name: "malformed timestamp", timestamp: "not-a-timestamp",
			signature:  func(ts string) string { return providerPricingTestSignature(secret, ts) },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed signature", timestamp: unixString(nowUnix),
			signature:  func(string) string { return "not-hex" },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "short signature", timestamp: unixString(nowUnix),
			signature:  func(string) string { return "abcd" },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "incorrect signature", timestamp: unixString(nowUnix),
			signature:  func(ts string) string { return providerPricingTestSignature("wrong-secret", ts) },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "expired timestamp", timestamp: unixString(nowUnix - 61),
			signature:  func(ts string) string { return providerPricingTestSignature(secret, ts) },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "future timestamp", timestamp: unixString(nowUnix + 61),
			signature:  func(ts string) string { return providerPricingTestSignature(secret, ts) },
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			handler := newProviderPricingHandler(
				stubProviderPricingReader{
					data: &service.ProviderPricingData{
						Currency: "CNY", PriceUnit: "per_1m_tokens", UpdatedAt: fixedNow,
						Models: []service.ProviderPricingModel{},
					},
					calls: &calls,
				},
				secret,
				func() time.Time { return fixedNow },
			)
			headers := map[string]string{}
			if test.timestamp != "" {
				headers["X-Hvoy-Ts"] = test.timestamp
			}
			if test.signature != nil {
				headers["X-Hvoy-Sign"] = test.signature(test.timestamp)
			}

			response := performProviderPricingRequest(t, handler, "example.com", headers)
			require.Equal(t, test.wantStatus, response.Code)
			if test.wantStatus == http.StatusUnauthorized {
				require.Equal(t, 0, calls)
				require.JSONEq(t, `{
					"schema_version":"1.1",
					"success":false,
					"message":"unauthorized",
					"data":null
				}`, response.Body.String())
			} else {
				require.Equal(t, 1, calls)
			}
		})
	}
}

func providerPricingTestSignature(secret, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp))
	return hex.EncodeToString(mac.Sum(nil))
}

func unixString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func performProviderPricingRequest(
	t *testing.T,
	handler *ProviderPricingHandler,
	host string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/provider/pricing", handler.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/provider/pricing", nil)
	req.Host = host
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}
