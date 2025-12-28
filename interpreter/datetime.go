package interpreter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alexisbouchez/phpgo/runtime"
)

// DateTimeObject represents a PHP DateTime object
type DateTimeObject struct {
	Time     time.Time
	Timezone *time.Location
}

func NewDateTime(t time.Time, tz *time.Location) *DateTimeObject {
	if tz == nil {
		tz = time.Local
	}
	return &DateTimeObject{Time: t.In(tz), Timezone: tz}
}

func (d *DateTimeObject) Type() string     { return "object" }
func (d *DateTimeObject) ToBool() bool     { return true }
func (d *DateTimeObject) ToInt() int64     { return d.Time.Unix() }
func (d *DateTimeObject) ToFloat() float64 { return float64(d.Time.UnixNano()) / 1e9 }
func (d *DateTimeObject) ToString() string { return d.Time.Format("2006-01-02 15:04:05") }
func (d *DateTimeObject) Inspect() string {
	return fmt.Sprintf("object(DateTime)#%p", d)
}

// DateTimeImmutableObject is like DateTime but immutable
type DateTimeImmutableObject struct {
	Time     time.Time
	Timezone *time.Location
}

func NewDateTimeImmutable(t time.Time, tz *time.Location) *DateTimeImmutableObject {
	if tz == nil {
		tz = time.Local
	}
	return &DateTimeImmutableObject{Time: t.In(tz), Timezone: tz}
}

func (d *DateTimeImmutableObject) Type() string     { return "object" }
func (d *DateTimeImmutableObject) ToBool() bool     { return true }
func (d *DateTimeImmutableObject) ToInt() int64     { return d.Time.Unix() }
func (d *DateTimeImmutableObject) ToFloat() float64 { return float64(d.Time.UnixNano()) / 1e9 }
func (d *DateTimeImmutableObject) ToString() string { return d.Time.Format("2006-01-02 15:04:05") }
func (d *DateTimeImmutableObject) Inspect() string {
	return fmt.Sprintf("object(DateTimeImmutable)#%p", d)
}

// DateTimeZoneObject represents a PHP DateTimeZone object
type DateTimeZoneObject struct {
	Location *time.Location
	Name     string
}

func NewDateTimeZone(name string) (*DateTimeZoneObject, error) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// Try common timezone abbreviations
		loc, err = parseTimezoneAbbrev(name)
		if err != nil {
			return nil, err
		}
	}
	return &DateTimeZoneObject{Location: loc, Name: name}, nil
}

func (d *DateTimeZoneObject) Type() string     { return "object" }
func (d *DateTimeZoneObject) ToBool() bool     { return true }
func (d *DateTimeZoneObject) ToInt() int64     { return 0 }
func (d *DateTimeZoneObject) ToFloat() float64 { return 0 }
func (d *DateTimeZoneObject) ToString() string { return d.Name }
func (d *DateTimeZoneObject) Inspect() string {
	return fmt.Sprintf("object(DateTimeZone)#%p", d)
}

func parseTimezoneAbbrev(abbrev string) (*time.Location, error) {
	abbrev = strings.ToUpper(abbrev)
	offsets := map[string]int{
		"UTC": 0, "GMT": 0,
		"EST": -5 * 3600, "EDT": -4 * 3600,
		"CST": -6 * 3600, "CDT": -5 * 3600,
		"MST": -7 * 3600, "MDT": -6 * 3600,
		"PST": -8 * 3600, "PDT": -7 * 3600,
		"CET": 1 * 3600, "CEST": 2 * 3600,
		"EET": 2 * 3600, "EEST": 3 * 3600,
		"JST": 9 * 3600,
		"IST": 5*3600 + 30*60,
		"AEST": 10 * 3600, "AEDT": 11 * 3600,
	}
	if offset, ok := offsets[abbrev]; ok {
		return time.FixedZone(abbrev, offset), nil
	}
	return nil, fmt.Errorf("unknown timezone: %s", abbrev)
}

// DateIntervalObject represents a PHP DateInterval object
type DateIntervalObject struct {
	Years    int
	Months   int
	Days     int
	Hours    int
	Minutes  int
	Seconds  int
	Invert   bool // true if negative interval
	Duration time.Duration
}

func NewDateInterval(spec string) (*DateIntervalObject, error) {
	interval := &DateIntervalObject{}

	// Parse ISO 8601 duration format: P[n]Y[n]M[n]DT[n]H[n]M[n]S
	if !strings.HasPrefix(spec, "P") {
		return nil, fmt.Errorf("invalid interval format: %s", spec)
	}

	spec = spec[1:] // Remove P prefix

	// Split on T for date and time parts
	parts := strings.SplitN(spec, "T", 2)
	datePart := parts[0]
	timePart := ""
	if len(parts) > 1 {
		timePart = parts[1]
	}

	// Parse date part
	re := regexp.MustCompile(`(\d+)([YMD])`)
	matches := re.FindAllStringSubmatch(datePart, -1)
	for _, m := range matches {
		val, _ := strconv.Atoi(m[1])
		switch m[2] {
		case "Y":
			interval.Years = val
		case "M":
			interval.Months = val
		case "D":
			interval.Days = val
		}
	}

	// Parse time part
	re = regexp.MustCompile(`(\d+)([HMS])`)
	matches = re.FindAllStringSubmatch(timePart, -1)
	for _, m := range matches {
		val, _ := strconv.Atoi(m[1])
		switch m[2] {
		case "H":
			interval.Hours = val
		case "M":
			interval.Minutes = val
		case "S":
			interval.Seconds = val
		}
	}

	// Calculate total duration (approximate, ignores variable month lengths)
	interval.Duration = time.Duration(interval.Years)*365*24*time.Hour +
		time.Duration(interval.Months)*30*24*time.Hour +
		time.Duration(interval.Days)*24*time.Hour +
		time.Duration(interval.Hours)*time.Hour +
		time.Duration(interval.Minutes)*time.Minute +
		time.Duration(interval.Seconds)*time.Second

	return interval, nil
}

func NewDateIntervalFromDiff(t1, t2 time.Time) *DateIntervalObject {
	diff := t2.Sub(t1)
	invert := diff < 0
	if invert {
		diff = -diff
	}

	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60

	return &DateIntervalObject{
		Years:    days / 365,
		Months:   (days % 365) / 30,
		Days:     (days % 365) % 30,
		Hours:    hours,
		Minutes:  minutes,
		Seconds:  seconds,
		Invert:   invert,
		Duration: diff,
	}
}

func (d *DateIntervalObject) Type() string     { return "object" }
func (d *DateIntervalObject) ToBool() bool     { return true }
func (d *DateIntervalObject) ToInt() int64     { return int64(d.Duration.Seconds()) }
func (d *DateIntervalObject) ToFloat() float64 { return d.Duration.Seconds() }
func (d *DateIntervalObject) ToString() string { return d.Format("%r%yY %mM %dD %hH %iM %sS") }
func (d *DateIntervalObject) Inspect() string {
	return fmt.Sprintf("object(DateInterval)#%p", d)
}

func (d *DateIntervalObject) Format(format string) string {
	result := format
	result = strings.ReplaceAll(result, "%Y", fmt.Sprintf("%02d", d.Years))
	result = strings.ReplaceAll(result, "%y", fmt.Sprintf("%d", d.Years))
	result = strings.ReplaceAll(result, "%M", fmt.Sprintf("%02d", d.Months))
	result = strings.ReplaceAll(result, "%m", fmt.Sprintf("%d", d.Months))
	result = strings.ReplaceAll(result, "%D", fmt.Sprintf("%02d", d.Days))
	result = strings.ReplaceAll(result, "%d", fmt.Sprintf("%d", d.Days))
	result = strings.ReplaceAll(result, "%H", fmt.Sprintf("%02d", d.Hours))
	result = strings.ReplaceAll(result, "%h", fmt.Sprintf("%d", d.Hours))
	result = strings.ReplaceAll(result, "%I", fmt.Sprintf("%02d", d.Minutes))
	result = strings.ReplaceAll(result, "%i", fmt.Sprintf("%d", d.Minutes))
	result = strings.ReplaceAll(result, "%S", fmt.Sprintf("%02d", d.Seconds))
	result = strings.ReplaceAll(result, "%s", fmt.Sprintf("%d", d.Seconds))
	result = strings.ReplaceAll(result, "%a", fmt.Sprintf("%d", int(d.Duration.Hours()/24)))
	if d.Invert {
		result = strings.ReplaceAll(result, "%R", "-")
		result = strings.ReplaceAll(result, "%r", "-")
	} else {
		result = strings.ReplaceAll(result, "%R", "+")
		result = strings.ReplaceAll(result, "%r", "")
	}
	return result
}

// isDateTimeClass checks if a class name is a DateTime-related class
func isDateTimeClass(name string) bool {
	switch name {
	case "DateTime", "DateTimeImmutable", "DateTimeZone", "DateInterval":
		return true
	}
	return false
}

// handleDateTimeNew creates a new DateTime-related object
func (i *Interpreter) handleDateTimeNew(className string, args []runtime.Value) runtime.Value {
	switch className {
	case "DateTime":
		t := time.Now()
		tz := time.Local

		if len(args) >= 1 && args[0] != runtime.NULL {
			dateStr := args[0].ToString()
			if dateStr != "" && dateStr != "now" {
				parsed, err := parseDateTimeString(dateStr)
				if err != nil {
					return runtime.NewError(fmt.Sprintf("DateTime::__construct(): Failed to parse time string (%s)", dateStr))
				}
				t = parsed
			}
		}

		if len(args) >= 2 && args[1] != runtime.NULL {
			if tzObj, ok := args[1].(*DateTimeZoneObject); ok {
				tz = tzObj.Location
			} else {
				tzName := args[1].ToString()
				loc, err := time.LoadLocation(tzName)
				if err != nil {
					loc, err = parseTimezoneAbbrev(tzName)
					if err != nil {
						return runtime.NewError(fmt.Sprintf("DateTime::__construct(): Unknown timezone: %s", tzName))
					}
				}
				tz = loc
			}
		}

		return NewDateTime(t, tz)

	case "DateTimeImmutable":
		t := time.Now()
		tz := time.Local

		if len(args) >= 1 && args[0] != runtime.NULL {
			dateStr := args[0].ToString()
			if dateStr != "" && dateStr != "now" {
				parsed, err := parseDateTimeString(dateStr)
				if err != nil {
					return runtime.NewError(fmt.Sprintf("DateTimeImmutable::__construct(): Failed to parse time string (%s)", dateStr))
				}
				t = parsed
			}
		}

		if len(args) >= 2 && args[1] != runtime.NULL {
			if tzObj, ok := args[1].(*DateTimeZoneObject); ok {
				tz = tzObj.Location
			}
		}

		return NewDateTimeImmutable(t, tz)

	case "DateTimeZone":
		if len(args) < 1 {
			return runtime.NewError("DateTimeZone::__construct() expects exactly 1 parameter")
		}
		tzName := args[0].ToString()
		tz, err := NewDateTimeZone(tzName)
		if err != nil {
			return runtime.NewError(fmt.Sprintf("DateTimeZone::__construct(): Unknown or bad timezone (%s)", tzName))
		}
		return tz

	case "DateInterval":
		if len(args) < 1 {
			return runtime.NewError("DateInterval::__construct() expects exactly 1 parameter")
		}
		spec := args[0].ToString()
		interval, err := NewDateInterval(spec)
		if err != nil {
			return runtime.NewError(fmt.Sprintf("DateInterval::__construct(): Unknown or bad format (%s)", spec))
		}
		return interval
	}

	return runtime.NewError(fmt.Sprintf("unknown DateTime class: %s", className))
}

// handleDateTimeStaticCall handles static method calls on DateTime classes
func (i *Interpreter) handleDateTimeStaticCall(className, methodName string, args []runtime.Value) runtime.Value {
	switch className {
	case "DateTime":
		switch methodName {
		case "createFromFormat":
			if len(args) < 2 {
				return runtime.FALSE
			}
			format := args[0].ToString()
			dateStr := args[1].ToString()

			goFormat := phpToGoFormat(format)
			t, err := time.Parse(goFormat, dateStr)
			if err != nil {
				return runtime.FALSE
			}

			tz := time.Local
			if len(args) >= 3 && args[2] != runtime.NULL {
				if tzObj, ok := args[2].(*DateTimeZoneObject); ok {
					tz = tzObj.Location
				}
			}

			return NewDateTime(t, tz)

		case "createFromTimestamp":
			if len(args) < 1 {
				return runtime.FALSE
			}
			timestamp := args[0].ToInt()
			return NewDateTime(time.Unix(timestamp, 0), time.Local)

		case "getLastErrors":
			result := runtime.NewArray()
			result.Set(runtime.NewString("warning_count"), runtime.NewInt(0))
			result.Set(runtime.NewString("warnings"), runtime.NewArray())
			result.Set(runtime.NewString("error_count"), runtime.NewInt(0))
			result.Set(runtime.NewString("errors"), runtime.NewArray())
			return result
		}

	case "DateTimeImmutable":
		switch methodName {
		case "createFromFormat":
			if len(args) < 2 {
				return runtime.FALSE
			}
			format := args[0].ToString()
			dateStr := args[1].ToString()

			goFormat := phpToGoFormat(format)
			t, err := time.Parse(goFormat, dateStr)
			if err != nil {
				return runtime.FALSE
			}

			tz := time.Local
			if len(args) >= 3 && args[2] != runtime.NULL {
				if tzObj, ok := args[2].(*DateTimeZoneObject); ok {
					tz = tzObj.Location
				}
			}

			return NewDateTimeImmutable(t, tz)

		case "createFromMutable":
			if len(args) < 1 {
				return runtime.FALSE
			}
			if dt, ok := args[0].(*DateTimeObject); ok {
				return NewDateTimeImmutable(dt.Time, dt.Timezone)
			}
			return runtime.FALSE
		}

	case "DateInterval":
		switch methodName {
		case "createFromDateString":
			if len(args) < 1 {
				return runtime.FALSE
			}
			spec := args[0].ToString()
			duration, err := parseRelativeDateString(spec)
			if err != nil {
				return runtime.FALSE
			}
			return &DateIntervalObject{Duration: duration}
		}
	}

	return runtime.NewError(fmt.Sprintf("undefined static method: %s::%s", className, methodName))
}

// callDateTimeMethod handles method calls on DateTime objects
func (i *Interpreter) callDateTimeMethod(obj runtime.Value, methodName string, args []runtime.Value) runtime.Value {
	switch o := obj.(type) {
	case *DateTimeObject:
		return i.callDateTimeObjectMethod(o, methodName, args)
	case *DateTimeImmutableObject:
		return i.callDateTimeImmutableMethod(o, methodName, args)
	case *DateTimeZoneObject:
		return i.callDateTimeZoneMethod(o, methodName, args)
	case *DateIntervalObject:
		return i.callDateIntervalMethod(o, methodName, args)
	}
	return runtime.NewError("unknown DateTime object type")
}

func (i *Interpreter) callDateTimeObjectMethod(dt *DateTimeObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "format":
		if len(args) < 1 {
			return runtime.NewError("DateTime::format() expects exactly 1 parameter")
		}
		format := args[0].ToString()
		return runtime.NewString(formatDateTime(dt.Time, format))

	case "modify":
		if len(args) < 1 {
			return runtime.NewError("DateTime::modify() expects exactly 1 parameter")
		}
		modifier := args[0].ToString()
		modified, err := modifyDateTimeValue(dt.Time, modifier)
		if err != nil {
			return runtime.FALSE
		}
		dt.Time = modified
		return dt

	case "add":
		if len(args) < 1 {
			return runtime.NewError("DateTime::add() expects exactly 1 parameter")
		}
		if interval, ok := args[0].(*DateIntervalObject); ok {
			dt.Time = dt.Time.Add(interval.Duration)
			return dt
		}
		return runtime.FALSE

	case "sub":
		if len(args) < 1 {
			return runtime.NewError("DateTime::sub() expects exactly 1 parameter")
		}
		if interval, ok := args[0].(*DateIntervalObject); ok {
			dt.Time = dt.Time.Add(-interval.Duration)
			return dt
		}
		return runtime.FALSE

	case "diff":
		if len(args) < 1 {
			return runtime.NewError("DateTime::diff() expects at least 1 parameter")
		}
		var t2 time.Time
		switch other := args[0].(type) {
		case *DateTimeObject:
			t2 = other.Time
		case *DateTimeImmutableObject:
			t2 = other.Time
		default:
			return runtime.FALSE
		}
		return NewDateIntervalFromDiff(dt.Time, t2)

	case "getTimestamp":
		return runtime.NewInt(dt.Time.Unix())

	case "setTimestamp":
		if len(args) < 1 {
			return runtime.NewError("DateTime::setTimestamp() expects exactly 1 parameter")
		}
		timestamp := args[0].ToInt()
		dt.Time = time.Unix(timestamp, 0).In(dt.Timezone)
		return dt

	case "getTimezone":
		return &DateTimeZoneObject{Location: dt.Timezone, Name: dt.Timezone.String()}

	case "setTimezone":
		if len(args) < 1 {
			return runtime.NewError("DateTime::setTimezone() expects exactly 1 parameter")
		}
		if tz, ok := args[0].(*DateTimeZoneObject); ok {
			dt.Time = dt.Time.In(tz.Location)
			dt.Timezone = tz.Location
			return dt
		}
		return runtime.FALSE

	case "setDate":
		if len(args) < 3 {
			return runtime.NewError("DateTime::setDate() expects exactly 3 parameters")
		}
		year := int(args[0].ToInt())
		month := time.Month(args[1].ToInt())
		day := int(args[2].ToInt())
		dt.Time = time.Date(year, month, day, dt.Time.Hour(), dt.Time.Minute(), dt.Time.Second(), dt.Time.Nanosecond(), dt.Timezone)
		return dt

	case "setTime":
		if len(args) < 2 {
			return runtime.NewError("DateTime::setTime() expects at least 2 parameters")
		}
		hour := int(args[0].ToInt())
		minute := int(args[1].ToInt())
		second := 0
		if len(args) >= 3 {
			second = int(args[2].ToInt())
		}
		microsecond := 0
		if len(args) >= 4 {
			microsecond = int(args[3].ToInt())
		}
		dt.Time = time.Date(dt.Time.Year(), dt.Time.Month(), dt.Time.Day(), hour, minute, second, microsecond*1000, dt.Timezone)
		return dt

	case "getOffset":
		_, offset := dt.Time.Zone()
		return runtime.NewInt(int64(offset))

	case "__toString":
		return runtime.NewString(dt.Time.Format("2006-01-02 15:04:05"))
	}

	return runtime.NewError(fmt.Sprintf("undefined method: DateTime::%s", methodName))
}

func (i *Interpreter) callDateTimeImmutableMethod(dt *DateTimeImmutableObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "format":
		if len(args) < 1 {
			return runtime.NewError("DateTimeImmutable::format() expects exactly 1 parameter")
		}
		format := args[0].ToString()
		return runtime.NewString(formatDateTime(dt.Time, format))

	case "modify":
		if len(args) < 1 {
			return runtime.NewError("DateTimeImmutable::modify() expects exactly 1 parameter")
		}
		modifier := args[0].ToString()
		modified, err := modifyDateTimeValue(dt.Time, modifier)
		if err != nil {
			return runtime.FALSE
		}
		// Return a new DateTimeImmutable (immutable)
		return NewDateTimeImmutable(modified, dt.Timezone)

	case "add":
		if len(args) < 1 {
			return runtime.NewError("DateTimeImmutable::add() expects exactly 1 parameter")
		}
		if interval, ok := args[0].(*DateIntervalObject); ok {
			return NewDateTimeImmutable(dt.Time.Add(interval.Duration), dt.Timezone)
		}
		return runtime.FALSE

	case "sub":
		if len(args) < 1 {
			return runtime.NewError("DateTimeImmutable::sub() expects exactly 1 parameter")
		}
		if interval, ok := args[0].(*DateIntervalObject); ok {
			return NewDateTimeImmutable(dt.Time.Add(-interval.Duration), dt.Timezone)
		}
		return runtime.FALSE

	case "diff":
		if len(args) < 1 {
			return runtime.NewError("DateTimeImmutable::diff() expects at least 1 parameter")
		}
		var t2 time.Time
		switch other := args[0].(type) {
		case *DateTimeObject:
			t2 = other.Time
		case *DateTimeImmutableObject:
			t2 = other.Time
		default:
			return runtime.FALSE
		}
		return NewDateIntervalFromDiff(dt.Time, t2)

	case "getTimestamp":
		return runtime.NewInt(dt.Time.Unix())

	case "setTimestamp":
		if len(args) < 1 {
			return runtime.NewError("DateTimeImmutable::setTimestamp() expects exactly 1 parameter")
		}
		timestamp := args[0].ToInt()
		return NewDateTimeImmutable(time.Unix(timestamp, 0).In(dt.Timezone), dt.Timezone)

	case "getTimezone":
		return &DateTimeZoneObject{Location: dt.Timezone, Name: dt.Timezone.String()}

	case "setTimezone":
		if len(args) < 1 {
			return runtime.NewError("DateTimeImmutable::setTimezone() expects exactly 1 parameter")
		}
		if tz, ok := args[0].(*DateTimeZoneObject); ok {
			return NewDateTimeImmutable(dt.Time.In(tz.Location), tz.Location)
		}
		return runtime.FALSE

	case "setDate":
		if len(args) < 3 {
			return runtime.NewError("DateTimeImmutable::setDate() expects exactly 3 parameters")
		}
		year := int(args[0].ToInt())
		month := time.Month(args[1].ToInt())
		day := int(args[2].ToInt())
		newTime := time.Date(year, month, day, dt.Time.Hour(), dt.Time.Minute(), dt.Time.Second(), dt.Time.Nanosecond(), dt.Timezone)
		return NewDateTimeImmutable(newTime, dt.Timezone)

	case "setTime":
		if len(args) < 2 {
			return runtime.NewError("DateTimeImmutable::setTime() expects at least 2 parameters")
		}
		hour := int(args[0].ToInt())
		minute := int(args[1].ToInt())
		second := 0
		if len(args) >= 3 {
			second = int(args[2].ToInt())
		}
		newTime := time.Date(dt.Time.Year(), dt.Time.Month(), dt.Time.Day(), hour, minute, second, 0, dt.Timezone)
		return NewDateTimeImmutable(newTime, dt.Timezone)

	case "getOffset":
		_, offset := dt.Time.Zone()
		return runtime.NewInt(int64(offset))

	case "__toString":
		return runtime.NewString(dt.Time.Format("2006-01-02 15:04:05"))
	}

	return runtime.NewError(fmt.Sprintf("undefined method: DateTimeImmutable::%s", methodName))
}

func (i *Interpreter) callDateTimeZoneMethod(tz *DateTimeZoneObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "getName":
		return runtime.NewString(tz.Name)

	case "getOffset":
		if len(args) < 1 {
			return runtime.NewError("DateTimeZone::getOffset() expects exactly 1 parameter")
		}
		var t time.Time
		switch dt := args[0].(type) {
		case *DateTimeObject:
			t = dt.Time
		case *DateTimeImmutableObject:
			t = dt.Time
		default:
			return runtime.FALSE
		}
		_, offset := t.In(tz.Location).Zone()
		return runtime.NewInt(int64(offset))

	case "getLocation":
		result := runtime.NewArray()
		result.Set(runtime.NewString("country_code"), runtime.NewString(""))
		result.Set(runtime.NewString("latitude"), runtime.NewFloat(0))
		result.Set(runtime.NewString("longitude"), runtime.NewFloat(0))
		result.Set(runtime.NewString("comments"), runtime.NewString(""))
		return result

	case "__toString":
		return runtime.NewString(tz.Name)
	}

	return runtime.NewError(fmt.Sprintf("undefined method: DateTimeZone::%s", methodName))
}

func (i *Interpreter) callDateIntervalMethod(interval *DateIntervalObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "format":
		if len(args) < 1 {
			return runtime.NewError("DateInterval::format() expects exactly 1 parameter")
		}
		format := args[0].ToString()
		return runtime.NewString(interval.Format(format))
	}

	return runtime.NewError(fmt.Sprintf("undefined method: DateInterval::%s", methodName))
}

// Helper functions

func parseDateTimeString(dateStr string) (time.Time, error) {
	// Common date formats to try
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02",
		"02-Jan-2006",
		"02 Jan 2006",
		"Jan 02, 2006",
		"January 02, 2006",
		"02/01/2006",
		"01/02/2006",
		"2006/01/02",
		time.RFC3339,
		time.RFC1123,
		time.RFC822,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	// Try relative date strings
	return parseRelativeDate(dateStr)
}

func parseRelativeDate(dateStr string) (time.Time, error) {
	now := time.Now()
	lower := strings.ToLower(strings.TrimSpace(dateStr))

	switch lower {
	case "now":
		return now, nil
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "tomorrow":
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()), nil
	case "yesterday":
		return time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location()), nil
	}

	// Try to parse relative expressions like "+1 day", "-2 weeks", etc.
	duration, err := parseRelativeDateString(lower)
	if err == nil {
		return now.Add(duration), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse date string: %s", dateStr)
}

func parseRelativeDateString(spec string) (time.Duration, error) {
	spec = strings.ToLower(strings.TrimSpace(spec))

	re := regexp.MustCompile(`([+-]?\d+)\s*(second|seconds|sec|secs|minute|minutes|min|mins|hour|hours|day|days|week|weeks|month|months|year|years)`)
	matches := re.FindAllStringSubmatch(spec, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid relative date string: %s", spec)
	}

	var total time.Duration
	for _, m := range matches {
		val, _ := strconv.Atoi(m[1])
		unit := m[2]

		var d time.Duration
		switch {
		case strings.HasPrefix(unit, "second"), strings.HasPrefix(unit, "sec"):
			d = time.Duration(val) * time.Second
		case strings.HasPrefix(unit, "minute"), strings.HasPrefix(unit, "min"):
			d = time.Duration(val) * time.Minute
		case strings.HasPrefix(unit, "hour"):
			d = time.Duration(val) * time.Hour
		case strings.HasPrefix(unit, "day"):
			d = time.Duration(val) * 24 * time.Hour
		case strings.HasPrefix(unit, "week"):
			d = time.Duration(val) * 7 * 24 * time.Hour
		case strings.HasPrefix(unit, "month"):
			d = time.Duration(val) * 30 * 24 * time.Hour
		case strings.HasPrefix(unit, "year"):
			d = time.Duration(val) * 365 * 24 * time.Hour
		}
		total += d
	}

	return total, nil
}

func modifyDateTimeValue(t time.Time, modifier string) (time.Time, error) {
	duration, err := parseRelativeDateString(modifier)
	if err != nil {
		return t, err
	}
	return t.Add(duration), nil
}

func formatDateTime(t time.Time, format string) string {
	result := ""
	escapeNext := false

	for _, c := range format {
		if escapeNext {
			result += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		switch c {
		// Day
		case 'd':
			result += fmt.Sprintf("%02d", t.Day())
		case 'D':
			result += t.Weekday().String()[:3]
		case 'j':
			result += fmt.Sprintf("%d", t.Day())
		case 'l':
			result += t.Weekday().String()
		case 'N':
			wd := int(t.Weekday())
			if wd == 0 {
				wd = 7
			}
			result += fmt.Sprintf("%d", wd)
		case 'S':
			day := t.Day()
			switch {
			case day == 1 || day == 21 || day == 31:
				result += "st"
			case day == 2 || day == 22:
				result += "nd"
			case day == 3 || day == 23:
				result += "rd"
			default:
				result += "th"
			}
		case 'w':
			result += fmt.Sprintf("%d", t.Weekday())
		case 'z':
			result += fmt.Sprintf("%d", t.YearDay()-1)
		// Week
		case 'W':
			_, week := t.ISOWeek()
			result += fmt.Sprintf("%02d", week)
		// Month
		case 'F':
			result += t.Month().String()
		case 'm':
			result += fmt.Sprintf("%02d", t.Month())
		case 'M':
			result += t.Month().String()[:3]
		case 'n':
			result += fmt.Sprintf("%d", t.Month())
		case 't':
			result += fmt.Sprintf("%d", daysInMonth(t.Year(), t.Month()))
		// Year
		case 'L':
			if isLeapYear(t.Year()) {
				result += "1"
			} else {
				result += "0"
			}
		case 'o':
			year, _ := t.ISOWeek()
			result += fmt.Sprintf("%d", year)
		case 'Y':
			result += fmt.Sprintf("%d", t.Year())
		case 'y':
			result += fmt.Sprintf("%02d", t.Year()%100)
		// Time
		case 'a':
			if t.Hour() < 12 {
				result += "am"
			} else {
				result += "pm"
			}
		case 'A':
			if t.Hour() < 12 {
				result += "AM"
			} else {
				result += "PM"
			}
		case 'g':
			h := t.Hour() % 12
			if h == 0 {
				h = 12
			}
			result += fmt.Sprintf("%d", h)
		case 'G':
			result += fmt.Sprintf("%d", t.Hour())
		case 'h':
			h := t.Hour() % 12
			if h == 0 {
				h = 12
			}
			result += fmt.Sprintf("%02d", h)
		case 'H':
			result += fmt.Sprintf("%02d", t.Hour())
		case 'i':
			result += fmt.Sprintf("%02d", t.Minute())
		case 's':
			result += fmt.Sprintf("%02d", t.Second())
		case 'u':
			result += fmt.Sprintf("%06d", t.Nanosecond()/1000)
		case 'v':
			result += fmt.Sprintf("%03d", t.Nanosecond()/1000000)
		// Timezone
		case 'e':
			result += t.Location().String()
		case 'O':
			_, offset := t.Zone()
			sign := "+"
			if offset < 0 {
				sign = "-"
				offset = -offset
			}
			result += fmt.Sprintf("%s%02d%02d", sign, offset/3600, (offset%3600)/60)
		case 'P':
			_, offset := t.Zone()
			sign := "+"
			if offset < 0 {
				sign = "-"
				offset = -offset
			}
			result += fmt.Sprintf("%s%02d:%02d", sign, offset/3600, (offset%3600)/60)
		case 'T':
			name, _ := t.Zone()
			result += name
		case 'Z':
			_, offset := t.Zone()
			result += fmt.Sprintf("%d", offset)
		// Full Date/Time
		case 'c':
			result += t.Format(time.RFC3339)
		case 'r':
			result += t.Format(time.RFC1123Z)
		case 'U':
			result += fmt.Sprintf("%d", t.Unix())
		default:
			result += string(c)
		}
	}

	return result
}

func phpToGoFormat(format string) string {
	replacements := map[byte]string{
		'd': "02",      // Day, 2 digits
		'j': "2",       // Day, no leading zero
		'D': "Mon",     // Day name, 3 letters
		'l': "Monday",  // Day name, full
		'm': "01",      // Month, 2 digits
		'n': "1",       // Month, no leading zero
		'M': "Jan",     // Month name, 3 letters
		'F': "January", // Month name, full
		'Y': "2006",    // Year, 4 digits
		'y': "06",      // Year, 2 digits
		'H': "15",      // Hour, 24-hour, 2 digits
		'G': "15",      // Hour, 24-hour, no leading zero (Go doesn't support this well)
		'h': "03",      // Hour, 12-hour, 2 digits
		'g': "3",       // Hour, 12-hour, no leading zero
		'i': "04",      // Minute, 2 digits
		's': "05",      // Second, 2 digits
		'A': "PM",      // AM/PM
		'a': "pm",      // am/pm
		'O': "-0700",   // Timezone offset
		'P': "-07:00",  // Timezone offset with colon
		'T': "MST",     // Timezone abbreviation
	}

	result := ""
	for i := 0; i < len(format); i++ {
		if format[i] == '\\' && i+1 < len(format) {
			result += string(format[i+1])
			i++
			continue
		}
		if repl, ok := replacements[format[i]]; ok {
			result += repl
		} else {
			result += string(format[i])
		}
	}
	return result
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
