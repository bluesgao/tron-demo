package main

import (
	"fmt"
	"github.com/shopspring/decimal"
)

func main() {
	realBetProportionalPercent := decimal.NewFromFloat(float64(100)).Sub(decimal.Zero).Div(decimal.NewFromInt(100))
	fmt.Printf("%s\n", realBetProportionalPercent.String())
}
