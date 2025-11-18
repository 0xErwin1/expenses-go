package service

import (
	"encoding/json"
	"strconv"
)

// FlexibleInt unmarshals JSON numbers that may arrive as string or number.
type FlexibleInt int

func (f *FlexibleInt) UnmarshalJSON(b []byte) error {
	// Handle null
	if string(b) == "null" {
		return nil
	}

	// If quoted, parse string
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*f = FlexibleInt(v)
		return nil
	}

	// Otherwise parse as number
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = FlexibleInt(n)
	return nil
}

func (f FlexibleInt) Int() int {
	return int(f)
}

// FlexibleFloat unmarshals JSON numbers that may arrive as string or number.
type FlexibleFloat float64

func (f *FlexibleFloat) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*f = FlexibleFloat(v)
		return nil
	}

	var n float64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = FlexibleFloat(n)
	return nil
}

func (f FlexibleFloat) Float64() float64 {
	return float64(f)
}
