package slu

import (
	"sort"

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

// EntityIndexList is a list of entity indices.
// It implements the sort.Interface, so it can be used for obtaining sorted list of entity indices.
type EntityIndexList []EntityIndex

// SortedEntityIndexList returns a sorted version of EntityIndexList constructed from Entities.
func SortedEntityIndexList(e Entities) EntityIndexList {
	r := make(EntityIndexList, 0, len(e))
	for k := range e {
		r = append(r, k)
	}

	sort.Sort(r)

	return r
}

func (e EntityIndexList) Len() int {
	return len(e)
}

func (e EntityIndexList) Less(i, j int) bool {
	return e[i].StartIndex < e[j].StartIndex
}

func (e EntityIndexList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

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

	for _, k := range SortedEntityIndexList(e) {
		if err := s.Write(e[k]); err != nil {
			return nil, err
		}
	}

	return s.Finalise()
}
