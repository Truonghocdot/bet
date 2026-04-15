package phone

import (
	"strings"
	"unicode"
)

func NormalizeVNPhone(input string) string {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return ""
	}

	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, "-", "")

	if strings.HasPrefix(raw, "+") {
		digits := stripNonDigits(raw[1:])
		if digits == "" {
			return ""
		}
		if strings.HasPrefix(digits, "84") {
			return "+84" + strings.TrimLeft(digits[2:], "0")
		}
		return "+" + digits
	}

	digits := stripNonDigits(raw)
	if digits == "" {
		return ""
	}

	if strings.HasPrefix(digits, "84") {
		return "+84" + strings.TrimLeft(digits[2:], "0")
	}

	if strings.HasPrefix(digits, "0") {
		return "+84" + strings.TrimLeft(digits[1:], "0")
	}

	return "+84" + strings.TrimLeft(digits, "0")
}

func VNPhoneVariants(input string) []string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}

	variants := make([]string, 0, 4)
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		for _, existing := range variants {
			if strings.EqualFold(existing, value) {
				return
			}
		}
		variants = append(variants, value)
	}

	add(trimmed)
	add(strings.ReplaceAll(trimmed, " ", ""))
	add(strings.ReplaceAll(trimmed, "-", ""))

	normalized := NormalizeVNPhone(trimmed)
	add(normalized)

	digits := stripNonDigits(trimmed)
	if digits != "" {
		add(digits)
		if strings.HasPrefix(digits, "84") {
			add("0" + strings.TrimLeft(digits[2:], "0"))
		} else if strings.HasPrefix(digits, "0") {
			add("+84" + strings.TrimLeft(digits[1:], "0"))
		} else {
			add("0" + strings.TrimLeft(digits, "0"))
		}
	}

	return variants
}

func stripNonDigits(input string) string {
	var b strings.Builder
	for _, r := range input {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
