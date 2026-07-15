package service

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	providerPricingCurrency  = "CNY"
	providerTokenPriceUnit   = "per_1m_tokens"
	providerPerCallPriceUnit = "per_call"
)

type ProviderPricingData struct {
	Currency   string
	PriceUnit  string
	SiteName   string
	SiteDomain string
	UpdatedAt  time.Time
	Models     []ProviderPricingModel
}

type ProviderPricingModel struct {
	ModelName          string
	GroupName          string
	PriceUnit          string
	InputPrice         *float64
	OutputPrice        *float64
	CacheInputPrice    *float64
	CacheCreatePrice   *float64
	CacheCreatePrice1H *float64
	UnitPrice          *float64
	Enabled            bool
	Note               string
}

type ProviderPricingService struct {
	listAvailable     func(context.Context) ([]AvailableChannel, error)
	getPublicSettings func(context.Context) (*PublicSettings, error)
	now               func() time.Time
}

func NewProviderPricingService(channelService *ChannelService, settingService *SettingService) *ProviderPricingService {
	return newProviderPricingService(channelService.ListAvailable, settingService.GetPublicSettings, time.Now)
}

func newProviderPricingService(
	listAvailable func(context.Context) ([]AvailableChannel, error),
	getPublicSettings func(context.Context) (*PublicSettings, error),
	now func() time.Time,
) *ProviderPricingService {
	return &ProviderPricingService{
		listAvailable:     listAvailable,
		getPublicSettings: getPublicSettings,
		now:               now,
	}
}

func (s *ProviderPricingService) Get(ctx context.Context) (*ProviderPricingData, error) {
	channels, err := s.listAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("list available channels: %w", err)
	}
	settings, err := s.getPublicSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get public settings: %w", err)
	}
	if settings == nil {
		return nil, fmt.Errorf("get public settings: empty result")
	}

	data := &ProviderPricingData{
		Currency:   providerPricingCurrency,
		PriceUnit:  providerTokenPriceUnit,
		SiteName:   settings.SiteName,
		SiteDomain: providerPricingSiteDomain(settings.APIBaseURL),
		UpdatedAt:  s.now().UTC(),
		Models:     make([]ProviderPricingModel, 0),
	}
	seen := make(map[string]struct{})
	for _, channel := range channels {
		if channel.Status != StatusActive {
			continue
		}
		for _, group := range channel.Groups {
			groupName := strings.TrimSpace(group.Name)
			if groupName == "" || group.SubscriptionType != SubscriptionTypeStandard || group.PeakRateEnabled {
				continue
			}
			for _, model := range channel.SupportedModels {
				modelName := strings.TrimSpace(model.Name)
				if modelName == "" || !strings.EqualFold(group.Platform, model.Platform) {
					continue
				}
				row, ok := buildProviderPricingModel(modelName, groupName, model.Pricing, group)
				if !ok {
					continue
				}
				key := strings.ToLower(row.GroupName) + "\x00" + strings.ToLower(row.ModelName)
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				data.Models = append(data.Models, row)
			}
		}
	}
	sort.SliceStable(data.Models, func(i, j int) bool {
		leftGroup := strings.ToLower(data.Models[i].GroupName)
		rightGroup := strings.ToLower(data.Models[j].GroupName)
		if leftGroup != rightGroup {
			return leftGroup < rightGroup
		}
		return strings.ToLower(data.Models[i].ModelName) < strings.ToLower(data.Models[j].ModelName)
	})
	return data, nil
}

func buildProviderPricingModel(modelName, groupName string, pricing *ChannelModelPricing, group AvailableGroupRef) (ProviderPricingModel, bool) {
	if pricing == nil || len(pricing.Intervals) > 0 || !providerPricingValuesValid(pricing) {
		return ProviderPricingModel{}, false
	}

	base := ProviderPricingModel{
		ModelName: modelName,
		GroupName: groupName,
		Enabled:   true,
		Note:      "",
	}
	switch pricing.BillingMode {
	case "", BillingModeToken:
		if pricing.InputPrice == nil {
			return ProviderPricingModel{}, false
		}
		multiplier := 1_000_000 * group.RateMultiplier
		input, ok := scaleProviderPrice(pricing.InputPrice, multiplier)
		if !ok {
			return ProviderPricingModel{}, false
		}
		output, outputOK := scaleProviderPrice(pricing.OutputPrice, multiplier)
		cacheInput, cacheInputOK := scaleProviderPrice(pricing.CacheReadPrice, multiplier)
		cacheCreate, cacheCreateOK := scaleProviderPrice(pricing.CacheWritePrice, multiplier)
		if !outputOK || !cacheInputOK || !cacheCreateOK {
			return ProviderPricingModel{}, false
		}
		base.InputPrice = input
		base.OutputPrice = output
		base.CacheInputPrice = cacheInput
		base.CacheCreatePrice = cacheCreate
		return base, true
	case BillingModePerRequest:
		unitPrice, ok := scaleProviderPrice(pricing.PerRequestPrice, group.RateMultiplier)
		if !ok || unitPrice == nil {
			return ProviderPricingModel{}, false
		}
		base.PriceUnit = providerPerCallPriceUnit
		base.UnitPrice = unitPrice
		return base, true
	case BillingModeImage:
		if group.ImagePrice1K != nil || group.ImagePrice2K != nil || group.ImagePrice4K != nil {
			return ProviderPricingModel{}, false
		}
		multiplier := group.RateMultiplier
		if group.ImageRateIndependent {
			multiplier = group.ImageRateMultiplier
		}
		unitPrice, ok := scaleProviderPrice(pricing.PerRequestPrice, multiplier)
		if !ok || unitPrice == nil {
			return ProviderPricingModel{}, false
		}
		base.PriceUnit = providerPerCallPriceUnit
		base.UnitPrice = unitPrice
		return base, true
	default:
		return ProviderPricingModel{}, false
	}
}

func providerPricingValuesValid(pricing *ChannelModelPricing) bool {
	for _, price := range []*float64{
		pricing.InputPrice,
		pricing.OutputPrice,
		pricing.CacheWritePrice,
		pricing.CacheReadPrice,
		pricing.ImageOutputPrice,
		pricing.PerRequestPrice,
	} {
		if price != nil && (*price < 0 || math.IsNaN(*price) || math.IsInf(*price, 0)) {
			return false
		}
	}
	return true
}

func providerPricingSiteDomain(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !parsed.IsAbs() || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return strings.ToLower(parsed.Hostname())
}

func scaleProviderPrice(price *float64, multiplier float64) (*float64, bool) {
	if price == nil {
		return nil, true
	}
	if multiplier < 0 || math.IsNaN(multiplier) || math.IsInf(multiplier, 0) {
		return nil, false
	}
	value := *price * multiplier
	if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return nil, false
	}
	value = math.Round(value*1e8) / 1e8
	return &value, true
}
