package slu

import (
	"errors"

	"github.com/google/uuid"
)

var emptyID = uuid.UUID{}

// AudioContext represents a single SLU audio context, which can have multiple segments.
// See Speechly documentation for more information about audio contexts.
type AudioContext struct {
	ID          uuid.UUID `json:"id"`
	Segments    Segments  `json:"segments"`
	IsFinalised bool      `json:"is_finalised"`
}

// NewAudioContext returns a new AudioContext.
func NewAudioContext() AudioContext {
	return AudioContext{
		Segments: make(map[int32]Segment),
	}
}

// SetID sets the ID to the context.
func (c *AudioContext) SetID(id string) error {
	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	c.ID = u
	return nil
}

// CheckID checks if provided id matches current context.
func (c AudioContext) CheckID(id string) error {
	if c.ID == emptyID || c.ID.String() == id {
		return nil
	}

	return errors.New("mismatched context ID")
}

// AddTranscript adds a transcript to the specific segment of the context.
func (c *AudioContext) AddTranscript(segmentID int32, t Transcript) error {
	s := c.Segments.Get(segmentID)
	if err := s.AddTranscript(t); err != nil {
		return err
	}

	c.Segments[segmentID] = s
	return nil
}

// AddEntity adds an entity to the specific segment of the context.
func (c *AudioContext) AddEntity(segmentID int32, e Entity) error {
	s := c.Segments.Get(segmentID)
	if err := s.AddEntity(e); err != nil {
		return err
	}

	c.Segments[segmentID] = s
	return nil
}

// SetIntent sets the intent to the specific segment of the context.
func (c *AudioContext) SetIntent(segmentID int32, i Intent) error {
	s := c.Segments.Get(segmentID)
	if err := s.SetIntent(i); err != nil {
		return err
	}

	c.Segments[segmentID] = s
	return nil
}

// FinaliseSegment finalises specific segment in the context.
func (c *AudioContext) FinaliseSegment(segmentID int32) error {
	if s, ok := c.Segments[segmentID]; ok {
		if err := s.Finalise(); err != nil {
			return err
		}

		c.Segments[segmentID] = s
		return nil
	}

	return errors.New("cannot finalise non-existing segment")
}

// Finalise finalises the audio context by finalising all segments in it.
func (c *AudioContext) Finalise() error {
	for _, s := range c.Segments {
		if err := s.Finalise(); err != nil {
			return err
		}
	}

	if len(c.Segments) == 0 {
		return errors.New("finalised context has no segments")
	}

	c.IsFinalised = true

	return nil
}
