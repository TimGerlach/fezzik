package fezzik

import (
	"fmt"
	"time"

	"github.com/GaryBoone/GoStats/stats"
)

func PrintStatsReport(description string, s *stats.Stats) {
	fmt.Println(description)
	fmt.Printf("  Count: %d, Min: %.4f, Max: %.4f, Mean: %.4f, StdDev: %.4f\n", s.Count(), s.Min(), s.Max(), s.Mean(), s.SampleStandardDeviation())
}

func DurationMapStats(input map[string]time.Duration) *stats.Stats {
	s := &stats.Stats{}
	for _, d := range input {
		s.Update(d.Seconds())
	}

	return s
}
