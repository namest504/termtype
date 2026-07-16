package store

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// minPBDurS is the shortest round (seconds) eligible for a personal best;
// shorter rounds are too noisy to be a fair pace.
const minPBDurS = 5.0

// Key identifies a comparable bucket of rounds. Bests and recents are always
// computed within one bucket so different modes/languages never compete.
type Key struct {
	Mode   string
	Lang   string
	Source string
}

// KeyOf returns the bucket a round belongs to.
func KeyOf(r Round) Key { return Key{Mode: r.Mode, Lang: r.Lang, Source: r.Source} }

// PBEligible reports whether a round may set a personal best.
func PBEligible(r Round) bool { return r.DurS >= minPBDurS }

// ModeString converts a time-attack limit to its history/config value.
func ModeString(limit time.Duration) string {
	if limit <= 0 {
		return "normal"
	}
	return fmt.Sprintf("ta%d", int(limit.Seconds()))
}

// Best returns the highest PB-eligible WPM in the bucket, and whether one exists.
func Best(rounds []Round, k Key) (float64, bool) {
	best, ok := 0.0, false
	for _, r := range rounds {
		if KeyOf(r) != k || !PBEligible(r) {
			continue
		}
		if !ok || r.WPM > best {
			best, ok = r.WPM, true
		}
	}
	return best, ok
}

// RecentWPMs returns the WPMs of the last n rounds in the bucket, oldest first.
func RecentWPMs(rounds []Round, k Key, n int) []float64 {
	var wpms []float64
	for _, r := range rounds {
		if KeyOf(r) == k {
			wpms = append(wpms, r.WPM)
		}
	}
	if len(wpms) > n {
		wpms = wpms[len(wpms)-n:]
	}
	return wpms
}

// FormatStats renders the --stats summary: one line per bucket with run
// count, best and average WPM, and average accuracy.
func FormatStats(rounds []Round) string {
	if len(rounds) == 0 {
		return "No history yet — play a round first!\n"
	}
	type agg struct {
		n      int
		wpmSum float64
		accSum float64
	}
	aggs := map[Key]*agg{}
	for _, r := range rounds {
		k := KeyOf(r)
		if aggs[k] == nil {
			aggs[k] = &agg{}
		}
		aggs[k].n++
		aggs[k].wpmSum += r.WPM
		aggs[k].accSum += r.Acc
	}
	keys := make([]Key, 0, len(aggs))
	for k := range aggs {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Mode != keys[j].Mode {
			return keys[i].Mode < keys[j].Mode
		}
		if keys[i].Lang != keys[j].Lang {
			return keys[i].Lang < keys[j].Lang
		}
		return keys[i].Source < keys[j].Source
	})
	var b strings.Builder
	for _, k := range keys {
		a := aggs[k]
		bestStr := "    -"
		if best, ok := Best(rounds, k); ok {
			bestStr = fmt.Sprintf("%5.1f", best)
		}
		src := ""
		if k.Source != "builtin" {
			src = " (" + k.Source + ")"
		}
		fmt.Fprintf(&b, "%-7s %-3s%s runs %-4d best %s  avg %5.1f wpm  acc %.1f%%\n",
			k.Mode, k.Lang, src, a.n, bestStr, a.wpmSum/float64(a.n), a.accSum/float64(a.n))
	}
	return b.String()
}
