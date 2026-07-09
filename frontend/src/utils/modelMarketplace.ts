import type {
  UserAvailableChannel,
  UserAvailableGroup,
  UserSupportedModel,
  UserSupportedModelPricing,
} from '@/api/channels'

export type MarketplaceGroup = UserAvailableGroup & {
  effective_rate_multiplier: number
}

export interface MarketplaceModelCard {
  id: string
  name: string
  platform: string
  channels: string[]
  groups: MarketplaceGroup[]
  pricing: UserSupportedModelPricing | null
  hasPromptCaching: boolean
}

export interface MarketplacePlatformStat {
  platform: string
  count: number
}

export interface ModelMarketplaceData {
  cards: MarketplaceModelCard[]
  groups: MarketplaceGroup[]
  platforms: MarketplacePlatformStat[]
  totalModels: number
}

export type MarketplaceGroupFilter = 'all' | number

const MARKETPLACE_PLATFORM_ORDER = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']
const FEATURED_PLATFORM_LIMIT = 3

export type MarketplacePriceKey =
  | 'input_price'
  | 'output_price'
  | 'cache_write_price'
  | 'cache_read_price'
  | 'image_output_price'
  | 'per_request_price'

export function buildModelMarketplace(
  channels: UserAvailableChannel[],
  userGroupRates: Record<number, number> = {},
): ModelMarketplaceData {
  const cardMap = new Map<string, MarketplaceModelCard>()
  const groupMap = new Map<number, MarketplaceGroup>()

  for (const channel of channels) {
    for (const section of channel.platforms) {
      const marketplaceGroups = section.groups.map((group) =>
        toMarketplaceGroup(group, userGroupRates),
      )
      for (const group of marketplaceGroups) {
        if (!groupMap.has(group.id)) groupMap.set(group.id, group)
      }

      for (const model of section.supported_models) {
        const platform = model.platform || section.platform
        const key = modelKey(platform, model.name)
        const existing = cardMap.get(key)
        if (existing) {
          existing.channels = appendUnique(existing.channels, channel.name)
          existing.groups = mergeGroups(existing.groups, marketplaceGroups)
          if (!existing.pricing && model.pricing) existing.pricing = model.pricing
          existing.hasPromptCaching = existing.hasPromptCaching || hasPromptCaching(model)
          continue
        }

        cardMap.set(key, {
          id: key,
          name: model.name,
          platform,
          channels: [channel.name],
          groups: marketplaceGroups,
          pricing: model.pricing,
          hasPromptCaching: hasPromptCaching(model),
        })
      }
    }
  }

  const cards = Array.from(cardMap.values()).sort(compareCards)
  const platforms = buildPlatformStats(cards)

  return {
    cards,
    groups: Array.from(groupMap.values()),
    platforms,
    totalModels: cards.length,
  }
}

export function getCardEffectiveRate(
  card: MarketplaceModelCard,
  selectedGroupId: MarketplaceGroupFilter,
): { rate: number; label: string } | null {
  const groups =
    selectedGroupId === 'all'
      ? card.groups
      : card.groups.filter((group) => group.id === selectedGroupId)
  if (groups.length === 0) return null

  const best = [...groups].sort((a, b) => {
    if (a.effective_rate_multiplier !== b.effective_rate_multiplier) {
      return a.effective_rate_multiplier - b.effective_rate_multiplier
    }
    return a.name.localeCompare(b.name)
  })[0]

  return {
    rate: best.effective_rate_multiplier,
    label: best.name,
  }
}

export function getMarketplacePrice(
  card: MarketplaceModelCard,
  priceKey: MarketplacePriceKey,
  selectedGroupId: MarketplaceGroupFilter,
  showRatePrice: boolean,
): number | null {
  const value = card.pricing?.[priceKey]
  if (value == null) return null
  if (!showRatePrice) return value
  const effective = getCardEffectiveRate(card, selectedGroupId)
  return effective ? roundPrice(value * effective.rate) : value
}

export function getFeaturedPlatformStats(
  platforms: MarketplacePlatformStat[],
): MarketplacePlatformStat[] {
  return platforms.slice(0, FEATURED_PLATFORM_LIMIT)
}

function toMarketplaceGroup(
  group: UserAvailableGroup,
  userGroupRates: Record<number, number>,
): MarketplaceGroup {
  return {
    ...group,
    effective_rate_multiplier: userGroupRates[group.id] ?? group.rate_multiplier,
  }
}

function modelKey(platform: string, name: string): string {
  return `${platform || 'unknown'}:${name}`
}

function hasPromptCaching(model: UserSupportedModel): boolean {
  const pricing = model.pricing
  if (!pricing) return false
  return (
    (pricing.cache_write_price != null && pricing.cache_write_price > 0) ||
    (pricing.cache_read_price != null && pricing.cache_read_price > 0)
  )
}

function appendUnique(values: string[], next: string): string[] {
  return values.includes(next) ? values : [...values, next]
}

function mergeGroups(
  current: MarketplaceGroup[],
  incoming: MarketplaceGroup[],
): MarketplaceGroup[] {
  const map = new Map<number, MarketplaceGroup>()
  for (const group of current) map.set(group.id, group)
  for (const group of incoming) map.set(group.id, group)
  return Array.from(map.values())
}

function compareCards(a: MarketplaceModelCard, b: MarketplaceModelCard): number {
  const platformResult = comparePlatforms(a.platform, b.platform)
  if (platformResult !== 0) return platformResult
  return a.name.localeCompare(b.name)
}

function roundPrice(value: number): number {
  return Number(value.toPrecision(12))
}

function buildPlatformStats(cards: MarketplaceModelCard[]): MarketplacePlatformStat[] {
  const counts = new Map<string, number>()
  for (const card of cards) {
    counts.set(card.platform, (counts.get(card.platform) ?? 0) + 1)
  }
  return Array.from(counts.entries())
    .map(([platform, count]) => ({ platform, count }))
    .sort((a, b) => {
      if (a.count !== b.count) return b.count - a.count
      return comparePlatforms(a.platform, b.platform)
    })
}

function comparePlatforms(a: string, b: string): number {
  const ai = MARKETPLACE_PLATFORM_ORDER.indexOf(a)
  const bi = MARKETPLACE_PLATFORM_ORDER.indexOf(b)
  if (ai === -1 && bi === -1) return a.localeCompare(b)
  if (ai === -1) return 1
  if (bi === -1) return -1
  return ai - bi
}
