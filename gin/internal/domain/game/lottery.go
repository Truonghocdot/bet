package game

import "fmt"

type LotteryPositionOutcome struct {
	Digit    int    `json:"digit"`
	BigSmall string `json:"big_small"`
	OddEven  string `json:"odd_even"`
}

type LotteryOutcome struct {
	Digits      []int                             `json:"digits"`
	Positions   map[string]LotteryPositionOutcome `json:"positions"`
	Sum         int                               `json:"sum"`
	SumBigSmall string                            `json:"sum_big_small"`
	SumOddEven  string                            `json:"sum_odd_even"`
	LastDigit   int                               `json:"last_digit"`
	Result      string                            `json:"result"`
	BigSmall    string                            `json:"big_small"`
	OddEven     string                            `json:"odd_even"`
	Tags        []string                          `json:"tags"`
}

func BuildLotteryOutcome(digits []int) (LotteryOutcome, bool) {
	if len(digits) != 5 {
		return LotteryOutcome{}, false
	}

	normalizedDigits := make([]int, 5)
	positions := make(map[string]LotteryPositionOutcome, 5)
	sum := 0
	result := ""

	for index, digit := range digits {
		if digit < 0 || digit > 9 {
			return LotteryOutcome{}, false
		}

		normalizedDigits[index] = digit
		sum += digit
		result += fmt.Sprintf("%d", digit)

		position := string(rune('a' + index))
		positionBigSmall := "small"
		if digit >= 5 {
			positionBigSmall = "big"
		}
		positionOddEven := "even"
		if digit%2 != 0 {
			positionOddEven = "odd"
		}

		positions[string(rune('A'+index))] = LotteryPositionOutcome{
			Digit:    digit,
			BigSmall: positionBigSmall,
			OddEven:  positionOddEven,
		}
		_ = position
	}

	last := normalizedDigits[4]
	lastBigSmall := "small"
	if last >= 5 {
		lastBigSmall = "big"
	}
	lastOddEven := "even"
	if last%2 != 0 {
		lastOddEven = "odd"
	}

	sumBigSmall := "small"
	if sum >= 23 {
		sumBigSmall = "big"
	}
	sumOddEven := "even"
	if sum%2 != 0 {
		sumOddEven = "odd"
	}

	tags := []string{
		fmt.Sprintf("pick5_%s", result),
		fmt.Sprintf("sum_%d", sum),
		fmt.Sprintf("last_%d", last),
		lastBigSmall,
		lastOddEven,
		fmt.Sprintf("sum_%s", sumBigSmall),
		fmt.Sprintf("sum_%s", sumOddEven),
	}

	for index, digit := range normalizedDigits {
		position := string(rune('a' + index))
		positionBigSmall := "small"
		if digit >= 5 {
			positionBigSmall = "big"
		}
		positionOddEven := "even"
		if digit%2 != 0 {
			positionOddEven = "odd"
		}

		tags = append(tags,
			fmt.Sprintf("pos_%s_%d", position, digit),
			fmt.Sprintf("pos_%s_%s", position, positionBigSmall),
			fmt.Sprintf("pos_%s_%s", position, positionOddEven),
		)
	}

	return LotteryOutcome{
		Digits:      normalizedDigits,
		Positions:   positions,
		Sum:         sum,
		SumBigSmall: sumBigSmall,
		SumOddEven:  sumOddEven,
		LastDigit:   last,
		Result:      result,
		BigSmall:    lastBigSmall,
		OddEven:     lastOddEven,
		Tags:        normalizeTags(tags),
	}, true
}
