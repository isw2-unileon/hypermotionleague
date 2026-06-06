// Package market computes the trading-window state of a league market: whether
// it is open, why it is closed, and when it next changes.
//
// Rule (Sprint 3): the market is open when BOTH hold —
//
//  1. the instant falls in the daily window [19:00, 00:00) in Europe/Madrid, and
//  2. no matchday of the league is in play (start_date <= now < end_date).
//
// All time logic is pure: callers pass `now` and the *time.Location, so it is
// fully testable without a clock or a database.
package market

import (
	"sort"
	"time"

	// Embed the IANA tz database in the binary so LoadLocation("Europe/Madrid")
	// works on minimal runtimes (Alpine) that ship no system tzdata.
	_ "time/tzdata"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// TZName is the league's reference timezone. The window is defined in this zone,
// so DST shifts (CET/CEST) are handled correctly by using the named zone rather
// than a fixed offset.
const TZName = "Europe/Madrid"

// windowOpenHour is the local start hour of the daily trading window. The window
// runs [19:00, 00:00) — i.e. it closes at midnight.
const windowOpenHour = 19

// scanHorizonDays bounds how far ahead nextChangeAt looks for a state flip. The
// daily window guarantees a change within ~24h unless a matchday spans longer;
// the horizon is generous enough to cover multi-day matchdays.
const scanHorizonDays = 30

// ClosedReason explains the market state. ReasonOpen is used when open; the
// other two say why it is closed.
type ClosedReason string

const (
	ReasonOpen           ClosedReason = "open"
	ReasonOutsideWindow  ClosedReason = "outside_window"
	ReasonActiveMatchday ClosedReason = "active_matchday"
)

// Window is the computed market state at a given instant.
type Window struct {
	IsOpen       bool
	NextChangeAt time.Time // absolute instant of the next open<->closed flip
	Reason       ClosedReason
}

// Location loads the reference timezone (Europe/Madrid). With the embedded tz
// database it does not depend on the host OS; an error is returned only if that
// database is somehow unavailable.
func Location() (*time.Location, error) {
	return time.LoadLocation(TZName)
}

// ComputeWindow returns the market state at `now` for a league with the given
// matchdays. A nil/empty slice means "no active matchday", so the result then
// depends only on the time window — the function never errors on missing data.
//
// When the market is closed by both conditions at once (outside the window AND a
// matchday in play), Reason is ReasonActiveMatchday: it is the more informative
// cause and the one that keeps the market shut even inside the time window.
func ComputeWindow(now time.Time, loc *time.Location, matchdays []models.Matchday) Window {
	open := isOpenAt(now, loc, matchdays)

	w := Window{
		IsOpen:       open,
		NextChangeAt: nextChangeAt(now, loc, matchdays, open),
	}
	switch {
	case open:
		w.Reason = ReasonOpen
	case hasActiveMatchday(now, matchdays):
		w.Reason = ReasonActiveMatchday
	default:
		w.Reason = ReasonOutsideWindow
	}
	return w
}

// isOpenAt evaluates the open rule at instant t.
func isOpenAt(t time.Time, loc *time.Location, matchdays []models.Matchday) bool {
	return withinWindow(t, loc) && !hasActiveMatchday(t, matchdays)
}

// withinWindow reports whether t falls in the daily [19:00, 00:00) window in loc.
// Evaluated on local wall-clock hour, so it follows CET/CEST automatically.
func withinWindow(t time.Time, loc *time.Location) bool {
	return t.In(loc).Hour() >= windowOpenHour
}

// hasActiveMatchday reports whether any matchday is in play at t
// (start_date <= t < end_date). These are absolute instants, so tz is irrelevant.
func hasActiveMatchday(t time.Time, matchdays []models.Matchday) bool {
	for _, m := range matchdays {
		if !t.Before(m.StartDate) && t.Before(m.EndDate) {
			return true
		}
	}
	return false
}

// nextChangeAt returns the next instant at which is_open flips. It collects the
// ordered candidate boundaries after now — the daily window edges plus every
// future matchday start/end — and returns the first whose open-state differs
// from openNow. Evaluating the same isOpenAt at each boundary makes midnight,
// matchdays that start mid-window, overlaps, and DST transitions all fall out
// without special cases.
func nextChangeAt(now time.Time, loc *time.Location, matchdays []models.Matchday, openNow bool) time.Time {
	cands := windowBoundaries(now, loc, scanHorizonDays)
	for _, m := range matchdays {
		if m.StartDate.After(now) {
			cands = append(cands, m.StartDate)
		}
		if m.EndDate.After(now) {
			cands = append(cands, m.EndDate)
		}
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].Before(cands[j]) })

	for _, t := range cands {
		if isOpenAt(t, loc, matchdays) != openNow {
			return t
		}
	}
	// No flip within the horizon (e.g. a matchday longer than scanHorizonDays):
	// return the furthest candidate so callers still get a valid absolute
	// timestamp rather than a zero value.
	if len(cands) > 0 {
		return cands[len(cands)-1]
	}
	return now
}

// windowBoundaries returns the daily window edges — 19:00 (open) and the
// following 00:00 (close) — for the next `days` days in loc, keeping only
// instants strictly after now. Built with time.Date in the named location so
// DST transitions produce the correct absolute instants.
func windowBoundaries(now time.Time, loc *time.Location, days int) []time.Time {
	y, mo, d := now.In(loc).Date()
	out := make([]time.Time, 0, (days+1)*2)
	for i := 0; i <= days; i++ {
		openEdge := time.Date(y, mo, d+i, windowOpenHour, 0, 0, 0, loc)
		closeEdge := time.Date(y, mo, d+i+1, 0, 0, 0, 0, loc) // midnight ending day d+i
		if openEdge.After(now) {
			out = append(out, openEdge)
		}
		if closeEdge.After(now) {
			out = append(out, closeEdge)
		}
	}
	return out
}
