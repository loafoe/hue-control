package color

import "math"

// XY represents coordinates in the CIE 1931 color space.
type XY struct {
	X float64
	Y float64
}

// RGBToXY converts RGB color values (0-255) to XY color space used by Philips Hue.
func RGBToXY(r, g, b int) XY {
	// Normalize RGB values to 0-1 range
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	// Apply gamma correction
	if rf > 0.04045 {
		rf = math.Pow((rf+0.055)/1.055, 2.4)
	} else {
		rf = rf / 12.92
	}
	if gf > 0.04045 {
		gf = math.Pow((gf+0.055)/1.055, 2.4)
	} else {
		gf = gf / 12.92
	}
	if bf > 0.04045 {
		bf = math.Pow((bf+0.055)/1.055, 2.4)
	} else {
		bf = bf / 12.92
	}

	// Convert to XYZ using the Wide RGB D65 conversion matrix
	X := rf*0.649926 + gf*0.103455 + bf*0.197109
	Y := rf*0.234327 + gf*0.743075 + bf*0.022598
	Z := rf*0.000000 + gf*0.053077 + bf*1.035763

	// Calculate xy values from XYZ
	sum := X + Y + Z
	if sum == 0 {
		return XY{X: 0, Y: 0}
	}

	return XY{
		X: X / sum,
		Y: Y / sum,
	}
}

// KelvinToMired converts color temperature in Kelvin to mired (micro reciprocal degree).
func KelvinToMired(kelvin int) int {
	if kelvin == 0 {
		return 0
	}
	return 1000000 / kelvin
}
