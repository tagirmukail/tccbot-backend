package trademath

import (
	"math"

	"github.com/tagirmukail/tccbot-backend/internal/utils"
)

type Singals struct {
	SMA float64  // sma with period
	WMA float64  // wma with period
	EMA float64  // ema with period
	BB  struct { // bolinger band
		TL float64 // top
		ML float64 // middle
		BL float64 // bottom
	}
}

type Calc struct{}

func (c *Calc) CalculateSignals(values []float64) Singals {
	tl, ml, bl := c.bolingerBandCalc(values)
	tl = utils.RoundFloat(tl, 4)
	ml = utils.RoundFloat(ml, 4)
	bl = utils.RoundFloat(bl, 4)
	return Singals{
		SMA: utils.RoundFloat(c.smaCalc(values), 4),
		WMA: utils.RoundFloat(c.wmaCalc(values), 4),
		EMA: utils.RoundFloat(c.emaCalc(values), 4),
		BB: struct {
			TL float64
			ML float64
			BL float64
		}{tl, ml, bl},
	}
}

func (c *Calc) bolingerBandCalc(values []float64) (tl, ml, bl float64) {
	var (
		sma       float64 // as middle line (ml)
		sum       float64
		stdDevNom float64
		d         float64 = 2 // is D coefficient
	)
	for i, value := range values {
		sum += value
		sma = sum / float64(i+1)
		stdDevNom += math.Pow(value-sma, 2)
	}
	// ML = SUM (CLOSE, N) / N = SMA20 (CLOSE, N)
	ml = sma
	// StdDev = SQRT ( SUM ( (CLOSE — SMA20 (CLOSE, N) )^2, N) /N)
	stdDev := math.Sqrt(stdDevNom / float64(len(values)))
	// TL = ML + (D * StdDev)
	tl = ml + (d * stdDev)
	// BL = ML - (D * StdDev)
	bl = ml - (d * stdDev)
	return tl, ml, bl
}

func (c *Calc) emaCalc(values []float64) float64 {
	var ema float64
	var smoothing float64 = 2
	var prevEMA float64
	for i, value := range values {
		if i == 0 {
			ema = value
			continue
		}
		period := float64(i) + 1
		multiplierN := smoothing / (1 + period)
		if i == 1 {
			ema = value*multiplierN + value*(1-multiplierN)
			prevEMA = ema
			continue
		}
		ema = value*multiplierN + prevEMA*(1-multiplierN)
		prevEMA = ema
	}
	return ema
}

func (c *Calc) smaCalc(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func (c *Calc) wmaCalc(values []float64) float64 {
	var (
		nom   float64
		denom float64
	)
	for i, value := range values {
		nom += float64(i+1) * value
		denom += float64(i + 1)
	}
	return nom / denom
}
