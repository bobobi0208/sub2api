# Provider Pricing API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `GET /api/provider/pricing` with Hvoy schema `1.1`, CNY prices derived from active pay-as-you-go groups, and optional HMAC authentication.

**Architecture:** Extend the existing internal available-channel view with the image billing metadata needed for an unambiguous public price, then add a focused `ProviderPricingService` that transforms that view into deterministic model/group rows. A `ProviderPricingHandler` owns Hvoy envelopes, request-host fallback, HMAC verification, and HTTP status codes; root route registration and Wire connect it to the application.

**Tech Stack:** Go 1.26, Gin, Viper, Google Wire, Testify, Docker Compose.

## Global Constraints

- Route: exact `GET /api/provider/pricing` outside `/api/v1`.
- Schema: Hvoy `1.1`, `application/json; charset=utf-8`.
- Currency: fixed `CNY`; conversion: fixed `1 USD = 1 CNY`.
- Token price: internal `USD/token * 1,000,000 * group multiplier`.
- Per-call price: internal `USD/call * group multiplier`.
- Publish active standard pay-as-you-go groups only; exclude subscription and peak-rate groups.
- Skip interval, resolution-tier, unsupported billing, missing required price, negative, NaN, or infinite price rows.
- `PROVIDER_PRICING_AUTH_SECRET` empty means public; otherwise require HMAC-SHA256 headers within 60 seconds.
- Do not add database tables or admin UI, change production pricing/mappings, deploy, or merge `prod`.

---

### Task 1: Configuration And Available Group Metadata

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/internal/service/channel_available.go`
- Modify: `backend/internal/service/channel_available_test.go`

**Interfaces:**
- Produces: `config.ProviderPricingConfig{AuthSecret string}` and `Config.ProviderPricing`.
- Produces: internal-only `AvailableGroupRef` image fields used by Task 2.
- Preserves: `handler.userAvailableGroup` JSON field whitelist unchanged.

- [ ] **Step 1: Write failing configuration and metadata tests**

Add a config test that calls `resetViperWithJWTSecret(t)`, sets `PROVIDER_PRICING_AUTH_SECRET=test-secret`, calls `Load()`, and asserts `cfg.ProviderPricing.AuthSecret == "test-secret"`. Extend `TestListAvailable_InactiveGroupIDSilentlyDropped` or add a focused test with this source group:

```go
Group{
    ID: 1, Name: "image", Platform: PlatformGemini,
    ImageRateIndependent: true, ImageRateMultiplier: 0.75,
    ImagePrice1K: testPtrFloat64(0.1),
    ImagePrice2K: testPtrFloat64(0.2),
    ImagePrice4K: testPtrFloat64(0.3),
}
```

Assert the resulting `AvailableGroupRef` carries the independent multiplier and all three tier prices.

- [ ] **Step 2: Run tests and verify RED**

Run:

```bash
cd backend
go test -tags=unit ./internal/config ./internal/service -run 'TestLoadProviderPricingAuthSecret|TestListAvailable_CarriesImagePricingMetadata' -count=1
```

Expected: compilation failure because `Config.ProviderPricing` and the new `AvailableGroupRef` fields do not exist.

- [ ] **Step 3: Add the minimal configuration and metadata implementation**

Add to `Config`:

```go
ProviderPricing ProviderPricingConfig `mapstructure:"provider_pricing"`
```

Add:

```go
type ProviderPricingConfig struct {
    AuthSecret string `mapstructure:"auth_secret"`
}
```

Set `provider_pricing.auth_secret` to `""` in `setDefaults()` and normalize it with `strings.TrimSpace` after Viper unmarshalling. Extend `AvailableGroupRef` and the `ListAvailable` mapping with:

```go
ImageRateIndependent bool
ImageRateMultiplier  float64
ImagePrice1K         *float64
ImagePrice2K         *float64
ImagePrice4K         *float64
```

- [ ] **Step 4: Run focused tests and verify GREEN**

Run the Step 2 command again. Expected: both focused tests pass.

- [ ] **Step 5: Verify the existing user DTO still hides internal image fields**

Extend `TestUserAvailableChannel_FieldWhitelist` to assert that serialized user group JSON does not contain `image_rate_independent`, `image_rate_multiplier`, `image_price_1k`, `image_price_2k`, or `image_price_4k`.

Run:

```bash
cd backend
go test -tags=unit ./internal/handler -run TestUserAvailableChannel_FieldWhitelist -count=1
```

Expected: PASS without changing the user DTO.

- [ ] **Step 6: Commit Task 1**

```bash
git add backend/internal/config/config.go backend/internal/config/config_test.go backend/internal/service/channel_available.go backend/internal/service/channel_available_test.go backend/internal/handler/available_channel_handler_test.go
git commit -m "feat: add provider pricing configuration"
```

---

### Task 2: Provider Pricing Transformation Service

**Files:**
- Create: `backend/internal/service/provider_pricing.go`
- Create: `backend/internal/service/provider_pricing_test.go`
- Modify: `backend/internal/service/wire.go`

**Interfaces:**
- Consumes: `ChannelService.ListAvailable(context.Context) ([]AvailableChannel, error)`.
- Consumes: `SettingService.GetPublicSettings(context.Context) (*PublicSettings, error)`.
- Produces: `NewProviderPricingService(*ChannelService, *SettingService) *ProviderPricingService`.
- Produces: `(*ProviderPricingService).Get(context.Context) (*ProviderPricingData, error)`.

- [ ] **Step 1: Write failing token conversion and identity tests**

Create a unit-tagged table test using injected functions for `ListAvailable`, `GetPublicSettings`, and a fixed clock. Use an active channel, a standard group with multiplier `1.5`, and token prices `2e-6`, `8e-6`, `5e-7`, and `2.5e-6`. Assert:

```go
require.Equal(t, "CNY", got.Currency)
require.Equal(t, "per_1m_tokens", got.PriceUnit)
require.Equal(t, "Example Site", got.SiteName)
require.Equal(t, "api.example.com", got.SiteDomain)
require.Equal(t, fixedNow.UTC(), got.UpdatedAt)
require.InDelta(t, 3, *got.Models[0].InputPrice, 1e-9)
require.InDelta(t, 12, *got.Models[0].OutputPrice, 1e-9)
require.InDelta(t, 0.75, *got.Models[0].CacheInputPrice, 1e-9)
require.InDelta(t, 3.75, *got.Models[0].CacheCreatePrice, 1e-9)
require.Nil(t, got.Models[0].CacheCreatePrice1H)
```

Also assert an invalid API base URL yields an empty `SiteDomain` for Handler fallback.

- [ ] **Step 2: Run token tests and verify RED**

Run:

```bash
cd backend
go test -tags=unit ./internal/service -run 'TestProviderPricingService_(TokenPrices|SiteDomain)' -count=1
```

Expected: compilation failure because `ProviderPricingService` does not exist.

- [ ] **Step 3: Implement service types and token conversion**

Define:

```go
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
```

The constructor stores production function dependencies and `time.Now`; an unexported constructor accepts function dependencies for unit tests. `Get` loads available channels and public settings, filters active channels/standard non-peak same-platform groups, rejects intervals, treats empty billing mode as token, multiplies token fields by `1_000_000 * RateMultiplier`, and rounds finite non-negative results with `math.Round(value*1e8)/1e8`.

- [ ] **Step 4: Run token tests and verify GREEN**

Run the Step 2 command again. Expected: PASS.

- [ ] **Step 5: Write failing per-call, image multiplier, filter, dedup, and sorting tests**

Add focused tests proving:

```text
per_request 0.2 * multiplier 1.5 = unit_price 0.3
image 0.2 * independent image multiplier 0.5 = unit_price 0.1
image with any ImagePrice1K/ImagePrice2K/ImagePrice4K is skipped
inactive channels, subscription groups, peak groups, platform mismatches, intervals,
missing input/per-request prices, video mode, negative/NaN/Inf values are skipped
duplicate lowercase model_name+group_name rows keep the first deterministic row
rows sort by lowercase group_name then lowercase model_name
```

- [ ] **Step 6: Run the new tests and verify RED**

Run:

```bash
cd backend
go test -tags=unit ./internal/service -run 'TestProviderPricingService_(PerCall|Image|Filters|Deduplicates|Sorts|RejectsInvalid)' -count=1
```

Expected: assertions fail because only token conversion exists.

- [ ] **Step 7: Implement per-call conversion and all filters**

For `BillingModePerRequest`, emit `PriceUnit: "per_call"` using `RateMultiplier`. For `BillingModeImage`, emit `PriceUnit: "per_call"`, reject explicit group resolution prices, and choose `ImageRateMultiplier` only when `ImageRateIndependent` is true; otherwise use `RateMultiplier`. Deduplicate with `strings.ToLower(model.Name)+"\x00"+strings.ToLower(group.Name)` and stable-sort group then model with lowercase comparisons.

- [ ] **Step 8: Run all provider service tests and verify GREEN**

Run:

```bash
cd backend
go test -tags=unit ./internal/service -run TestProviderPricingService -count=1
```

Expected: PASS.

- [ ] **Step 9: Register the service and commit Task 2**

Add `NewProviderPricingService` to `service.ProviderSet`, then run:

```bash
cd backend
go test -tags=unit ./internal/service -run 'TestProviderPricingService|TestProviderSet' -count=1
cd ..
git add backend/internal/service/provider_pricing.go backend/internal/service/provider_pricing_test.go backend/internal/service/wire.go
git commit -m "feat: build provider pricing feed"
```

Expected: tests pass and commit succeeds.

---

### Task 3: Hvoy Handler And Optional HMAC

**Files:**
- Create: `backend/internal/handler/provider_pricing_handler.go`
- Create: `backend/internal/handler/provider_pricing_handler_test.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/wire.go`

**Interfaces:**
- Consumes: `ProviderPricingService.Get(context.Context)` and `config.Config.ProviderPricing.AuthSecret`.
- Produces: `NewProviderPricingHandler(*service.ProviderPricingService, *config.Config) *ProviderPricingHandler`.
- Produces: `(*ProviderPricingHandler).Get(*gin.Context)`.

- [ ] **Step 1: Write failing public success and serialization tests**

Use a fake provider reader and fixed handler clock. Call `GET /api/provider/pricing` with no secret and assert HTTP 200, exact content type, schema `1.1`, `success: true`, empty message, site Host fallback without port, token nullable price fields, and a per-call row shaped as:

```json
{
  "model_name": "gpt-image-2",
  "group_name": "sora",
  "price_unit": "per_call",
  "unit_price": 0.2,
  "enabled": true,
  "note": ""
}
```

- [ ] **Step 2: Run success tests and verify RED**

Run:

```bash
cd backend
go test -tags=unit ./internal/handler -run 'TestProviderPricingHandler_(PublicSuccess|SerializesModels|FallsBackToRequestHost)' -count=1
```

Expected: compilation failure because the handler does not exist.

- [ ] **Step 3: Implement the Hvoy envelope and success response**

Define an envelope with `schema_version`, `success`, `message`, and nullable `data`. Map service rows to handler DTO maps so token rows include token fields (including JSON null optionals) and per-call rows include `price_unit` and `unit_price` without an `input_price`. Set a missing `SiteDomain` from `c.Request.Host` using `url.URL{Host: host}.Hostname()`.

- [ ] **Step 4: Run success tests and verify GREEN**

Run the Step 2 command again. Expected: PASS.

- [ ] **Step 5: Write failing HMAC and service-error tests**

Cover valid `hex(HMAC-SHA256(secret, timestamp))`, missing headers, non-integer timestamp, malformed hex, wrong signature, timestamp older than 60 seconds, timestamp more than 60 seconds in the future, and exactly plus/minus 60 seconds. Assert all authentication failures return the same HTTP 401 envelope with `message: "unauthorized"` and `data: null`. Make the reader fail and assert HTTP 500 with `message: "provider pricing unavailable"`.

- [ ] **Step 6: Run HMAC/error tests and verify RED**

Run:

```bash
cd backend
go test -tags=unit ./internal/handler -run 'TestProviderPricingHandler_(HMAC|Unauthorized|ServiceError)' -count=1
```

Expected: assertions fail because authentication and error mapping are not implemented.

- [ ] **Step 7: Implement HMAC verification and error mapping**

Parse `X-Hvoy-Ts` with `strconv.ParseInt`, require `now-60 <= ts <= now+60`, decode `X-Hvoy-Sign` with `hex.DecodeString`, require 32 bytes, calculate HMAC-SHA256 over the exact timestamp header, and compare bytes with `hmac.Equal`. Log service errors with `slog.Error` without logging the secret or signature headers.

- [ ] **Step 8: Run all handler tests and verify GREEN**

Run:

```bash
cd backend
go test -tags=unit ./internal/handler -run TestProviderPricingHandler -count=1
```

Expected: PASS.

- [ ] **Step 9: Wire the handler aggregate and commit Task 3**

Add `ProviderPricing *ProviderPricingHandler` to `Handlers`, add the constructor parameter and field assignment in `ProvideHandlers`, and add `NewProviderPricingHandler` to `handler.ProviderSet`. Then commit:

```bash
git add backend/internal/handler/provider_pricing_handler.go backend/internal/handler/provider_pricing_handler_test.go backend/internal/handler/handler.go backend/internal/handler/wire.go
git commit -m "feat: expose Hvoy provider pricing handler"
```

---

### Task 4: Root Route, Wire Generation, And Deployment Configuration

**Files:**
- Create: `backend/internal/server/routes/provider_pricing.go`
- Create: `backend/internal/server/routes/provider_pricing_test.go`
- Modify: `backend/internal/server/router.go`
- Modify generated: `backend/cmd/server/wire_gen.go`
- Modify: `deploy/config.example.yaml`
- Modify: `deploy/.env.example`
- Modify: `deploy/docker-compose.yml`

**Interfaces:**
- Produces: `routes.RegisterProviderPricingRoutes(*gin.Engine, *handler.ProviderPricingHandler)`.
- Registers: exact `GET /api/provider/pricing` with no JWT/admin/API-key middleware.
- Documents and passes through: `PROVIDER_PRICING_AUTH_SECRET`.

- [ ] **Step 1: Write the failing exact-route test**

Create a unit-tagged route test that registers a real handler with a fake feed, sends `GET /api/provider/pricing`, asserts HTTP 200, and sends `POST /api/provider/pricing` plus `GET /api/v1/provider/pricing`, asserting neither resolves to the endpoint.

- [ ] **Step 2: Run route test and verify RED**

Run:

```bash
cd backend
go test -tags=unit ./internal/server/routes -run TestRegisterProviderPricingRoutes -count=1
```

Expected: compilation failure because the registrar does not exist.

- [ ] **Step 3: Add route registration**

Implement:

```go
func RegisterProviderPricingRoutes(r *gin.Engine, h *handler.ProviderPricingHandler) {
    r.GET("/api/provider/pricing", h.Get)
}
```

Call it from `registerRoutes` before creating the `/api/v1` group.

- [ ] **Step 4: Run route test and verify GREEN**

Run the Step 2 command again. Expected: PASS.

- [ ] **Step 5: Update deployment configuration surfaces**

Add to `deploy/config.example.yaml`:

```yaml
provider_pricing:
  # Empty keeps GET /api/provider/pricing public.
  # When set, callers must send X-Hvoy-Ts and X-Hvoy-Sign.
  auth_secret: ""
```

Add `PROVIDER_PRICING_AUTH_SECRET=` to `deploy/.env.example` with the same behavior documented, and add this Compose pass-through under the application service:

```yaml
- PROVIDER_PRICING_AUTH_SECRET=${PROVIDER_PRICING_AUTH_SECRET:-}
```

- [ ] **Step 6: Regenerate Wire and verify generated graph**

Run:

```bash
cd backend
go generate ./cmd/server
go test ./cmd/server -count=1
```

Expected: exit 0 and `wire_gen.go` constructs `ProviderPricingService` and `ProviderPricingHandler` before `ProvideHandlers`.

- [ ] **Step 7: Run focused contract tests**

Run:

```bash
cd backend
go test -tags=unit ./internal/config ./internal/service ./internal/handler ./internal/server/routes -run 'ProviderPricing|ListAvailable_CarriesImagePricingMetadata|UserAvailableChannel_FieldWhitelist' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit Task 4**

```bash
git add backend/internal/server/routes/provider_pricing.go backend/internal/server/routes/provider_pricing_test.go backend/internal/server/router.go backend/cmd/server/wire_gen.go deploy/config.example.yaml deploy/.env.example deploy/docker-compose.yml
git commit -m "feat: register provider pricing endpoint"
```

---

### Task 5: Full Verification And Requirement Audit

**Files:**
- Verify all files changed in Tasks 1-4.
- Compare against: `docs/superpowers/specs/2026-07-16-provider-pricing-api-design.md`.

**Interfaces:**
- Verifies the complete Hvoy contract and application build.

- [ ] **Step 1: Format and inspect the diff**

Run:

```bash
gofmt -w backend/internal/config/config.go backend/internal/config/config_test.go backend/internal/service/channel_available.go backend/internal/service/channel_available_test.go backend/internal/service/provider_pricing.go backend/internal/service/provider_pricing_test.go backend/internal/handler/available_channel_handler_test.go backend/internal/handler/provider_pricing_handler.go backend/internal/handler/provider_pricing_handler_test.go backend/internal/handler/handler.go backend/internal/handler/wire.go backend/internal/service/wire.go backend/internal/server/router.go backend/internal/server/routes/provider_pricing.go backend/internal/server/routes/provider_pricing_test.go
git diff --check
git status --short
```

Expected: `git diff --check` exits 0; status contains only planned files.

- [ ] **Step 2: Run the full backend unit suite**

Run:

```bash
cd backend
go test -tags=unit ./... -count=1
```

Expected: exit 0 with no failed package.

- [ ] **Step 3: Run the standard backend suite and build**

Run:

```bash
cd backend
go test ./... -count=1
CGO_ENABLED=0 go build -trimpath -o /tmp/sub2api-provider-pricing-server ./cmd/server
```

Expected: both commands exit 0.

- [ ] **Step 4: Audit every design requirement**

Confirm from tests/diff: exact root route, schema/version/content type, fixed CNY and 1:1 conversion, token/per-call multipliers, standard-only filtering, peak/interval/tier skips, deterministic dedup/sort, site identity/Host fallback, all HMAC boundary cases, Hvoy-shaped 401/500, internal image fields not exposed to existing DTO, no migration/UI/production data changes.

- [ ] **Step 5: Commit any verification-only corrections**

If formatting or audit required changes, rerun the affected RED/GREEN test and full verification, then commit only those corrections:

```bash
git add backend/internal/config/config.go backend/internal/config/config_test.go backend/internal/service/channel_available.go backend/internal/service/channel_available_test.go backend/internal/service/provider_pricing.go backend/internal/service/provider_pricing_test.go backend/internal/service/wire.go backend/internal/handler/available_channel_handler_test.go backend/internal/handler/provider_pricing_handler.go backend/internal/handler/provider_pricing_handler_test.go backend/internal/handler/handler.go backend/internal/handler/wire.go backend/internal/server/router.go backend/internal/server/routes/provider_pricing.go backend/internal/server/routes/provider_pricing_test.go backend/cmd/server/wire_gen.go deploy/config.example.yaml deploy/.env.example deploy/docker-compose.yml
git commit -m "test: harden provider pricing contract"
```

- [ ] **Step 6: Finish the development branch workflow**

Use `superpowers:finishing-a-development-branch`, detect normal repository/worktree state and base branch `prod`, then present the four integration options without merging, pushing, or discarding until the user selects one.
