package command

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kr/text"
	"github.com/mitchellh/cli"
	"github.com/posener/complete"

	"github.com/ryanuber/columnize"
)

// maxLineLength is the maximum width of any line.
const maxLineLength int = 78

// FormatKV takes a set of strings and formats them into properly
// aligned k = v pairs using the columnize library.
func FormatKV(in []string) string {
	columnConf := columnize.DefaultConfig()
	columnConf.Empty = "<none>"
	columnConf.Glue = " = "
	return columnize.Format(in, columnConf)
}

// FormatList takes a set of strings and formats them into properly
// aligned output, replacing any blank fields with a placeholder
// for awk-ability.
func FormatList(in []string) string {
	columnConf := columnize.DefaultConfig()
	columnConf.Empty = "<none>"
	return columnize.Format(in, columnConf)
}

// FormatListWithSpaces takes a set of strings and formats them into properly
// aligned output. It should be used sparingly since it doesn't replace empty
// values and hence not awk/sed friendly
func FormatListWithSpaces(in []string) string {
	columnConf := columnize.DefaultConfig()
	return columnize.Format(in, columnConf)
}

// Limit returns a string at the max length specified.
func Limit(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length]
}

// WrapAtLengthWithPadding wraps the given text at the maxLineLength, taking
// into account any provided left padding.
func WrapAtLengthWithPadding(s string, pad int) string {
	wrapped := text.Wrap(s, maxLineLength-pad)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", pad) + line
	}
	return strings.Join(lines, "\n")
}

// WrapAtLength wraps the given text to maxLineLength.
func WrapAtLength(s string) string {
	return WrapAtLengthWithPadding(s, 0)
}

// formatTime formats the time to string based on RFC822
func FormatTime(t time.Time) string {
	if t.Unix() < 1 {
		// It's more confusing to display the UNIX epoch or a zero value than nothing
		return ""
	}
	// Return ISO_8601 time format GH-3806
	return t.Format("2006-01-02T15:04:05Z07:00")
}

// formatUnixNanoTime is a helper for formatting time for output.
func FormatUnixNanoTime(nano int64) string {
	t := time.Unix(0, nano)
	return FormatTime(t)
}

// formatTimeDifference takes two times and determines their duration difference
// truncating to a passed unit.
// E.g. formatTimeDifference(first=1m22s33ms, second=1m28s55ms, time.Second) -> 6s
func FormatTimeDifference(first, second time.Time, d time.Duration) string {
	return second.Truncate(d).Sub(first.Truncate(d)).String()
}

// fmtInt formats v into the tail of buf.
// It returns the index where the output begins.
func FmtInt(buf []byte, v uint64) int {
	w := len(buf)
	for v > 0 {
		w--
		buf[w] = byte(v%10) + '0'
		v /= 10
	}
	return w
}

// PrettyTimeDiff prints a human readable time difference.
// It uses abbreviated forms for each period - s for seconds, m for minutes, h for hours,
// d for days, mo for months, and y for years. Time difference is rounded to the nearest second,
// and the top two least granular periods are returned. For example, if the time difference
// is 10 months, 12 days, 3 hours and 2 seconds, the string "10mo12d" is returned. Zero values return the empty string
func PrettyTimeDiff(first, second time.Time) string {
	// handle zero values
	if first.IsZero() || first.UnixNano() == 0 {
		return ""
	}
	// round to the nearest second
	first = first.Round(time.Second)
	second = second.Round(time.Second)

	// calculate time difference in seconds
	var d time.Duration
	messageSuffix := "ago"
	if second.Equal(first) || second.After(first) {
		d = second.Sub(first)
	} else {
		d = first.Sub(second)
		messageSuffix = "from now"
	}

	u := uint64(d.Seconds())

	var buf [32]byte
	w := len(buf)
	secs := u % 60

	// track indexes of various periods
	var indexes []int

	if secs > 0 {
		w--
		buf[w] = 's'
		// u is now seconds
		w = FmtInt(buf[:w], secs)
		indexes = append(indexes, w)
	}
	u /= 60
	// u is now minutes
	if u > 0 {
		mins := u % 60
		if mins > 0 {
			w--
			buf[w] = 'm'
			w = FmtInt(buf[:w], mins)
			indexes = append(indexes, w)
		}
		u /= 60
		// u is now hours
		if u > 0 {
			hrs := u % 24
			if hrs > 0 {
				w--
				buf[w] = 'h'
				w = FmtInt(buf[:w], hrs)
				indexes = append(indexes, w)
			}
			u /= 24
		}
		// u is now days
		if u > 0 {
			days := u % 30
			if days > 0 {
				w--
				buf[w] = 'd'
				w = FmtInt(buf[:w], days)
				indexes = append(indexes, w)
			}
			u /= 30
		}
		// u is now months
		if u > 0 {
			months := u % 12
			if months > 0 {
				w--
				buf[w] = 'o'
				w--
				buf[w] = 'm'
				w = FmtInt(buf[:w], months)
				indexes = append(indexes, w)
			}
			u /= 12
		}
		// u is now years
		if u > 0 {
			w--
			buf[w] = 'y'
			w = FmtInt(buf[:w], u)
			indexes = append(indexes, w)
		}
	}
	start := w
	end := len(buf)

	// truncate to the first two periods
	num_periods := len(indexes)
	if num_periods > 2 {
		end = indexes[num_periods-3]
	}
	if start == end { //edge case when time difference is less than a second
		return "0s " + messageSuffix
	} else {
		return string(buf[start:end]) + " " + messageSuffix
	}

}

// MergeAutocompleteFlags is used to join multiple flag completion sets.
func MergeAutocompleteFlags(flags ...complete.Flags) complete.Flags {
	merged := make(map[string]complete.Predictor, len(flags))
	for _, f := range flags {
		for k, v := range f {
			merged[k] = v
		}
	}
	return merged
}

// CommandErrorText is used to easily render the same messaging across commads
// when an error is printed.
func CommandErrorText(cmd NamedCommand) string {
	appName := os.Getenv("CLI_APP_NAME")
	return fmt.Sprintf("For additional help try '%s %s --help'", appName, cmd.Name())
}

// uiErrorWriter is a io.Writer that wraps underlying ui.ErrorWriter().
// ui.ErrorWriter expects full lines as inputs and it emits its own line breaks.
//
// uiErrorWriter scans input for individual lines to pass to ui.ErrorWriter. If data
// doesn't contain a new line, it buffers result until next new line or writer is closed.
type uiErrorWriter struct {
	ui  cli.Ui
	buf bytes.Buffer
}

func (w *uiErrorWriter) Write(data []byte) (int, error) {
	read := 0
	for len(data) != 0 {
		a, token, err := bufio.ScanLines(data, false)
		if err != nil {
			return read, err
		}

		if a == 0 {
			r, err := w.buf.Write(data)
			return read + r, err
		}

		w.ui.Error(w.buf.String() + string(token))
		data = data[a:]
		w.buf.Reset()
		read += a
	}

	return read, nil
}

func (w *uiErrorWriter) Close() error {
	// emit what's remaining
	if w.buf.Len() != 0 {
		w.ui.Error(w.buf.String())
		w.buf.Reset()
	}
	return nil
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
