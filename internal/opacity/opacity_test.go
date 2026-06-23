package opacity

import "testing"

func TestToAlpha(t *testing.T) {
	tests := []struct {
		name    string
		opacity int
		alpha   byte
	}{
		{name: "fully opaque", opacity: 100, alpha: 255},
		{name: "half", opacity: 50, alpha: 128},
		{name: "minimum", opacity: 20, alpha: 51},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToAlpha(tt.opacity); got != tt.alpha {
				t.Fatalf("ToAlpha(%d) = %d, want %d", tt.opacity, got, tt.alpha)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		opacity int
		wantErr bool
	}{
		{name: "minimum", opacity: 20},
		{name: "maximum", opacity: 100},
		{name: "too low", opacity: 19, wantErr: true},
		{name: "too high", opacity: 101, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.opacity)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate(%d) error = %v, wantErr %v", tt.opacity, err, tt.wantErr)
			}
		})
	}
}
