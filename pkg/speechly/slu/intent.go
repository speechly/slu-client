package slu

import (
	sluv1 "github.com/speechly/api/go/speechly/slu/v1"
)

// Intent is the intent detected by SLU API.
type Intent struct {
	Value       string `json:"value"`
	IsFinalised bool   `json:"is_finalised"`
}

// Parse parses response from API into Intent.
func (i *Intent) Parse(v *sluv1.SLUIntent, isTentative bool) error {
	if v == nil {
		return errNilValue
	}

	i.Value = v.Intent
	i.IsFinalised = !isTentative

	return nil
}
