package api

import (
	"encoding/json"
	"strings"
	"time"
)

type dateOnly time.Time

func (d *dateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}

	*d = dateOnly(t)

	return nil
}

func (d *dateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*d))
}

func (d *dateOnly) Format(s string) string {
	t := time.Time(*d)
	return t.Format(s)
}
