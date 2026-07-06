export const DEFAULT_PAYMENT_CURRENCY = 'CNY'
export const USD_TO_CNY_EXCHANGE_RATE = 1
export const SUBSCRIPTION_USD_TO_CNY_EXCHANGE_RATE = USD_TO_CNY_EXCHANGE_RATE

const PAYMENT_CURRENCY_SYMBOLS: Record<string, string> = {
  USD: '$',
  CNY: '¥',
  RMB: '¥',
  EUR: '€',
  GBP: '£',
  JPY: '¥',
  HKD: 'HK$',
  TWD: 'NT$',
  KRW: '₩',
  AUD: 'A$',
  CAD: 'C$',
  SGD: 'S$',
  NZD: 'NZ$',
  MOP: 'MOP$',
  MYR: 'RM',
  THB: '฿',
  PHP: '₱',
  INR: '₹',
}

export function normalizePaymentCurrency(currency?: string | null): string {
  const normalized = String(currency || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(normalized) ? normalized : DEFAULT_PAYMENT_CURRENCY
}

export function currencySymbol(currency?: string | null): string {
  const normalized = normalizePaymentCurrency(currency)
  return PAYMENT_CURRENCY_SYMBOLS[normalized] || normalized
}

function paymentCurrencyFractionDigits(currency: string): number {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
    }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

function roundPaymentCurrencyAmount(amount: number, currency: string): number {
  const fractionDigits = paymentCurrencyFractionDigits(currency)
  const factor = 10 ** fractionDigits
  return Math.round((Number.isFinite(amount) ? amount : 0) * factor) / factor
}

export function subscriptionPaymentAmountForCurrency(planPriceUSD: number, currency?: string | null): number {
  return paymentAmountForUSDValue(planPriceUSD, currency)
}

export function rechargePaymentAmountForCurrency(amountUSD: number, currency?: string | null): number {
  return paymentAmountForUSDValue(amountUSD, currency)
}

export function usdValueForPaymentCurrencyAmount(amount: number, currency?: string | null): number {
  const normalized = normalizePaymentCurrency(currency)
  const paymentAmount = Number.isFinite(amount) ? amount : 0
  const usdAmount = normalized === 'CNY' || normalized === 'RMB'
    ? paymentAmount / USD_TO_CNY_EXCHANGE_RATE
    : paymentAmount
  return roundPaymentCurrencyAmount(usdAmount, 'USD')
}

export function dailySubscriptionPriceUSD(planPriceUSD: number, validityDays: number): number {
  const planPrice = Number.isFinite(planPriceUSD) ? planPriceUSD : 0
  const days = Number.isFinite(validityDays) && validityDays > 0 ? validityDays : 1
  return roundPaymentCurrencyAmount(planPrice / days, 'USD')
}

function paymentAmountForUSDValue(amountUSD: number, currency?: string | null): number {
  const normalized = normalizePaymentCurrency(currency)
  const planPrice = Number.isFinite(amountUSD) ? amountUSD : 0
  const paymentAmount = normalized === 'CNY' || normalized === 'RMB'
    ? planPrice * USD_TO_CNY_EXCHANGE_RATE
    : planPrice
  return roundPaymentCurrencyAmount(paymentAmount, normalized)
}

export function formatPaymentAmount(amount: number, currency?: string | null, locale?: string): string {
  const normalized = normalizePaymentCurrency(currency)
  const fractionDigits = paymentCurrencyFractionDigits(normalized)
  try {
    return new Intl.NumberFormat(locale || undefined, {
      style: 'currency',
      currency: normalized,
      currencyDisplay: 'narrowSymbol',
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    }).format(Number.isFinite(amount) ? amount : 0)
  } catch {
    return `${normalized} ${(Number.isFinite(amount) ? amount : 0).toFixed(fractionDigits)}`
  }
}
