package types

import "fmt"

type Action uint8

const (
	NotTrend Action = iota
	UpTrend
	DownTrend
)

func (a *Action) Validate() error {
	switch *a {
	case NotTrend, UpTrend, DownTrend:
		return nil
	default:
		return fmt.Errorf("unknown action - %v", a)
	}
}
