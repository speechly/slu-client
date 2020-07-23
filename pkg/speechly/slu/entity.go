package slu

import (
	"speechly/slu-client/internal/json"
	"speechly/slu-client/pkg/speechly"
)

// Entity is the entity detected by SLU API.
type Entity struct {
	Type        string `json:"type"`
	Value       string `json:"value"`
	StartIndex  int32  `json:"start_index"`
	EndIndex    int32  `json:"end_index"`
	IsFinalised bool   `json:"is_finalised"`
}

// Parse parses response from API into Entity.
func (e *Entity) Parse(v *speechly.SLUEntity, isTentative bool) error {
	if v == nil {
		return errNilValue
	}

	e.Type = v.Entity
	e.Value = v.Value
	e.StartIndex = v.StartPosition
	e.EndIndex = v.EndPosition
	e.IsFinalised = !isTentative

	return nil
}

type entityIndex struct {
	Left, Right int32
}

type entities map[entityIndex]Entity

func (e entities) MarshalJSON() ([]byte, error) {
	s, err := json.NewArraySerialiser(len(e) * 100)
	if err != nil {
		return nil, err
	}

	for _, t := range e {
		if err := s.Write(t); err != nil {
			return nil, err
		}
	}

	return s.Finalise()
}
