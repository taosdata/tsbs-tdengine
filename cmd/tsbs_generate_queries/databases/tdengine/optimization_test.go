package tdengine

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/timescale/tsbs/pkg/query"
)

func TestIoTFastHintScope(t *testing.T) {
	cases := []struct {
		name string
		fill func(*IoT, query.Query)
		fast bool
	}{
		{"last-loc", (*IoT).LastLocPerTruck, false},
		{"low-fuel", (*IoT).TrucksWithLowFuel, false},
		{"high-load", (*IoT).TrucksWithHighLoad, false},
		{"stationary-trucks", (*IoT).StationaryTrucks, false},
		{"long-driving-sessions", (*IoT).TrucksWithLongDrivingSessions, true},
		{"long-daily-sessions", (*IoT).TrucksWithLongDailySessions, true},
		{"avg-vs-projected-fuel-consumption", (*IoT).AvgVsProjectedFuelConsumption, false},
		{"avg-daily-driving-duration", (*IoT).AvgDailyDrivingDuration, true},
		{"avg-daily-driving-session", (*IoT).AvgDailyDrivingSession, true},
		{"avg-load", (*IoT).AvgLoad, false},
		{"daily-activity", (*IoT).DailyTruckActivity, true},
		{"breakdown-frequency", (*IoT).TruckBreakdownFrequency, true},
	}
	start := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			generate := func(enabled bool) string {
				rand.Seed(123)
				base := &BaseGenerator{UseFastInterval: enabled, UseIntervalLimit: enabled}
				generator, err := base.NewIoT(start, start.Add(4*24*time.Hour), 10)
				if err != nil {
					t.Fatal(err)
				}
				q := query.NewTDengine()
				defer q.Release()
				c.fill(generator.(*IoT), q)
				return string(q.SqlQuery)
			}
			baseline, hinted := generate(false), generate(true)
			count := strings.Count(hinted, "/*+ FAST_INTERVAL() */ ")
			if c.fast && count != 1 || !c.fast && count != 0 {
				t.Fatalf("unexpected FAST hint count %d: %s", count, hinted)
			}
			if strings.ReplaceAll(hinted, "/*+ FAST_INTERVAL() */ ", "") != baseline {
				t.Fatalf("query changed beyond its hint:\nbefore: %s\nafter: %s", baseline, hinted)
			}
			if c.fast {
				index := strings.Index(hinted, "/*+ FAST_INTERVAL() */")
				if strings.Contains(strings.ToLower(hinted[:index]), "interval(") {
					t.Fatalf("hint must precede the innermost raw-data interval: %s", hinted)
				}
				prefix := strings.TrimSpace(strings.ToLower(hinted[:index]))
				if !strings.HasSuffix(prefix, "select") || strings.EqualFold(prefix, "select") {
					t.Fatalf("hint must belong to the inner SELECT: %s", hinted)
				}
			}
			if strings.Contains(hinted, "INTERVAL_LIMIT") {
				t.Fatalf("unexpected interval-limit hint in IoT query: %s", hinted)
			}
		})
	}
}

func TestCPUOptimizationHints(t *testing.T) {
	cases := []struct {
		name  string
		fill  func(*Devops, query.Query)
		fast  bool
		limit bool
	}{
		{"single-groupby-one-host", func(d *Devops, q query.Query) { d.GroupByTime(q, 1, 1, time.Hour) }, true, false},
		{"single-groupby-eight-hosts", func(d *Devops, q query.Query) { d.GroupByTime(q, 8, 5, time.Hour) }, true, false},
		{"double-groupby-1", func(d *Devops, q query.Query) { d.GroupByTimeAndPrimaryTag(q, 1) }, true, false},
		{"double-groupby-all", func(d *Devops, q query.Query) { d.GroupByTimeAndPrimaryTag(q, 10) }, true, false},
		{"cpu-max-all-1", func(d *Devops, q query.Query) { d.MaxAllCPU(q, 1, 8*time.Hour) }, true, false},
		{"cpu-max-all-8", func(d *Devops, q query.Query) { d.MaxAllCPU(q, 8, 8*time.Hour) }, true, false},
		{"groupby-orderby-limit", (*Devops).GroupByOrderByLimit, false, true},
		{"lastpoint", (*Devops).LastPointPerHost, false, false},
		{"high-cpu", func(d *Devops, q query.Query) { d.HighCPUForHosts(q, 0) }, false, false},
	}
	start := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			generate := func(fast, limit bool) string {
				rand.Seed(123)
				base := &BaseGenerator{UseFastInterval: fast, UseIntervalLimit: limit}
				generator, err := base.NewDevops(start, start.Add(4*24*time.Hour), 10)
				if err != nil {
					t.Fatal(err)
				}
				q := query.NewTDengine()
				defer q.Release()
				c.fill(generator.(*Devops), q)
				return string(q.SqlQuery)
			}
			baseline := generate(false, false)
			for _, flags := range [][2]bool{{true, false}, {false, true}, {true, true}} {
				hinted := generate(flags[0], flags[1])
				if strings.Contains(hinted, "FAST_INTERVAL()") != (c.fast && flags[0]) ||
					strings.Contains(hinted, "INTERVAL_LIMIT()") != (c.limit && flags[1]) {
					t.Fatalf("incorrect hints for flags %v: %s", flags, hinted)
				}
				plain := strings.ReplaceAll(hinted, "/*+ FAST_INTERVAL() */ ", "")
				plain = strings.ReplaceAll(plain, "/*+ INTERVAL_LIMIT() */ ", "")
				if plain != baseline {
					t.Fatalf("query changed beyond its hint:\nbefore: %s\nafter: %s", baseline, hinted)
				}
			}
		})
	}
}
