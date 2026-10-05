package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitTickets(t *testing.T) {
	tests := []struct {
		name          string
		amount        float64
		denominations []float64
		want          ticketSplit
	}{
		{
			name:          "uses the denomination leaving the smallest remainder",
			amount:        14,
			denominations: []float64{8, 10},
			want:          ticketSplit{Tickets: 1, Covered: 10, Remainder: 4},
		},
		{
			name:          "covers the whole amount when a combination matches it",
			amount:        16,
			denominations: []float64{8, 10},
			want:          ticketSplit{Tickets: 2, Covered: 16, Remainder: 0},
		},
		{
			name:          "mixes denominations and prefers the fewest tickets on ties",
			amount:        66.33,
			denominations: []float64{8, 10},
			want:          ticketSplit{Tickets: 7, Covered: 66, Remainder: 0.33},
		},
		{
			name:          "prefers larger denominations when the covered amount is the same",
			amount:        80,
			denominations: []float64{8, 10},
			want:          ticketSplit{Tickets: 8, Covered: 80, Remainder: 0},
		},
		{
			name:          "keeps the single denomination behaviour",
			amount:        26,
			denominations: []float64{8},
			want:          ticketSplit{Tickets: 3, Covered: 24, Remainder: 2},
		},
		{
			name:          "does not depend on the denominations order",
			amount:        9,
			denominations: []float64{10, 8},
			want:          ticketSplit{Tickets: 1, Covered: 8, Remainder: 1},
		},
		{
			name:          "handles decimal denominations without float drift",
			amount:        15.30,
			denominations: []float64{5.10},
			want:          ticketSplit{Tickets: 3, Covered: 15.30, Remainder: 0},
		},
		{
			name:          "uses no ticket when the amount is below every denomination",
			amount:        7.99,
			denominations: []float64{8, 10},
			want:          ticketSplit{Tickets: 0, Covered: 0, Remainder: 7.99},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := splitTickets(tt.amount, tt.denominations, 2)

			// Assert
			assert.Equal(t, tt.want.Tickets, got.Tickets)
			assert.InDelta(t, tt.want.Covered, got.Covered, 1e-9)
			assert.InDelta(t, tt.want.Remainder, got.Remainder, 1e-9)
		})
	}
}
