//go:build unit

package service

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProviderPricingService_TokenPrices(t *testing.T) {
	fixedNow := time.Date(2026, time.July, 16, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	service := newProviderPricingService(
		func(context.Context) ([]AvailableChannel, error) {
			return []AvailableChannel{{
				Name:   "openai-channel",
				Status: StatusActive,
				Groups: []AvailableGroupRef{{
					ID:               1,
					Name:             "standard",
					Platform:         PlatformOpenAI,
					SubscriptionType: SubscriptionTypeStandard,
					RateMultiplier:   1.5,
				}},
				SupportedModels: []SupportedModel{{
					Name:     "gpt-5.5",
					Platform: PlatformOpenAI,
					Pricing: &ChannelModelPricing{
						BillingMode:     BillingModeToken,
						InputPrice:      testPtrFloat64(2e-6),
						OutputPrice:     testPtrFloat64(8e-6),
						CacheReadPrice:  testPtrFloat64(5e-7),
						CacheWritePrice: testPtrFloat64(2.5e-6),
					},
				}},
			}}, nil
		},
		func(context.Context) (*PublicSettings, error) {
			return &PublicSettings{
				SiteName:   "Example Site",
				APIBaseURL: "https://api.example.com/v1",
			}, nil
		},
		func() time.Time { return fixedNow },
	)

	got, err := service.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "CNY", got.Currency)
	require.Equal(t, "per_1m_tokens", got.PriceUnit)
	require.Equal(t, "Example Site", got.SiteName)
	require.Equal(t, "api.example.com", got.SiteDomain)
	require.Equal(t, fixedNow.UTC(), got.UpdatedAt)
	require.Len(t, got.Models, 1)

	model := got.Models[0]
	require.Equal(t, "gpt-5.5", model.ModelName)
	require.Equal(t, "standard", model.GroupName)
	require.Empty(t, model.PriceUnit)
	require.InDelta(t, 3, *model.InputPrice, 1e-9)
	require.InDelta(t, 12, *model.OutputPrice, 1e-9)
	require.InDelta(t, 0.75, *model.CacheInputPrice, 1e-9)
	require.InDelta(t, 3.75, *model.CacheCreatePrice, 1e-9)
	require.Nil(t, model.CacheCreatePrice1H)
	require.Nil(t, model.UnitPrice)
	require.True(t, model.Enabled)
	require.Empty(t, model.Note)
}

func TestProviderPricingService_SiteDomain(t *testing.T) {
	service := newProviderPricingService(
		func(context.Context) ([]AvailableChannel, error) { return nil, nil },
		func(context.Context) (*PublicSettings, error) {
			return &PublicSettings{SiteName: "Example", APIBaseURL: "not-an-absolute-url"}, nil
		},
		time.Now,
	)

	got, err := service.Get(context.Background())
	require.NoError(t, err)
	require.Empty(t, got.SiteDomain)
	require.NotNil(t, got.Models)
}

func TestProviderPricingService_PerCall(t *testing.T) {
	got := getProviderPricingForTest(t, []AvailableChannel{{
		Name:   "request-channel",
		Status: StatusActive,
		Groups: []AvailableGroupRef{{
			Name:             "request-group",
			Platform:         PlatformOpenAI,
			SubscriptionType: SubscriptionTypeStandard,
			RateMultiplier:   1.5,
		}},
		SupportedModels: []SupportedModel{{
			Name:     "request-model",
			Platform: PlatformOpenAI,
			Pricing: &ChannelModelPricing{
				BillingMode:     BillingModePerRequest,
				PerRequestPrice: testPtrFloat64(0.2),
			},
		}},
	}})

	require.Len(t, got.Models, 1)
	model := got.Models[0]
	require.Equal(t, providerPerCallPriceUnit, model.PriceUnit)
	require.InDelta(t, 0.3, *model.UnitPrice, 1e-9)
	require.Nil(t, model.InputPrice)
}

func TestProviderPricingService_Image(t *testing.T) {
	got := getProviderPricingForTest(t, []AvailableChannel{{
		Name:   "image-channel",
		Status: StatusActive,
		Groups: []AvailableGroupRef{
			{
				Name:                 "image-independent",
				Platform:             PlatformGemini,
				SubscriptionType:     SubscriptionTypeStandard,
				RateMultiplier:       2,
				ImageRateIndependent: true,
				ImageRateMultiplier:  0.5,
			},
			{
				Name:             "image-default",
				Platform:         PlatformGemini,
				SubscriptionType: SubscriptionTypeStandard,
				RateMultiplier:   2,
			},
			{
				Name:             "image-tiered",
				Platform:         PlatformGemini,
				SubscriptionType: SubscriptionTypeStandard,
				RateMultiplier:   2,
				ImagePrice2K:     testPtrFloat64(0.3),
			},
		},
		SupportedModels: []SupportedModel{{
			Name:     "image-model",
			Platform: PlatformGemini,
			Pricing: &ChannelModelPricing{
				BillingMode:     BillingModeImage,
				PerRequestPrice: testPtrFloat64(0.2),
			},
		}},
	}})

	require.Len(t, got.Models, 2)
	byGroup := make(map[string]ProviderPricingModel, len(got.Models))
	for _, model := range got.Models {
		byGroup[model.GroupName] = model
	}
	require.InDelta(t, 0.1, *byGroup["image-independent"].UnitPrice, 1e-9)
	require.InDelta(t, 0.4, *byGroup["image-default"].UnitPrice, 1e-9)
	require.NotContains(t, byGroup, "image-tiered")
}

func TestProviderPricingService_Filters(t *testing.T) {
	validPrice := &ChannelModelPricing{BillingMode: BillingModeToken, InputPrice: testPtrFloat64(1e-6)}
	got := getProviderPricingForTest(t, []AvailableChannel{
		{
			Name:   "inactive-channel",
			Status: StatusDisabled,
			Groups: []AvailableGroupRef{{
				Name: "inactive", Platform: PlatformOpenAI,
				SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1,
			}},
			SupportedModels: []SupportedModel{{Name: "inactive-model", Platform: PlatformOpenAI, Pricing: validPrice}},
		},
		{
			Name:   "active-channel",
			Status: StatusActive,
			Groups: []AvailableGroupRef{
				{Name: "standard", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1},
				{Name: "subscription", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription, RateMultiplier: 1},
				{Name: "peak", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1, PeakRateEnabled: true},
				{Name: "wrong-platform", Platform: PlatformAnthropic, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1},
				{Name: "", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1},
			},
			SupportedModels: []SupportedModel{
				{Name: "valid", Platform: PlatformOpenAI, Pricing: validPrice},
				{Name: "interval", Platform: PlatformOpenAI, Pricing: &ChannelModelPricing{
					BillingMode: BillingModeToken, InputPrice: testPtrFloat64(1e-6),
					Intervals: []PricingInterval{{MinTokens: 0}},
				}},
				{Name: "missing-input", Platform: PlatformOpenAI, Pricing: &ChannelModelPricing{BillingMode: BillingModeToken}},
				{Name: "video", Platform: PlatformOpenAI, Pricing: &ChannelModelPricing{BillingMode: BillingModeVideo, PerRequestPrice: testPtrFloat64(0.2)}},
				{Name: "", Platform: PlatformOpenAI, Pricing: validPrice},
			},
		},
	})

	require.Len(t, got.Models, 1)
	require.Equal(t, "valid", got.Models[0].ModelName)
	require.Equal(t, "standard", got.Models[0].GroupName)
}

func TestProviderPricingService_FiltersConfiguredIntervalsAfterGlobalFallback(t *testing.T) {
	channelRepo := &mockChannelRepository{
		listAllFn: func(context.Context) ([]Channel, error) {
			return []Channel{{
				ID:       1,
				Name:     "image-channel",
				Status:   StatusActive,
				GroupIDs: []int64{1},
				ModelPricing: []ChannelModelPricing{{
					Platform:    PlatformGemini,
					Models:      []string{"image-model"},
					BillingMode: BillingModeImage,
					Intervals:   []PricingInterval{{TierLabel: "1K"}, {TierLabel: "2K"}},
				}},
			}}, nil
		},
	}
	groupRepo := &stubGroupRepoForAvailable{activeGroups: []Group{{
		ID: 1, Name: "standard", Platform: PlatformGemini, Status: StatusActive,
		SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1,
	}}}
	channelService := NewChannelService(
		channelRepo,
		groupRepo,
		nil,
		newStubPricingServiceFromMap(map[string]*LiteLLMModelPricing{
			"image-model": {Mode: "image_generation", OutputCostPerImage: 0.2},
		}),
	)
	channels, err := channelService.ListAvailable(context.Background())
	require.NoError(t, err)
	require.Len(t, channels, 1)
	require.Len(t, channels[0].SupportedModels, 1)
	require.NotNil(t, channels[0].SupportedModels[0].Pricing)
	require.NotNil(t, channels[0].SupportedModels[0].Pricing.PerRequestPrice)
	require.True(t, channels[0].SupportedModels[0].HasConfiguredIntervals)

	got := getProviderPricingForTest(t, channels)
	require.Empty(t, got.Models, "configured intervals must stay excluded after global fallback")
}

func TestProviderPricingService_RejectsInvalid(t *testing.T) {
	got := getProviderPricingForTest(t, []AvailableChannel{{
		Name:   "invalid-channel",
		Status: StatusActive,
		Groups: []AvailableGroupRef{{
			Name: "standard", Platform: PlatformOpenAI,
			SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1,
		}},
		SupportedModels: []SupportedModel{
			{Name: "valid", Platform: PlatformOpenAI, Pricing: &ChannelModelPricing{InputPrice: testPtrFloat64(1e-6)}},
			{Name: "negative", Platform: PlatformOpenAI, Pricing: &ChannelModelPricing{InputPrice: testPtrFloat64(-1e-6)}},
			{Name: "nan", Platform: PlatformOpenAI, Pricing: &ChannelModelPricing{InputPrice: testPtrFloat64(math.NaN())}},
			{Name: "infinity", Platform: PlatformOpenAI, Pricing: &ChannelModelPricing{InputPrice: testPtrFloat64(math.Inf(1))}},
			{Name: "invalid-optional", Platform: PlatformOpenAI, Pricing: &ChannelModelPricing{
				InputPrice: testPtrFloat64(1e-6), OutputPrice: testPtrFloat64(math.NaN()),
			}},
		},
	}})

	require.Len(t, got.Models, 1)
	require.Equal(t, "valid", got.Models[0].ModelName)
	require.Nil(t, got.Models[0].OutputPrice)
}

func TestProviderPricingService_DeduplicatesAndSorts(t *testing.T) {
	price := &ChannelModelPricing{InputPrice: testPtrFloat64(1e-6)}
	got := getProviderPricingForTest(t, []AvailableChannel{
		{
			Name:   "first",
			Status: StatusActive,
			Groups: []AvailableGroupRef{
				{Name: "Beta", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1},
				{Name: "alpha", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1},
			},
			SupportedModels: []SupportedModel{
				{Name: "zeta", Platform: PlatformOpenAI, Pricing: price},
				{Name: "Alpha", Platform: PlatformOpenAI, Pricing: price},
			},
		},
		{
			Name:   "second",
			Status: StatusActive,
			Groups: []AvailableGroupRef{{
				Name: "ALPHA", Platform: PlatformOpenAI,
				SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 9,
			}},
			SupportedModels: []SupportedModel{{Name: "ZETA", Platform: PlatformOpenAI, Pricing: price}},
		},
	})

	require.Len(t, got.Models, 4)
	require.Equal(t, []string{
		"alpha/Alpha",
		"alpha/zeta",
		"Beta/Alpha",
		"Beta/zeta",
	}, providerPricingModelKeys(got.Models))
	require.InDelta(t, 1, *got.Models[1].InputPrice, 1e-9, "first duplicate must win")
}

func TestProviderPricingService_Errors(t *testing.T) {
	t.Run("available channels", func(t *testing.T) {
		sentinel := errors.New("channels unavailable")
		service := newProviderPricingService(
			func(context.Context) ([]AvailableChannel, error) { return nil, sentinel },
			func(context.Context) (*PublicSettings, error) { return &PublicSettings{}, nil },
			time.Now,
		)

		got, err := service.Get(context.Background())
		require.Nil(t, got)
		require.ErrorIs(t, err, sentinel)
	})

	t.Run("public settings", func(t *testing.T) {
		sentinel := errors.New("settings unavailable")
		service := newProviderPricingService(
			func(context.Context) ([]AvailableChannel, error) { return nil, nil },
			func(context.Context) (*PublicSettings, error) { return nil, sentinel },
			time.Now,
		)

		got, err := service.Get(context.Background())
		require.Nil(t, got)
		require.ErrorIs(t, err, sentinel)
	})
}

func getProviderPricingForTest(t *testing.T, channels []AvailableChannel) *ProviderPricingData {
	t.Helper()
	service := newProviderPricingService(
		func(context.Context) ([]AvailableChannel, error) { return channels, nil },
		func(context.Context) (*PublicSettings, error) {
			return &PublicSettings{SiteName: "Example", APIBaseURL: "https://api.example.com"}, nil
		},
		func() time.Time { return time.Date(2026, time.July, 16, 4, 0, 0, 0, time.UTC) },
	)
	got, err := service.Get(context.Background())
	require.NoError(t, err)
	return got
}

func providerPricingModelKeys(models []ProviderPricingModel) []string {
	keys := make([]string, 0, len(models))
	for _, model := range models {
		keys = append(keys, model.GroupName+"/"+model.ModelName)
	}
	return keys
}
