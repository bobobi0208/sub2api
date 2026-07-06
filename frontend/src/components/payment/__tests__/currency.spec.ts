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
  it('uses one-to-one CNY payment amounts for USD subscription plan prices', () => {
    expect(subscriptionPaymentAmountForCurrency(1, 'CNY')).toBe(1)
    expect(subscriptionPaymentAmountForCurrency(1.91, 'CNY')).toBe(1.91)
  })

  it('keeps USD subscription plan prices for USD payments', () => {
    expect(subscriptionPaymentAmountForCurrency(1.91, 'USD')).toBe(1.91)
  })
})

describe('rechargePaymentAmountForCurrency', () => {
  it('uses one-to-one CNY gateway payment amounts for USD recharge amounts', () => {
    expect(rechargePaymentAmountForCurrency(1, 'CNY')).toBe(1)
    expect(rechargePaymentAmountForCurrency(1.91, 'CNY')).toBe(1.91)
  })

  it('keeps USD recharge amounts for USD payment providers', () => {
    expect(rechargePaymentAmountForCurrency(1.91, 'USD')).toBe(1.91)
  })
})

describe('usdValueForPaymentCurrencyAmount', () => {
  it('uses one-to-one CNY payment limits as USD input amounts', () => {
    expect(usdValueForPaymentCurrencyAmount(1, 'CNY')).toBe(1)
    expect(usdValueForPaymentCurrencyAmount(12.99, 'CNY')).toBe(12.99)
  })
})

describe('dailySubscriptionPriceUSD', () => {
  it('returns the USD price divided by validity days', () => {
    expect(dailySubscriptionPriceUSD(12.5, 7)).toBe(1.79)
    expect(dailySubscriptionPriceUSD(50, 30)).toBe(1.67)
  })
})
