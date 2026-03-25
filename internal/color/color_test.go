package color

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRGBToXY(t *testing.T) {
	// Pure Red (255, 0, 0)
	xy := RGBToXY(255, 0, 0)
	// Calculated: X=0.649926, Y=0.234327, Z=0, sum=0.884253
	// x = 0.649926 / 0.884253 = 0.7350
	// y = 0.234327 / 0.884253 = 0.2650
	assert.InDelta(t, 0.735, xy.X, 0.001)
	assert.InDelta(t, 0.265, xy.Y, 0.001)

	// Pure Green (0, 255, 0)
	xy = RGBToXY(0, 255, 1) // Using 1 for blue to avoid zero sum if needed, but matrix has Z for green too.
	// Actually let's just use (0, 255, 0)
	xy = RGBToXY(0, 255, 0)
	// X = 0.103455, Y = 0.743075, Z = 0.053077, sum = 0.899607
	// x = 0.103455 / 0.899607 = 0.1150
	// y = 0.743075 / 0.899607 = 0.8260
	assert.InDelta(t, 0.115, xy.X, 0.001)
	assert.InDelta(t, 0.826, xy.Y, 0.001)

	// Pure Blue (0, 0, 255)
	xy = RGBToXY(0, 0, 255)
	// X = 0.197109, Y = 0.022598, Z = 1.035763, sum = 1.25547
	// x = 0.197109 / 1.25547 = 0.1570
	// y = 0.022598 / 1.25547 = 0.0180
	assert.InDelta(t, 0.157, xy.X, 0.001)
	assert.InDelta(t, 0.018, xy.Y, 0.001)
}

func TestKelvinToMired(t *testing.T) {
	assert.Equal(t, 400, KelvinToMired(2500))
	assert.Equal(t, 222, KelvinToMired(4500))
	assert.Equal(t, 153, KelvinToMired(6500))
	assert.Equal(t, 0, KelvinToMired(0))
}
