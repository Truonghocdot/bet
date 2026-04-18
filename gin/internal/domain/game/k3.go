package game

import (
	"fmt"
	"sort"
	"strings"
)

type K3Outcome struct {
	Dice     []int    `json:"dice"`
	Sum      int      `json:"sum"`
	Result   string   `json:"result"`
	BigSmall string   `json:"big_small"`
	OddEven  string   `json:"odd_even"`
	IsTriple bool     `json:"is_triple"`
	Tags     []string `json:"tags"`
}

func BuildK3Outcome(dice []int) (K3Outcome, bool) {
	if len(dice) != 3 {
		return K3Outcome{}, false
	}

	normalizedDice := make([]int, 3)
	counts := make(map[int]int, 3)
	sum := 0
	for i, value := range dice {
		if value < 1 || value > 6 {
			return K3Outcome{}, false
		}
		normalizedDice[i] = value
		counts[value]++
		sum += value
	}

	bigSmall := "small"
	if sum >= 11 {
		bigSmall = "big"
	}

	oddEven := "even"
	if sum%2 != 0 {
		oddEven = "odd"
	}

	tags := []string{
		fmt.Sprintf("sum_%d", sum),
		bigSmall,
		oddEven,
	}

	isTriple := len(counts) == 1
	switch len(counts) {
	case 1:
		tags = append(tags, "triple_any", fmt.Sprintf("triple_%d", normalizedDice[0]))
	case 2:
		for value, count := range counts {
			if count == 2 {
				tags = append(tags, fmt.Sprintf("pair_%d", value))
				break
			}
		}
	case 3:
		sorted := append([]int(nil), normalizedDice...)
		sort.Ints(sorted)
		if sorted[0]+1 == sorted[1] && sorted[1]+1 == sorted[2] {
			tags = append(tags, "serial_any")
		}
		for _, value := range sorted {
			tags = append(tags, fmt.Sprintf("diff_%d", value))
		}
	}

	return K3Outcome{
		Dice:     normalizedDice,
		Sum:      sum,
		Result:   fmt.Sprintf("%d-%d-%d", normalizedDice[0], normalizedDice[1], normalizedDice[2]),
		BigSmall: bigSmall,
		OddEven:  oddEven,
		IsTriple: isTriple,
		Tags:     normalizeTags(tags),
	}, true
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
