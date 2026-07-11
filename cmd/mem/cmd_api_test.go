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
