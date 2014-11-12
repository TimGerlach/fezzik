package fezzik

import (
	"time"

	"github.com/GaryBoone/GoStats/stats"
	"github.com/pivotal-cf-experimental/veritas/say"
)

func PrintStatsReport(description string, s *stats.Stats) {
	say.Println(0, description)
	say.Println(1, say.Cyan("Count: %d - %.4f <= <%.4f> ± %.4f <= %.4f", s.Count(), s.Min(), s.Mean(), s.SampleStandardDeviation(), s.Max()))
}

func DurationMapStats(input map[string]time.Duration) *stats.Stats {
	s := &stats.Stats{}
	for _, d := range input {
		s.Update(d.Seconds())
	}

	return s
}
