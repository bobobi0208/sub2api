# Provider Pricing API Design

## Goal

Expose `GET /api/provider/pricing` using Hvoy schema version `1.1`, so Hvoy can fetch this site's active pay-as-you-go model prices in CNY.

The published price is the price a normal group user is charged:

```text
token price = internal USD/token price * 1,000,000 * group multiplier
per-call price = internal USD/call price * group multiplier
```

The endpoint uses a fixed conversion of `1 USD = 1 CNY`. The multiplication by `1,000,000` only converts the internal per-token unit into Hvoy's per-one-million-token unit.

## Non-Goals

- Do not add an admin UI.
- Do not create a separate public-pricing table.
- Do not publish subscription groups.
- Do not approximate interval, resolution-tier, peak-period, or other non-unique prices.
- Do not change channel pricing, model mappings, group configuration, or production data.
- Do not support Hvoy schema versions other than `1.1` in this change.

## External Contract

The route is public and is registered at the exact path:

```text
GET /api/provider/pricing
```

Successful responses use `application/json; charset=utf-8` and this envelope:

```json
{
  "schema_version": "1.1",
  "success": true,
  "message": "",
  "data": {
    "currency": "CNY",
    "price_unit": "per_1m_tokens",
    "site_name": "Example Site",
    "site_domain": "api.example.com",
    "updated_at": "2026-07-16T12:00:00Z",
    "models": []
  }
}
```

`site_name` comes from the existing public site-name setting. `site_domain` is the hostname from the existing API base URL; when that setting is empty or invalid, the handler uses the current request host after stripping its port.

`updated_at` is the UTC generation time of the response. Models are sorted by lowercase `group_name`, then lowercase `model_name`, to keep the response deterministic.

## Data Source

The provider-pricing service reuses `ChannelService.ListAvailable`. This preserves the existing channel model identity rules:

- supported models are the union of channel model mappings and explicit channel pricing;
- mapped models use the billing model's pricing;
- missing explicit channel prices may use the existing global pricing fallback;
- only active groups attached to a channel appear in the available-channel view.

`AvailableGroupRef` is extended with internal image-rate and explicit image-tier metadata so the provider service can decide whether a unique image price exists. The existing user-facing available-channel DTO remains unchanged and does not expose the additional fields.

The service then applies provider-specific filtering:

1. Keep active channels only.
2. Keep active groups whose `subscription_type` is the standard pay-as-you-go type.
3. Skip peak-rate-enabled groups because they do not have one stable public price.
4. Match models only to groups with the same platform.
5. Skip pricing entries that contain intervals or otherwise require a tier choice.
6. Skip entries whose billing mode cannot be represented by Hvoy.
7. Deduplicate by lowercase `model_name + group_name`.

Exclusive standard groups are not implicitly removed. The selected scope is all active pay-as-you-go groups, not a separate publication allowlist.

## Price Mapping

### Token Billing

Token billing maps to the default response unit `per_1m_tokens`.

`input_price` is required. Entries without an input price are skipped. Optional prices are emitted as JSON `null` when unavailable.

| Hvoy field | Sub2API source | Transformation |
| --- | --- | --- |
| `input_price` | `InputPrice` | `value * 1_000_000 * group multiplier` |
| `output_price` | `OutputPrice` | `value * 1_000_000 * group multiplier` |
| `cache_input_price` | `CacheReadPrice` | `value * 1_000_000 * group multiplier` |
| `cache_create_price` | `CacheWritePrice` | `value * 1_000_000 * group multiplier` |
| `cache_create_price_1h` | unavailable in channel price shape | `null` |

### Per-Request And Image Billing

An entry with a unique flat `PerRequestPrice` maps to:

```json
{
  "price_unit": "per_call",
  "unit_price": 0.2
}
```

Ordinary per-request prices use the group token multiplier. Image prices use the independent image multiplier when the group enables independent image rates; otherwise they use the group token multiplier.

Image entries are skipped when the group has explicit resolution prices or the channel price has intervals, because those prices cannot be represented as one Hvoy `unit_price`.

Video and other unsupported billing modes are skipped.

All numeric outputs are rounded to at most eight decimal places to avoid float serialization artifacts. Prices are never silently replaced with zero.

Every included row has `enabled: true` and `note: ""`, because inactive channel/group rows are filtered out before transformation.

## Optional HMAC Authentication

Add a backend configuration section:

```yaml
provider_pricing:
  auth_secret: ""
```

The corresponding environment variable is:

```text
PROVIDER_PRICING_AUTH_SECRET
```

Behavior:

- Empty secret: the endpoint is public and signature headers are ignored.
- Non-empty secret: both `X-Hvoy-Ts` and `X-Hvoy-Sign` are required.
- `X-Hvoy-Ts` must be a Unix-seconds integer within 60 seconds of server time.
- The expected signature is lowercase hex `HMAC-SHA256(auth_secret, X-Hvoy-Ts)`.
- Compare decoded signature bytes using `hmac.Equal` to avoid timing-sensitive string comparison.

Authentication failures return HTTP `401`. They do not reveal whether the timestamp, signature format, or signature value was wrong.

## Error Handling

Errors retain the Hvoy envelope shape:

```json
{
  "schema_version": "1.1",
  "success": false,
  "message": "unauthorized",
  "data": null
}
```

- Invalid or missing required signature: `401 unauthorized`.
- Channel/group/settings lookup failure: `500 provider pricing unavailable`.
- A malformed individual pricing entry is skipped rather than failing the entire feed.

Internal errors are logged without the configured HMAC secret or request signatures.

## Components

### Service

A focused provider-pricing service owns filtering, price conversion, deduplication, and deterministic sorting. It depends on the existing channel and setting services rather than repositories or SQL.

The service returns protocol-neutral response data plus site identity. It does not depend on Gin and accepts a clock so time-dependent tests are deterministic.

### Handler

The handler owns HTTP headers, request-host fallback, optional HMAC verification, status codes, and JSON serialization.

### Routing And Wiring

The handler is added to the root `Handlers` aggregate and Wire provider set. A small public route registrar adds the exact `/api/provider/pricing` route outside `/api/v1`.

No database migration is needed.

## Test Strategy

Service tests cover:

- token prices multiplied by one million and the group multiplier;
- fixed 1:1 CNY output;
- nullable output and cache prices;
- unique flat per-call prices;
- independent image multiplier behavior;
- inactive channels, subscription groups, peak groups, interval prices, and incomplete prices being skipped;
- platform matching, deduplication, and stable sorting;
- site-name and API-domain extraction.

Handler tests cover:

- public access when no secret is configured;
- valid HMAC access;
- missing headers, malformed timestamp/signature, expired/future timestamp, and incorrect signature returning `401`;
- Hvoy success and error envelopes;
- request-host fallback and content type.

Route/config tests cover:

- exact registration of `GET /api/provider/pricing`;
- environment loading for `PROVIDER_PRICING_AUTH_SECRET`.

The focused tests run first, followed by the complete backend test suite and backend build.

## Known Production Data Dependency

The feed intentionally uses the same supported-model source as the model marketplace. Production channel ID 2 currently has no channel mapping or pricing identity for `gpt-5.6-sol`, `gpt-5.6-terra`, or `gpt-5.6-luna`, even though active accounts and the global pricing catalog support them. Those models will remain absent from both the marketplace and this feed until the channel configuration is updated. That production-data correction is deliberately outside this code change.
