package timeext

import (
	"strings"
	"sync/atomic"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type Interval struct {
	// Lower inclusive bound.
	From time.Time
	// Upper inclusive bound.
	To time.Time
}

const MoscowTimeLocationName = "Europe/Moscow"

var moscowTimeLocationStorage atomic.Pointer[struct {
	Err error
	Loc *time.Location
}]

func MustMoscowTimeLocation() *time.Location {
	r, err := MoscowTimeLocation()
	if err != nil {
		panic(err)
	}

	return r
}

func MoscowTimeLocation() (*time.Location, error) {
	res := moscowTimeLocationStorage.Load()
	if res == nil {
		loc, err := time.LoadLocation(MoscowTimeLocationName)
		if err != nil {
			res = &struct {
				Err error
				Loc *time.Location
			}{
				Err: err,
			}
		} else {
			res = &struct {
				Err error
				Loc *time.Location
			}{
				Loc: loc,
			}
		}

		moscowTimeLocationStorage.Store(res)
	}

	return res.Loc, res.Err
}

func MustMoscowNow() time.Time {
	return time.Now().In(MustMoscowTimeLocation())
}

func FormatRFC3339NoTZInMoscowLocation(t time.Time) string {
	const format = "2006-01-02T15:04:05"

	return t.In(MustMoscowTimeLocation()).Format(format)
}

func FormatRFC3339DateInMoscowLocation(t time.Time) string {
	const format = "2006-01-02"

	return t.In(MustMoscowTimeLocation()).Format(format)
}

func ParseRFC3339AsMoscowLocation(v string) (time.Time, error) {
	dateTimeTFormat := strings.Replace(v, " ", "T", 1)

	parsedTime, err := time.Parse(time.RFC3339, dateTimeTFormat+"+03:00")
	if err != nil {
		return time.Time{}, errs.BadArgsWrapWith(err)
	}

	return parsedTime, nil
}

func ParseRFC3339DateAsMoscowLocation(v string) (time.Time, error) {
	parsedTime, err := time.Parse(time.RFC3339, v+"T00:00:00+03:00")
	if err != nil {
		return time.Time{}, errs.BadArgsWrapWith(err)
	}

	return parsedTime, nil
}

func BeginOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func BeginOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Microsecond*999999), t.Location()) //nolint:mnd,revive
}

func BeginOfPreviousDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()-1, 0, 0, 0, 0, t.Location())
}

func BeginOfNextDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
}

func BeginOfNextMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
}

func ChangeLocation(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

func SplitByDay(from, to time.Time, loc *time.Location) []Interval {
	from = from.In(loc)
	to = to.In(loc)

	if from.Compare(to) > 0 {
		return nil
	}

	res := make([]Interval, 0, int(to.Sub(from).Hours()/24.0)+1) //nolint:mnd,revive

	for {
		next := time.Date(from.Year(), from.Month(), from.Day(), 23, 59, 59, int(time.Microsecond*999999), from.Location()) //nolint:mnd,revive
		if next.Compare(to) >= 0 {
			res = append(res, Interval{From: from, To: next})

			return res
		}

		res = append(res, Interval{From: from, To: next})
		from = time.Date(from.Year(), from.Month(), from.Day()+1, 0, 0, 0, 0, from.Location()) //nolint:mnd,revive
	}
}

func SplitByMonth(from, to time.Time, loc *time.Location) []Interval {
	from = from.In(loc)
	to = to.In(loc)

	if from.Compare(to) > 0 {
		return nil
	}

	res := make([]Interval, 0, int(to.Sub(from).Hours()/720.)+1) //nolint:mnd,revive

	for {
		next := time.Date(from.Year(), from.Month()+1, 0, 23, 59, 59, int(time.Microsecond*999999), from.Location()) //nolint:mnd,revive
		if next.Compare(to) >= 0 {
			res = append(res, Interval{From: from, To: EndOfDay(to)})

			return res
		}

		res = append(res, Interval{From: from, To: next})
		from = time.Date(from.Year(), from.Month()+1, 1, 0, 0, 0, 0, from.Location()) //nolint:mnd,revive
	}
}

func IsLastDayOfMonth(t time.Time) bool {
	if tomorrow := t.AddDate(0, 0, 1); t.Month() != tomorrow.Month() {
		return true
	}

	return false
}

// SecondPrecision indicates whether specified moment precised up to second.
func SecondPrecision(v time.Time) bool {
	return v.Equal(time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), 0, v.Location()))
}
