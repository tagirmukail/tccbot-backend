package trademath

import (
	"math"
)

type MAIndication uint16

const (
	SMAIndication MAIndication = iota
	EMAIndication
	WMAIndication
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

type MACD struct {
	HistogramValue float64 // histogram value
	Value          float64 // macd value
	Sig            float64 // MACD signal curve
}

type RSI struct {
	Value float64
}

type Calc struct{}

// CalculateMACD - calculate macd, indication - use only EMA or WMA.
// recommendation: count for values: fast=12, slow=26, prevMACDValues=8
func (c *Calc) CalculateMACD(
	fastValues []float64, slowValues []float64, beforeMACDValues []float64, indication MAIndication,
) MACD {
	var (
		macd           MACD
		prevMACDValues = append([]float64{}, beforeMACDValues...)
		maCalcFunc     func(values []float64) float64
	)
	if len(fastValues) >= len(slowValues) {
		return macd
	}

	switch indication {
	case EMAIndication:
		maCalcFunc = c.EMACalc
	case WMAIndication:
		maCalcFunc = c.WMACalc
	default:
		return macd
	}

	fastMA := maCalcFunc(fastValues)
	slowMA := maCalcFunc(slowValues)

	macd.Value = RoundFloat(fastMA-slowMA, 3)

	prevMACDValues = append(prevMACDValues, macd.Value)
	macd.Sig = c.EMACalc(prevMACDValues)

	macd.HistogramValue = RoundFloat(macd.Value-macd.Sig, 3)
	return macd
}

// CalculateRSI - calculate rsi by values and ma
// recommendation: count values = 14, and indication - WMA
func (c *Calc) CalculateRSI(values []float64, indication MAIndication) RSI {
	var (
		rsi        RSI
		gainValues []float64
		lossValues []float64
		maCalcFunc func(values []float64) float64
	)
	if len(values) < 2 {
		return rsi
	}
	switch indication {
	case SMAIndication:
		maCalcFunc = c.SMACalc
	case EMAIndication:
		maCalcFunc = c.EMACalc
	case WMAIndication:
		maCalcFunc = c.WMACalc
	default:
		return rsi
	}

	for i, value := range values {
		if i == 0 {
			continue
		}
		change := value - values[i-1]
		if change > 0 {
			gainValues = append(gainValues, change)
		} else if change < 0 {
			lossValues = append(lossValues, math.Abs(change))
		}
	}

	gain := maCalcFunc(gainValues)
	loss := maCalcFunc(lossValues)
	var rs float64
	if loss != 0 {
		rs = gain / loss
	} else {
		rs = 100
	}
	rsiVal := 100 - (100 / (1 + rs))
	rsi.Value = RoundFloat(rsiVal, 3)
	return rsi
}

func (c *Calc) CalculateSignals(values []float64) Singals {
	tl, ml, bl := c.BolingerBandCalc(values)
	return Singals{
		SMA: c.SMACalc(values),
		WMA: c.WMACalc(values),
		EMA: c.EMACalc(values),
		BB: struct {
			TL float64
			ML float64
			BL float64
		}{tl, ml, bl},
	}
}

func (c *Calc) BolingerBandCalc(values []float64) (tl, ml, bl float64) {
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
	tl = RoundFloat(tl, 4)
	ml = RoundFloat(ml, 4)
	bl = RoundFloat(bl, 4)
	return tl, ml, bl
}

func (c *Calc) EMACalc(values []float64) float64 {
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
	return RoundFloat(ema, 4)
}

func (c *Calc) SMACalc(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return RoundFloat(sum/float64(len(values)), 4)
}

func (c *Calc) WMACalc(values []float64) float64 {
	var (
		nom   float64
		denom float64
	)
	for i, value := range values {
		nom += float64(i+1) * value
		denom += float64(i + 1)
	}
	return RoundFloat(nom/denom, 4)
}

// RoundFloat rounds your floating point number to the desired decimal place
func RoundFloat(x float64, prec int) float64 {
	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)
	intermed += .5
	x = .5
	if frac < 0.0 {
		x = -.5
		intermed--
	}
	if frac >= x {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow
}

const SatoshisPerBTC = 100000000

func RemoveEmptyValues(indexes []float64, vals []float64) ([]float64, []float64) {
	var indxesResult = make([]float64, 0)
	var result = make([]float64, 0)
	for i, indx := range indexes {
		if vals[i] == 0 {
			continue
		}
		indxesResult = append(indxesResult, indx)
		result = append(result, vals[i])
	}

	return indxesResult, result
}

func ConvertToBTC(balance int64) float64 {
	return float64(balance) / SatoshisPerBTC
}

func ConvertFromBTCToContracts(balance float64) float64 {
	return balance * SatoshisPerBTC
}
