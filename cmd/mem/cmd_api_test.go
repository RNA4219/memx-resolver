package main

import "testing"

func TestIsLoopbackAddress(t *testing.T) {
	tests := []struct {
		address string
		want    bool
	}{
		{"127.0.0.1:8080", true},
		{"[::1]:8080", true},
		{"localhost:8080", true},
		{"0.0.0.0:8080", false},
		{"192.0.2.1:8080", false},
		{":8080", false},
		{"invalid", false},
	}
	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			if got := isLoopbackAddress(tt.address); got != tt.want {
				t.Fatalf("isLoopbackAddress(%q) = %v, want %v", tt.address, got, tt.want)
			}
		})
	}
}

func TestValidateBindAddressV2(t *testing.T) {
	tests := []struct {
		name  string
		addr  string
		allow bool
		want  bool
	}{
		{"loopback", "127.0.0.1:7766", false, false},
		{"wildcard-rejected", "0.0.0.0:7766", false, true},
		{"hostname-rejected", "example.com:7766", false, true},
		{"wildcard-allowed", "0.0.0.0:7766", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBindAddress(tt.addr, tt.allow)
			if (err != nil) != tt.want {
				t.Fatalf("validateBindAddress(%q, %v) error=%v, wantError=%v", tt.addr, tt.allow, err, tt.want)
			}
		})
	}
}
