package postgres

import (
	"fmt"
	"math/big"
	"strings"
)

func parseNumeric(value string) (*big.Rat, error) {
	rat := new(big.Rat)
	if _, ok := rat.SetString(strings.TrimSpace(value)); !ok {
		return nil, fmt.Errorf("invalid numeric value: %s", value)
	}
	return rat, nil
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
