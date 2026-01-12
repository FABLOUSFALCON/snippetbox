package main

import (
	"testing"
	"time"

	"github.com/FABLOUSFALCON/snippetbox/internal/assert"
)

func TestHumanDate(t *testing.T) {
	testCases := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2024, 3, 17, 10, 15, 0, 0, time.UTC),
			want: "17 Mar 2024 at 10:15",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2024, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17 Mar 2024 at 09:15",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			hd := humanDate(tC.tm)

			assert.Equal(t, hd, tC.want)
		})
	}
}
