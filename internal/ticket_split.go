package internal

import "math"

// ticketSplit is the result of paying an amount with tickets of fixed denominations.
type ticketSplit struct {
	// Tickets is the number of tickets used.
	Tickets int
	// Covered is the part of the amount paid with tickets.
	Covered float64
	// Remainder is the part of the amount left to pay with another account.
	Remainder float64
}

// splitTickets finds the combination of ticket denominations covering as much of the amount as possible,
// preferring the one using the fewest tickets when several cover the same amount.
// Computations are done in the currency minor units (decimalPlaces) to avoid floating point drift.
func splitTickets(amount float64, denominations []float64, decimalPlaces int) ticketSplit {
	scale := math.Pow10(decimalPlaces)
	amountMinor := int(math.Round(amount * scale))

	// Work in steps of the denominations greatest common divisor to keep the table small.
	step := 0
	values := make([]int, 0, len(denominations))
	for _, d := range denominations {
		value := int(math.Round(d * scale))
		if value <= 0 {
			continue
		}
		values = append(values, value)
		step = gcd(step, value)
	}
	if step == 0 || amountMinor < step {
		return ticketSplit{Remainder: amount}
	}

	target := amountMinor / step
	// minTickets[s] is the fewest tickets summing exactly to s steps, or -1 when s is unreachable.
	minTickets := make([]int, target+1)
	for s := 1; s <= target; s++ {
		minTickets[s] = -1
		for _, value := range values {
			prev := s - value/step
			if prev < 0 || minTickets[prev] < 0 {
				continue
			}
			if minTickets[s] < 0 || minTickets[prev]+1 < minTickets[s] {
				minTickets[s] = minTickets[prev] + 1
			}
		}
	}

	covered := target
	for minTickets[covered] < 0 {
		covered--
	}
	coveredMinor := covered * step
	return ticketSplit{
		Tickets:   minTickets[covered],
		Covered:   float64(coveredMinor) / scale,
		Remainder: float64(amountMinor-coveredMinor) / scale,
	}
}

// gcd returns the greatest common divisor of a and b.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
