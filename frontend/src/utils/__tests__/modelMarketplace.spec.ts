import { describe, expect, it } from 'vitest'

import {
  buildModelMarketplace,
  getCardEffectiveRate,
  getMarketplacePrice,
} from '../modelMarketplace'
import type { UserAvailableChannel } from '@/api/channels'

const channels: UserAvailableChannel[] = [
  {
    name: 'Primary',
    description: 'primary channel',
    platforms: [
      {
        platform: 'anthropic',
        groups: [
          {
            id: 1,
            name: 'Public Claude',
            platform: 'anthropic',
            subscription_type: 'standard',
            rate_multiplier: 1,
            is_exclusive: false,
          },
          {
            id: 2,
            name: 'VIP Claude',
            platform: 'anthropic',
            subscription_type: 'subscription',
            rate_multiplier: 0.6,
            is_exclusive: true,
          },
        ],
        supported_models: [
          {
            name: 'claude-opus-4',
            platform: 'anthropic',
            pricing: {
              billing_mode: 'token',
              input_price: 0.000005,
              output_price: 0.000025,
              cache_write_price: 0.00000625,
              cache_read_price: 0.0000005,
              image_output_price: null,
              per_request_price: null,
              intervals: [],
            },
          },
        ],
      },
      {
        platform: 'openai',
        groups: [
          {
            id: 3,
            name: 'GPT',
            platform: 'openai',
            subscription_type: 'standard',
            rate_multiplier: 0.3,
            is_exclusive: false,
          },
        ],
        supported_models: [
          {
            name: 'gpt-4.1',
            platform: 'openai',
            pricing: {
              billing_mode: 'token',
              input_price: 0.000002,
              output_price: 0.000008,
              cache_write_price: null,
              cache_read_price: 0.0000002,
              image_output_price: null,
              per_request_price: null,
              intervals: [],
            },
          },
        ],
      },
    ],
  },
  {
    name: 'Backup',
    description: '',
    platforms: [
      {
        platform: 'anthropic',
        groups: [
          {
            id: 4,
            name: 'Discount Claude',
            platform: 'anthropic',
            subscription_type: 'standard',
            rate_multiplier: 0.8,
            is_exclusive: false,
          },
        ],
        supported_models: [
          {
            name: 'claude-opus-4',
            platform: 'anthropic',
            pricing: {
              billing_mode: 'token',
              input_price: 0.000005,
              output_price: 0.000025,
              cache_write_price: 0.00000625,
              cache_read_price: 0.0000005,
              image_output_price: null,
              per_request_price: null,
              intervals: [],
            },
          },
        ],
      },
    ],
  },
]

describe('buildModelMarketplace', () => {
  it('dedupes models, aggregates groups, and counts platforms', () => {
    const result = buildModelMarketplace(channels, { 2: 0.4 })

    expect(result.cards.map((card) => `${card.platform}:${card.name}`)).toEqual([
      'anthropic:claude-opus-4',
      'openai:gpt-4.1',
    ])
    expect(result.cards[0].groups.map((group) => group.id)).toEqual([1, 2, 4])
    expect(result.groups.map((group) => group.id)).toEqual([1, 2, 3, 4])
    expect(result.platforms).toEqual([
      { platform: 'anthropic', count: 1 },
      { platform: 'openai', count: 1 },
    ])
    expect(result.totalModels).toBe(2)
  })

  it('uses user group rates before group defaults when calculating displayed prices', () => {
    const result = buildModelMarketplace(channels, { 2: 0.4 })
    const claude = result.cards[0]

    expect(getCardEffectiveRate(claude, 'all')).toEqual({
      rate: 0.4,
      label: 'VIP Claude',
    })
    expect(getCardEffectiveRate(claude, 4)).toEqual({
      rate: 0.8,
      label: 'Discount Claude',
    })
    expect(getMarketplacePrice(claude, 'input_price', 'all', true)).toBe(0.000002)
    expect(getMarketplacePrice(claude, 'input_price', 'all', false)).toBe(0.000005)
  })
})
