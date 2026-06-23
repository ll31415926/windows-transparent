package opacity

import "fmt"

const (
	Min = 20
	Max = 100
)

func Validate(value int) error {
	if value < Min || value > Max {
		return fmt.Errorf("opacity must be between %d and %d", Min, Max)
	}

	return nil
}

func ToAlpha(value int) byte {
	return byte((value*255 + 50) / 100)
}
