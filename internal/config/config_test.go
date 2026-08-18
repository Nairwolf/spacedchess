package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		name             string
		databaseURL      string
		port             string
		cookieSecure     string
		wantAddr         string
		wantCookieSecure bool
		wantErr          bool
	}{
		{
			name:    "missing DATABASE_URL",
			port:    "8080",
			wantErr: true,
		},
		{
			name:        "default port",
			databaseURL: "postgres://localhost/db",
			wantAddr:    ":8080",
		},
		{
			name:        "port override",
			databaseURL: "postgres://localhost/db",
			port:        "9000",
			wantAddr:    ":9000",
		},
		{
			name:             "cookie secure",
			databaseURL:      "postgres://localhost/db",
			cookieSecure:     "true",
			wantAddr:         ":8080",
			wantCookieSecure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", tt.databaseURL)
			t.Setenv("PORT", tt.port)
			t.Setenv("COOKIE_SECURE", tt.cookieSecure)

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Addr != tt.wantAddr {
				t.Errorf("Addr = %q, want %q", cfg.Addr, tt.wantAddr)
			}
			if cfg.DatabaseURL != tt.databaseURL {
				t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, tt.databaseURL)
			}
			if cfg.CookieSecure != tt.wantCookieSecure {
				t.Errorf("CookieSecure = %v, want %v", cfg.CookieSecure, tt.wantCookieSecure)
			}
		})
	}
}
