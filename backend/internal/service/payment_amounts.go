package service

import (
	"math"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/shopspring/decimal"
)

const defaultBalanceRechargeMultiplier = 1.0
const minBalanceRechargeUSD = 1.0
const usdToCNYExchangeRate = 1.0

func normalizeBalanceRechargeMultiplier(multiplier float64) float64 {
	if math.IsNaN(multiplier) || math.IsInf(multiplier, 0) || multiplier <= 0 {
		return defaultBalanceRechargeMultiplier
	}
	return multiplier
}

func calculateCreditedBalance(amountUSD float64) float64 {
	return decimal.NewFromFloat(amountUSD).Round(2).InexactFloat64()
}

func calculateSubscriptionPaymentBaseAmount(planPriceUSD float64, currency string) float64 {
	return calculateUSDValuePaymentBaseAmount(planPriceUSD, currency)
}

func calculateBalanceRechargePaymentBaseAmount(amountUSD float64, currency string) float64 {
	return calculateUSDValuePaymentBaseAmount(amountUSD, currency)
}

func calculateUSDValuePaymentBaseAmount(amountUSD float64, currency string) float64 {
	normalizedCurrency, err := payment.NormalizePaymentCurrency(currency)
	if err != nil {
		normalizedCurrency = payment.DefaultPaymentCurrency
	}
	amount := decimal.NewFromFloat(amountUSD)
	if normalizedCurrency == payment.DefaultPaymentCurrency {
		amount = amount.Mul(decimal.NewFromFloat(usdToCNYExchangeRate))
	}
	return amount.Round(int32(payment.CurrencyMaxFractionDigits(normalizedCurrency))).InexactFloat64()
}

func calculateGatewayRefundAmount(orderAmount, payAmount, refundAmount float64, currency string) float64 {
	if orderAmount <= 0 || payAmount <= 0 || refundAmount <= 0 {
		return 0
	}
	fractionDigits := int32(payment.CurrencyMaxFractionDigits(currency))
	if math.Abs(refundAmount-orderAmount) <= paymentAmountToleranceForCurrency(currency) {
		return decimal.NewFromFloat(payAmount).Round(fractionDigits).InexactFloat64()
	}
	return decimal.NewFromFloat(payAmount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(orderAmount)).
		Round(fractionDigits).
		InexactFloat64()
}
