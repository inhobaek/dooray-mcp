package utils

import (
	"fmt"
	"strings"
	"time"
)

type JsonTime struct {
	time.Time
	DateOnly bool
}

func NewJsonTime(t time.Time) JsonTime {
	return JsonTime{Time: t}
}

func NewJsonDate(t time.Time) JsonTime {
	return JsonTime{Time: t, DateOnly: true}
}

func (jt *JsonTime) String() string {
	if jt.DateOnly {
		return FormatDateOnly(jt.Time)
	}
	return FormatTimeToISO8601(jt.Time)
}

func (jt JsonTime) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, `"%s"`, jt.String()), nil
}

func (jt *JsonTime) UnmarshalJSON(b []byte) error {
	timeString := strings.Trim(string(b), `"`)
	t, err := ConvertISO8601ToTime(timeString)
	if err != nil {
		return fmt.Errorf("invalid date format: %s, %v", timeString, err)
	}
	jt.Time = t
	jt.DateOnly = IsDateOnlyFormat(timeString)
	return nil
}

func (jt *JsonTime) ToTime() time.Time {
	return jt.Time
}