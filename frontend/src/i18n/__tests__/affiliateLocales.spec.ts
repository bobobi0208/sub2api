import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('affiliate rebate locale copy', () => {
  it('describes rebate records as eligible consumption orders instead of top-ups', () => {
    expect(zh.admin.affiliates.records.orderAmount).toBe('消费金额')
    expect(zh.admin.affiliates.records.payAmount).toBe('实付金额')
    expect(zh.admin.affiliates.records.orderAmount).not.toContain('充值')

    expect(en.admin.affiliates.records.orderAmount).toBe('Order Amount')
    expect(en.admin.affiliates.records.orderAmount).not.toContain('Top-up')
  })
})
