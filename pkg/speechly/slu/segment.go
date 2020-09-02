package slu

import (
	"errors"

	"github.com/speechly/slu-client/internal/json"
)

// Segment represents a single SLU segment, which is bounded by a single SLU intent.
// See Speechly documentation for more information about segments.
type Segment struct {
	ID          int32       `json:"id"`
	IsFinalised bool        `json:"is_finalised"`
	Transcripts Transcripts `json:"transcripts"`
	Entities    Entities    `json:"entities"`
	Intent      Intent      `json:"intent"`
}

// NewSegment returns a new Segment with specified ID.
func NewSegment(id int32) Segment {
	return Segment{
		Transcripts: make(map[int32]Transcript),
		Entities:    make(map[EntityIndex]Entity),
	}
}

// AddTranscript adds a transcript to a segment.
// This cannot be called after segment was finalised.
func (s *Segment) AddTranscript(t Transcript) error {
	if s.IsFinalised {
		return errors.New("cannot add a transcript to finalised segment")
	}

	if v, ok := s.Transcripts[t.Index]; ok && v.IsFinalised {
		return errors.New("cannot override finalised transcript")
	}

	s.Transcripts[t.Index] = t

	return nil
}

// AddEntity adds an entity to a segment.
func (s *Segment) AddEntity(e Entity) error {
	if s.IsFinalised {
		return errors.New("cannot add an entity to finalised segment")
	}

	i := EntityIndex{e.StartIndex, e.EndIndex}
	if v, ok := s.Entities[i]; ok && v.IsFinalised {
		return errors.New("cannot override finalised entity")
	}

	s.Entities[i] = e

	return nil
}

// SetIntent sets the intent of a segment.
func (s *Segment) SetIntent(i Intent) error {
	if s.IsFinalised {
		return errors.New("cannot add an intent to finalised segment")
	}

	if s.Intent.IsFinalised {
		return errors.New("cannot override finalised intent")
	}

	s.Intent = i
	return nil
}

// Finalise finalises the segment, by setting the IsFinalised flag to true
// and removing all tentative intents and transcripts from the segment.
// If the segment does not have at least one finalised transcript, an error is returned.
func (s *Segment) Finalise() error {
	if s.IsFinalised {
		return nil
	}

	for k, t := range s.Transcripts {
		if !t.IsFinalised {
			delete(s.Transcripts, k)
		}
	}

	if len(s.Transcripts) == 0 {
		return errors.New("finalised segment has no transcripts")
	}

	for k, e := range s.Entities {
		if !e.IsFinalised {
			delete(s.Entities, k)
		}
	}

	s.IsFinalised = true

	return nil
}

// Segments is a map of SLU segments.
type Segments map[int32]Segment

// NewSegments returns a new Segments constructed from s.
func NewSegments(s []Segment) Segments {
	r := make(Segments, len(s))

	for _, v := range s {
		r[v.ID] = v
	}

	return r
}

// Get retrieves a Segment by id.
// If a Segment with given id is not present, it will instead be added to s and returned.
func (s Segments) Get(id int32) Segment {
	seg, ok := s[id]
	if !ok {
		seg = NewSegment(id)
		s[id] = seg
	}

	return seg
}

// MarshalJSON implements json.Marshaler.
func (s Segments) MarshalJSON() ([]byte, error) {
	ser, err := json.NewArraySerialiser(len(s) * 100)
	if err != nil {
		return nil, err
	}

	for _, t := range s {
		if err := ser.Write(t); err != nil {
			return nil, err
		}
	}

	return ser.Finalise()
}
