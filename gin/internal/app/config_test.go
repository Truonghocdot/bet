package app

import "testing"

func TestLoadConfigEngineSettlementEnabled(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{name: "defaults enabled", value: " ", expected: true},
		{name: "explicitly disabled", value: "false", expected: false},
		{name: "explicitly enabled", value: "true", expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("ENGINE_SETTLEMENT_ENABLED", test.value)

			config := LoadConfig()
			if config.EngineSettlementEnabled != test.expected {
				t.Fatalf("expected settlement enabled=%t, got %t", test.expected, config.EngineSettlementEnabled)
			}
		})
	}
}
