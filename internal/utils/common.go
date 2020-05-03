package utils

import (
	"fmt"
	"math"
	"time"
)

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

func ConvertToBTC(balance int64) float64 {
	return float64(balance) / SatoshisPerBTC
}

func FromTime(now time.Time, binSize string, count int) (from time.Time, err error) {
	var (
		toTime   time.Time
		duration time.Duration
	)
	switch binSize {
	case "1m":
		duration = time.Duration(count) * time.Minute
		toTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
	case "5m":
		duration = time.Duration(count) * (5 * time.Minute)
		toTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
	case "1h":
		duration = time.Duration(count) * (1 * time.Hour)
		toTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
	case "1d":
		duration = time.Duration(count) * (24 * time.Hour)
		toTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	default:
		err = fmt.Errorf("unsupported bin size:%s", binSize)
		return
	}
	from = toTime.Add(-duration)
	return
}
