package trademath

import (
	"math"

	"github.com/markcheno/go-talib"
)

type MAIndication uint16

const (
	SMAIndication MAIndication = iota
	EMAIndication
	WMAIndication
)

type Signals struct {
	SMA float64  // sma with period
	WMA float64  // wma with period
	EMA float64  // ema with period
	BB  struct { // bolinger band
		TL float64 // top
		ML float64 // middle
		BL float64 // bottom
	}
}

type BB struct { // bolinger band
	TL float64 // top
	ML float64 // middle
	BL float64 // bottom
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

func (c *Calc) CalcMACD(
	values []float64,
	inFastPeriod int, inFastMAType talib.MaType,
	inSlowPeriod int, inSlowMAType talib.MaType,
	inSignalPeriod int, inSignalMAType talib.MaType,
) MACD {
	macd, macdSig, macdHist := talib.MacdExt(values, inFastPeriod, inFastMAType, inSlowPeriod, inSlowMAType,
		inSignalPeriod, inSignalMAType)

	return MACD{
		HistogramValue: RoundFloat(macdHist[len(macdHist)-1], 3),
		Value:          RoundFloat(macd[len(macd)-1], 3),
		Sig:            RoundFloat(macdSig[len(macdSig)-1], 3),
	}
}

// CalculateMACD - calculate macd, indication - use only EMA or WMA.
// recommendation: count for values: fast=12, slow=26, prevMACDValues=8
// Deprecated
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

func (c *Calc) CalcRSI(values []float64, intTimePeriod int) RSI {
	rsi := talib.Rsi(values, intTimePeriod)
	return RSI{Value: RoundFloat(rsi[len(rsi)-1], 4)}
}

// CalculateRSI - calculate rsi by values and ma
// recommendation: count values = 14, and indication - WMA
// Deprecated
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

// Deprecated
func (c *Calc) CalculateSignals(values []float64) Signals {
	tl, ml, bl := c.BolingerBandCalc(values)
	return Signals{
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

func (c *Calc) CalcBB(values []float64, maType talib.MaType) (tlV, mlV, blV float64) {
	tl, ml, bl := talib.BBands(values, len(values), 2, 2, maType)
	if len(tl) == 0 || len(ml) == 0 || len(bl) == 0 {
		return 0, 0, 0
	}
	return RoundFloat(tl[len(tl)-1], 4), RoundFloat(ml[len(ml)-1], 4), RoundFloat(bl[len(bl)-1], 4)
}

func (c *Calc) CalcSignals(values []float64, maType talib.MaType) Signals {
	tl, ml, bl := talib.BBands(values, len(values), 2, 2, maType)
	sma := talib.Sma(values, len(values))
	ema := talib.Ema(values, len(values))
	wma := talib.Wma(values, len(values))
	return Signals{
		SMA: RoundFloat(sma[len(sma)-1], 4),
		WMA: RoundFloat(wma[len(wma)-1], 4),
		EMA: RoundFloat(ema[len(ema)-1], 4),
		BB: struct {
			TL float64
			ML float64
			BL float64
		}{RoundFloat(tl[len(tl)-1], 4), RoundFloat(ml[len(ml)-1], 4), RoundFloat(bl[len(bl)-1], 4)},
	}
}

// Deprecated
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

// Deprecated
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

// Deprecated
func (c *Calc) SMACalc(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return RoundFloat(sum/float64(len(values)), 4)
}

// Deprecated
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

func ConvertToBTC(b int64) float64 {
	return float64(b) / SatoshisPerBTC
}

func ConvertFromBTCToContracts(balance float64) float64 {
	return balance * SatoshisPerBTC
}

func CalculateUnrealizedPNL(openPrice, lastPrice float64, contractsCount int64) float64 {
	result := (1/openPrice - 1/lastPrice) * float64(contractsCount)
	return RoundFloat(result, 8)
}
