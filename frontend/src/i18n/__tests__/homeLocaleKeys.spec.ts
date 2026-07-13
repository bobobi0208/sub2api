import { describe, expect, it } from 'vitest'

import enLanding from '../locales/en/landing'
import zhLanding from '../locales/zh/landing'

const requiredPaths = [
  'stats.subtitle',
  'stats.uptime',
  'stats.uptimeLabel',
  'stats.latency',
  'stats.latencyLabel',
  'stats.requests',
  'stats.requestsLabel',
  'stats.monitor',
  'stats.monitorLabel',
  'pricing.badge',
  'pricing.title',
  'pricing.subtitle',
  'pricing.popular',
  'pricing.cta',
  'pricing.perDay',
  'pricing.daysUnit',
  'pricing.guarantee'
]

function getPath(root: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>(
    (value, key) =>
      value && typeof value === 'object'
        ? (value as Record<string, unknown>)[key]
        : undefined,
    root
  )
}

describe.each([
  ['zh', zhLanding.home],
  ['en', enLanding.home]
])('home locale %s', (_, home) => {
  it.each(requiredPaths)('defines %s', (path) => {
    expect(getPath(home, path)).toEqual(expect.any(String))
    expect(getPath(home, path)).not.toBe('')
  })
})

it('uses the approved subscription hero copy', () => {
  expect(zhLanding.home.heroSubtitle).toBe('一次订阅，畅享专属 AI 额度')
  expect(enLanding.home.heroSubtitle).toBe('One Subscription, Dedicated AI Capacity')
})
