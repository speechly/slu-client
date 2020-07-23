package slu

import (
	"errors"

	"speechly/slu-client/internal/json"
	"speechly/slu-client/pkg/speechly"
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

type transcripts map[int32]Transcript

func (t transcripts) MarshalJSON() ([]byte, error) {
	s, err := json.NewArraySerialiser(len(t) * 100)
	if err != nil {
		return nil, err
	}

	for _, v := range t {
		if err := s.Write(v); err != nil {
			return nil, err
		}
	}

	return s.Finalise()
}
