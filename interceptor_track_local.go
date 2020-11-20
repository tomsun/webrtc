// +build !js

package webrtc

import (
	"sync/atomic"

	"github.com/pion/rtp"
)

type interceptorTrackLocalWriter struct {
	TrackLocalWriter
	writeRTP atomic.Value
}

func (i *interceptorTrackLocalWriter) setWriteRTP(writeRTP WriteRTP) {
	i.writeRTP.Store(writeRTP)
}

func (i *interceptorTrackLocalWriter) WriteRTP(header *rtp.Header, payload []byte) (int, error) {
	writeRTP := i.writeRTP.Load().(WriteRTP)

	if writeRTP == nil {
		return 0, nil
	}

	return writeRTP(&rtp.Packet{Header: *header, Payload: payload}, make(map[interface{}]interface{}))
}
