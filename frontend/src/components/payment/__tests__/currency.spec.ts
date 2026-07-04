import { describe, expect, it } from 'vitest'
import {
  currencySymbol,
  dailySubscriptionPriceUSD,
  formatPaymentAmount,
  rechargePaymentAmountForCurrency,
  subscriptionPaymentAmountForCurrency,
  usdValueForPaymentCurrencyAmount,
} from '../currency'

describe('formatPaymentAmount', () => {
  it('uses the currency default fraction digits', () => {
    expect(formatPaymentAmount(100, 'JPY', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'KRW', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'HKD', 'en-US')).toContain('.00')
  })
})

describe('currencySymbol', () => {
  it('maps common payment currencies and falls back safely', () => {
    expect(currencySymbol('USD')).toBe('$')
    expect(currencySymbol('cny')).toBe('¥')
    expect(currencySymbol('EUR')).toBe('€')
    expect(currencySymbol('')).toBe('¥')
    expect(currencySymbol('XYZ')).toBe('XYZ')
  })
})

describe('subscriptionPaymentAmountForCurrency', () => {
  it('converts USD subscription plan prices to CNY payment amounts', () => {
    expect(subscriptionPaymentAmountForCurrency(1, 'CNY')).toBe(6.8)
    expect(subscriptionPaymentAmountForCurrency(1.91, 'CNY')).toBe(12.99)
  })

  it('keeps USD subscription plan prices for USD payments', () => {
    expect(subscriptionPaymentAmountForCurrency(1.91, 'USD')).toBe(1.91)
  })
})

describe('rechargePaymentAmountForCurrency', () => {
  it('converts USD recharge amounts to CNY gateway payment amounts', () => {
    expect(rechargePaymentAmountForCurrency(1, 'CNY')).toBe(6.8)
    expect(rechargePaymentAmountForCurrency(1.91, 'CNY')).toBe(12.99)
  })

  it('keeps USD recharge amounts for USD payment providers', () => {
    expect(rechargePaymentAmountForCurrency(1.91, 'USD')).toBe(1.91)
  })
})

describe('usdValueForPaymentCurrencyAmount', () => {
  it('converts CNY payment limits back to USD input amounts', () => {
    expect(usdValueForPaymentCurrencyAmount(6.8, 'CNY')).toBe(1)
    expect(usdValueForPaymentCurrencyAmount(12.99, 'CNY')).toBe(1.91)
  })
})

describe('dailySubscriptionPriceUSD', () => {
  it('returns the USD price divided by validity days', () => {
    expect(dailySubscriptionPriceUSD(12.5, 7)).toBe(1.79)
    expect(dailySubscriptionPriceUSD(50, 30)).toBe(1.67)
  })
})
