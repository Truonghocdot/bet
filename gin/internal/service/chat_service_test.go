package service

import "testing"

func TestValidChatBody(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		valid bool
	}{
		{name: "plain text", body: "Chuc moi nguoi choi vui ve", valid: true},
		{name: "empty", body: "   ", valid: false},
		{name: "http URL", body: "http://example.com", valid: false},
		{name: "bare domain", body: "vao example.xyz ngay", valid: false},
		{name: "www URL", body: "www.example.vn", valid: false},
		{name: "too long", body: string(make([]rune, 281)), valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validChatBody(test.body); got != test.valid {
				t.Fatalf("validChatBody(%q) = %t, want %t", test.body, got, test.valid)
			}
		})
	}
}

func TestNormalizeChatBody(t *testing.T) {
	if got, want := normalizeChatBody("  Xin\n\n chao\t moi nguoi  "), "Xin chao moi nguoi"; got != want {
		t.Fatalf("normalizeChatBody() = %q, want %q", got, want)
	}
}
