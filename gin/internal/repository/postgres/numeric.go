package postgres

import (
	"fmt"
	"math/big"
	"strings"
)

const (
	moneyPrecision  = 30
	moneyScale      = 8
	minimumBetStake = "1000"
)

var (
	moneyScaleFactor           = new(big.Int).Exp(big.NewInt(10), big.NewInt(moneyScale), nil)
	moneyStorageExclusiveLimit = new(big.Int).Exp(big.NewInt(10), big.NewInt(moneyPrecision-moneyScale), nil)
)

func parseNumeric(value string) (*big.Rat, error) {
	rat := new(big.Rat)
	if _, ok := rat.SetString(strings.TrimSpace(value)); !ok {
		return nil, fmt.Errorf("invalid numeric value: %s", value)
	}
	return rat, nil
}

func normalizeMoneyForStorage(value string) (string, error) {
	decimal := strings.TrimSpace(value)
	if !isOrdinaryDecimal(decimal) {
		return "", fmt.Errorf("invalid money decimal value: %q", value)
	}

	amount, err := parseNumeric(decimal)
	if err != nil {
		return "", err
	}

	scaled := new(big.Rat).Mul(amount, new(big.Rat).SetInt(moneyScaleFactor))
	if !scaled.IsInt() {
		return "", fmt.Errorf("money value has more than %d decimal places: %s", moneyScale, value)
	}

	absAmount := new(big.Rat).Abs(new(big.Rat).Set(amount))
	limit := new(big.Rat).SetInt(moneyStorageExclusiveLimit)
	if absAmount.Cmp(limit) >= 0 {
		return "", fmt.Errorf("money value exceeds numeric(%d,%d): %s", moneyPrecision, moneyScale, value)
	}

	return amount.FloatString(moneyScale), nil
}

func isOrdinaryDecimal(value string) bool {
	if value == "" {
		return false
	}

	start := 0
	if value[0] == '+' || value[0] == '-' {
		start++
		if start == len(value) {
			return false
		}
	}

	dot := -1
	for index := start; index < len(value); index++ {
		switch {
		case value[index] >= '0' && value[index] <= '9':
		case value[index] == '.' && dot == -1:
			dot = index
		default:
			return false
		}
	}

	return dot != start && dot != len(value)-1
}

func normalizeBetStake(value string) (string, error) {
	normalized, err := normalizeMoneyForStorage(value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidBetAmount, err)
	}
	if compareNumeric(normalized, minimumBetStake) < 0 {
		return "", fmt.Errorf("%w: minimum stake is %s", ErrInvalidBetAmount, minimumBetStake)
	}
	return normalized, nil
}

func normalizePendingPayoutExposure(pendingPotentialPayout, newPotentialPayout string) (string, error) {
	exposure, err := addNumeric(pendingPotentialPayout, newPotentialPayout)
	if err != nil {
		return "", err
	}

	return normalizeMoneyForStorage(exposure)
}

func addNumeric(left, right string) (string, error) {
	lv, err := parseNumeric(left)
	if err != nil {
		return "", err
	}
	rv, err := parseNumeric(right)
	if err != nil {
		return "", err
	}
	return new(big.Rat).Add(lv, rv).FloatString(8), nil
}

func AddNumeric(left, right string) (string, error) {
	return addNumeric(left, right)
}

func subtractNumeric(left, right string) (string, error) {
	lv, err := parseNumeric(left)
	if err != nil {
		return "", err
	}
	rv, err := parseNumeric(right)
	if err != nil {
		return "", err
	}
	return new(big.Rat).Sub(lv, rv).FloatString(8), nil
}

func SubtractNumeric(left, right string) (string, error) {
	return subtractNumeric(left, right)
}

func multiplyNumeric(left, right string) (string, error) {
	lv, err := parseNumeric(left)
	if err != nil {
		return "", err
	}
	rv, err := parseNumeric(right)
	if err != nil {
		return "", err
	}
	return new(big.Rat).Mul(lv, rv).FloatString(8), nil
}

func compareNumeric(left, right string) int {
	lv, err := parseNumeric(left)
	if err != nil {
		return -1
	}
	rv, err := parseNumeric(right)
	if err != nil {
		return -1
	}
	return lv.Cmp(rv)
}

func calculateBetTaxAndNet(amount string) (string, string, error) {
	original, err := parseNumeric(amount)
	if err != nil {
		return "", "", err
	}
	if original.Sign() <= 0 {
		return "", "", fmt.Errorf("invalid bet amount: %s", amount)
	}

	taxRate := big.NewRat(2, 100)
	tax := new(big.Rat).Mul(original, taxRate)
	net := new(big.Rat).Sub(original, tax)
	if net.Sign() <= 0 {
		return "", "", fmt.Errorf("net bet amount must be positive")
	}

	return tax.FloatString(8), net.FloatString(8), nil
}
