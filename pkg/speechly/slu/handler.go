package slu

import (
	"bytes"
	"context"
	"errors"
	"io"

	"golang.org/x/sync/errgroup"

	"speechly/slu-client/pkg/logger"
	"speechly/slu-client/pkg/speechly"
)

var (
	startReq = speechly.SLURequest{
		StreamingRequest: &speechly.SLURequest_Event{Event: &speechly.SLUEvent{Event: speechly.SLUEvent_START}},
	}
	stopReq = speechly.SLURequest{
		StreamingRequest: &speechly.SLURequest_Event{Event: &speechly.SLUEvent{Event: speechly.SLUEvent_STOP}},
	}
)

// AudioContextHandler is a handler for a single AudioContext stream.
// It handles the sending of audio data by starting a new audio context, sending audio data and stopping the context.
// It handles the receiving of SLU responses by processing them into an AudioContext state
// and exposing an API to access the snapshots of that state.
//
// An example of such state progression is something like this:
//
// * First state:
// {
//   "id": "ae275a56-66b6-49a3-addb-90df03edb946",
//   "segments": [],
//   "is_finalised": false
// }
//
// * Second state:
// {
//   "id": "ae275a56-66b6-49a3-addb-90df03edb946",
//   "segments": [
//     {
//       "id": 0,
//       "is_finalised": false,
//       "transcripts": [
//         {
//           "word": "TURN",
//           "index": 2,
//           "start_time": 480,
//           "end_time": 720,
//           "is_finalised": false
//         }
//       ],
//       "entities": [],
//       "intent": {
//         "value": "",
//         "is_finalised": false
//       }
//     }
//   ],
//   "is_finalised": false
// }
//
// * Third state:
// {
//   "id": "ae275a56-66b6-49a3-addb-90df03edb946",
//   "segments": [
//     {
//       "id": 0,
//       "is_finalised": false,
//       "transcripts": [
//         {
//           "word": "TURN",
//           "index": 2,
//           "start_time": 480,
//           "end_time": 900,
//           "is_finalised": false
//         },
//         {
//           "word": "OFF",
//           "index": 3,
//           "start_time": 900,
//           "end_time": 1020,
//           "is_finalised": false
//         }
//       ],
//       "entities": [],
//       "intent": {
//         "value": "turn_off",
//         "is_finalised": true
//       }
//     }
//   ],
//   "is_finalised": false
// }
//
// And so forth, until eventually the context is finalised by the API.
type AudioContextHandler interface {
	// Read reads the next AudioContext state from the handler.
	// If the context has been stopped, this will return io.EOF.
	// If any error has happened while handling the context, this will return it.
	Read() (AudioContext, error)

	// Close closes the handler by signalling it to send the StopContext event and exit the loop.
	Close() error
}

type ctxHandler struct {
	str      speechly.SLU_StreamClient
	src      AudioSource
	res      chan AudioContext
	log      logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	doneFunc func()
	runErr   error
}

func newCtxHandler(
	ctx context.Context, str speechly.SLU_StreamClient, src AudioSource, chanSize int, log logger.Logger, done func(),
) (*ctxHandler, error) {
	if err := str.Send(&startReq); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	r := &ctxHandler{
		str:      str,
		src:      src,
		res:      make(chan AudioContext, chanSize),
		log:      log,
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
		doneFunc: done,
	}

	go r.run()

	return r, nil
}

func (r *ctxHandler) Read() (AudioContext, error) {
	select {
	case c, more := <-r.res:
		if !more {
			return AudioContext{}, io.EOF
		}

		return c, nil
	case <-r.done:
		if r.runErr == nil {
			return AudioContext{}, io.EOF
		}

		return AudioContext{}, r.runErr
	}
}

func (r *ctxHandler) Close() error {
	r.cancel()
	<-r.done
	return r.runErr
}

// nolint: funlen, gocognit, gocyclo // It's a long function, but most of it is just a switch case.
func (r *ctxHandler) run() {
	defer func() {
		r.cancel() // Make sure we cancel context to avoid leaking it, if Close() is never called.
		close(r.done)
		go r.doneFunc()
	}()

	g, ctx := errgroup.WithContext(r.ctx)

	g.Go(func() error {
		defer func() {
			if err := r.str.Send(&stopReq); err != nil {
				r.log.Warn("failed to send stop request to API", err)
			}

			if err := r.src.Close(); err != nil {
				r.log.Warn("failed to close audio source", err)
			}
		}()

		var (
			buf = bytes.Buffer{}
			req = speechly.SLURequest_Audio{}
			msg = speechly.SLURequest{StreamingRequest: &req}
		)

		for done := false; !done; {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				buf.Reset()

				_, err := r.src.WriteTo(&buf)
				if err == io.EOF {
					done = true
				} else if err != nil {
					return err
				}

				req.Audio = buf.Bytes()
				if err := r.str.Send(&msg); err != nil {
					return err
				}
			}
		}

		return nil
	})

	g.Go(func() error {
		defer close(r.res)

		var (
			cn = NewAudioContext()
			t  = Transcript{}
			e  = Entity{}
			i  = Intent{}
		)

		for done := false; !done; {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				res, err := r.str.Recv()
				if err == io.EOF {
					return errors.New("unexpected io.EOF from API")
				}

				if err != nil {
					return err
				}

				var (
					id  = res.GetAudioContext()
					sid = res.GetSegmentId()
				)

				if err := cn.CheckID(id); err != nil {
					return err
				}

				switch v := res.GetStreamingResponse().(type) {
				case *speechly.SLUResponse_Transcript:
					if err := t.Parse(v.Transcript, false); err != nil {
						return err
					}

					if err := cn.AddTranscript(sid, t); err != nil {
						return err
					}
				case *speechly.SLUResponse_Entity:
					if err := e.Parse(v.Entity, false); err != nil {
						return err
					}

					if err := cn.AddEntity(sid, e); err != nil {
						return err
					}
				case *speechly.SLUResponse_Intent:
					if err := i.Parse(v.Intent, false); err != nil {
						return err
					}

					if err := cn.SetIntent(sid, i); err != nil {
						return err
					}

				case *speechly.SLUResponse_TentativeTranscript:
					for _, v := range v.TentativeTranscript.GetTentativeWords() {
						if err := t.Parse(v, true); err != nil {
							return err
						}

						if err := cn.AddTranscript(sid, t); err != nil {
							return err
						}
					}
				case *speechly.SLUResponse_TentativeEntities:
					for _, v := range v.TentativeEntities.GetTentativeEntities() {
						if err := e.Parse(v, true); err != nil {
							return err
						}

						if err := cn.AddEntity(sid, e); err != nil {
							return err
						}
					}
				case *speechly.SLUResponse_TentativeIntent:
					if err := i.Parse(v.TentativeIntent, true); err != nil {
						return err
					}

					if err := cn.SetIntent(sid, i); err != nil {
						return err
					}
				case *speechly.SLUResponse_SegmentEnd:
					if err := cn.FinaliseSegment(sid); err != nil {
						return err
					}
				case *speechly.SLUResponse_Started:
					if err := cn.SetID(res.GetAudioContext()); err != nil {
						return err
					}
				case *speechly.SLUResponse_Finished:
					if err := cn.Finalise(); err != nil {
						return err
					}

					done = true
				default:
					return errors.New("unknown response type")
				}

				select {
				case r.res <- cn:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		r.runErr = err
	}
}
