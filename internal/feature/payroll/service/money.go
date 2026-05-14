package service

import (
	"fmt"
	"strconv"
	"strings"
)

const rateScale = int64(1_000_000)

type Money struct {
	cents int64
}

func ZeroMoney() Money {
	return Money{}
}

func ParseMoney(value string) (Money, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Money{}, fmt.Errorf("empty money value")
	}

	negative := false
	if strings.HasPrefix(value, "-") {
		negative = true
		value = strings.TrimPrefix(value, "-")
	}

	parts := strings.SplitN(value, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Money{}, fmt.Errorf("invalid money value")
	}

	var fraction int64
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 2 {
			frac = frac[:2]
		}
		for len(frac) < 2 {
			frac += "0"
		}
		fraction, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return Money{}, fmt.Errorf("invalid money value")
		}
	}

	cents := whole*100 + fraction
	if negative {
		cents = -cents
	}
	return Money{cents: cents}, nil
}

func MustMoney(value string) Money {
	money, err := ParseMoney(value)
	if err != nil {
		return ZeroMoney()
	}
	return money
}

func (m Money) Add(other Money) Money {
	return Money{cents: m.cents + other.cents}
}

func (m Money) Sub(other Money) Money {
	return Money{cents: m.cents - other.cents}
}

func (m Money) MulRate(rateMicros int64) Money {
	return Money{cents: (m.cents*rateMicros + rateScale/2) / rateScale}
}

func (m Money) DivInt(divisor int64) Money {
	if divisor == 0 {
		return ZeroMoney()
	}
	return Money{cents: m.cents / divisor}
}

func (m Money) MulInt(multiplier int64) Money {
	return Money{cents: m.cents * multiplier}
}

func (m Money) IsNegative() bool {
	return m.cents < 0
}

func (m Money) String() string {
	sign := ""
	cents := m.cents
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func ParseRateMicros(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty rate")
	}

	parts := strings.SplitN(value, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rate")
	}

	var fraction int64
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 6 {
			frac = frac[:6]
		}
		for len(frac) < 6 {
			frac += "0"
		}
		fraction, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid rate")
		}
	}
	return whole*rateScale + fraction, nil
}
