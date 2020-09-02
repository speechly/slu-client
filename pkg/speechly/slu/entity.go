package slu

import (
	"github.com/speechly/slu-client/internal/json"
	"github.com/speechly/slu-client/pkg/speechly"
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

// EntityIndex is a struct used for indexing entities.
type EntityIndex struct {
	StartIndex, EndIndex int32
}

// Entities is a map of SLU entities.
type Entities map[EntityIndex]Entity

// NewEntities returns new Entities constructed from e.
func NewEntities(e []Entity) Entities {
	r := make(Entities, len(e))
	i := EntityIndex{}

	for _, v := range e {
		i.StartIndex = v.StartIndex
		i.EndIndex = v.EndIndex

		r[i] = v
	}

	return r
}

// MarshalJSON implements json.Marshaler.
func (e Entities) MarshalJSON() ([]byte, error) {
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
