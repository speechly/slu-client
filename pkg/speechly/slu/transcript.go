package slu

import (
	"errors"
	"sort"

	"github.com/speechly/slu-client/internal/json"
	"github.com/speechly/slu-client/pkg/speechly"
)

var errNilValue = errors.New("cannot parse nil value")

// Transcript is a transcript of a word, detected by Speechly SLU API.
type Transcript struct {
	Word        string `json:"word"`
	Index       int32  `json:"index"`
	StartTime   int32  `json:"start_time"`
	EndTime     int32  `json:"end_time"`
	IsFinalised bool   `json:"is_finalised"`
}

// Parse parses response from API into Transcript.
func (t *Transcript) Parse(v *speechly.SLUTranscript, isTentative bool) error {
	if v == nil {
		return errNilValue
	}

	t.Word = v.Word
	t.Index = v.Index
	t.StartTime = v.StartTime
	t.EndTime = v.EndTime
	t.IsFinalised = !isTentative

	return nil
}

// Transcripts is a map of SLU transcripts.
type Transcripts map[int32]Transcript

// NewTranscripts returns a new Transcripts constructed from t.
func NewTranscripts(t []Transcript) Transcripts {
	r := make(Transcripts, len(t))

	for _, v := range t {
		r[v.Index] = v
	}

	return r
}

// MarshalJSON implements json.Marshaler.
func (t Transcripts) MarshalJSON() ([]byte, error) {
	s, err := json.NewArraySerialiser(len(t) * 100)
	if err != nil {
		return nil, err
	}

	// Get sorted keys.
	r := make([]int, 0, len(t))
	for _, v := range t {
		r = append(r, int(v.Index))
	}
	sort.Ints(r)

	for _, i := range r {
		if err := s.Write(t[int32(i)]); err != nil {
			return nil, err
		}
	}

	return s.Finalise()
}
