package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const dateFormat = "2006-01-02"

type Date struct {
	time.Time
}

func NewDate(t time.Time) Date {
	return Date{Time: t}
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}
	return Date{Time: t}, nil
}

func Today() Date {
	return Date{Time: time.Now().Truncate(24 * time.Hour)}
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format(dateFormat) + `"`), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` {
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}
	d.Time = t
	return nil
}

func (d *Date) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		d.Time = v
		return nil
	case string:
		t, err := time.Parse(dateFormat, v)
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Date", src)
	}
}

func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format(dateFormat), nil
}

func (d Date) String() string {
	return d.Time.Format(dateFormat)
}

func (d Date) IsZero() bool {
	return d.Time.IsZero()
}
