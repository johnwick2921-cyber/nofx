package kernel

import "time"

// IsCMEOpen reports whether CME Globex is open for index futures at the given time.
// Globex hours (Chicago time):
//
//	Sunday 17:00 → Friday 16:00, with a 60-minute daily break at 16:00–17:00.
//
// Holidays observed: New Year, MLK Day, Presidents Day, Good Friday, Memorial Day,
// Juneteenth, Independence Day, Labor Day, Thanksgiving (+ day after), Christmas Eve,
// Christmas Day, New Year's Eve. Each may have shortened hours; for v1 we treat them
// as full closures and refuse to trade. Refine in Plan 3 if it becomes restrictive.
func IsCMEOpen(t time.Time) bool {
	chicago, _ := time.LoadLocation("America/Chicago")
	ct := t.In(chicago)
	if isCMEHoliday(ct) {
		return false
	}
	wd := ct.Weekday()
	hour := ct.Hour()
	switch wd {
	case time.Saturday:
		return false
	case time.Sunday:
		return hour >= 17
	case time.Friday:
		return hour < 16
	default: // Mon-Thu
		return hour != 16
	}
}

// isCMEHoliday returns true if t falls on a CME-observed full-closure holiday.
// CME may have shortened-hours days (e.g. Good Friday, day after Thanksgiving),
// but for v1 we treat shortened days as full closures and refuse to trade.
// Refine in a later plan if this becomes operationally restrictive.
func isCMEHoliday(ct time.Time) bool {
	year := ct.Year()
	month := ct.Month()
	day := ct.Day()
	weekday := ct.Weekday()

	// Fixed-date holidays
	md := ct.Format("01-02")
	switch md {
	case "01-01": // New Year's Day
		return true
	case "06-19": // Juneteenth
		return true
	case "07-04": // Independence Day
		return true
	case "12-24": // Christmas Eve (early close treated as closure)
		return true
	case "12-25": // Christmas Day
		return true
	case "12-31": // New Year's Eve (early close treated as closure)
		return true
	}

	// MLK Day — 3rd Monday of January
	if month == time.January && weekday == time.Monday && (day-1)/7 == 2 {
		return true
	}

	// Presidents Day — 3rd Monday of February
	if month == time.February && weekday == time.Monday && (day-1)/7 == 2 {
		return true
	}

	// Good Friday — Friday before Easter
	if month == time.March || month == time.April {
		easter := easterSunday(year)
		goodFri := easter.AddDate(0, 0, -2)
		if ct.Year() == goodFri.Year() && ct.Month() == goodFri.Month() && ct.Day() == goodFri.Day() {
			return true
		}
	}

	// Memorial Day — last Monday of May
	if month == time.May && weekday == time.Monday {
		// Check if next Monday is in June (i.e. this is the last Monday of May)
		nextMon := ct.AddDate(0, 0, 7)
		if nextMon.Month() == time.June {
			return true
		}
	}

	// Labor Day — 1st Monday of September
	if month == time.September && weekday == time.Monday && day <= 7 {
		return true
	}

	// Thanksgiving — 4th Thursday of November (plus day after as early-close)
	if month == time.November && weekday == time.Thursday && (day-1)/7 == 3 {
		return true
	}
	// Day after Thanksgiving — Friday after 4th Thursday
	if month == time.November && weekday == time.Friday {
		thursday := ct.AddDate(0, 0, -1)
		if thursday.Month() == time.November && (thursday.Day()-1)/7 == 3 {
			return true
		}
	}

	return false
}

// easterSunday returns the date of Easter Sunday in the given year (Western/Gregorian).
// Used only for Good Friday calculation.
func easterSunday(year int) time.Time {
	// Anonymous Gregorian algorithm (Meeus/Jones/Butcher)
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
