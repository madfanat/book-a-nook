package config

import "testing"

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name, database, address, first, second string
		valid                                  bool
	}{
		{"defaults", "postgres://localhost/bookanook", "", "first", "second", true},
		{"container address", "postgres://db/bookanook", ":8080", "first", "second", true},
		{"missing database", "", "", "first", "second", false},
		{"invalid address", "postgres://db/bookanook", "localhost", "first", "second", false},
		{"invalid port", "postgres://db/bookanook", ":99999", "first", "second", false},
		{"missing token", "postgres://db/bookanook", "", "", "second", false},
		{"same tokens", "postgres://db/bookanook", "", "same", "same", false},
		{"token whitespace", "postgres://db/bookanook", "", "first token", "second", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", tc.database)
			t.Setenv("HTTP_ADDR", tc.address)
			t.Setenv("DEV_TOKEN_TIM", tc.first)
			t.Setenv("DEV_TOKEN_HELEN", tc.second)
			cfg, err := Load()
			if (err == nil) != tc.valid {
				t.Fatalf("Load() error = %v", err)
			}
			if tc.valid && (cfg.Tokens[tc.first] != "tim" || cfg.Tokens[tc.second] != "helen") {
				t.Fatal("incorrect token identities")
			}
			if tc.valid && tc.address == "" && cfg.HTTPAddr != "127.0.0.1:8080" {
				t.Fatal("incorrect default address")
			}
		})
	}
}
