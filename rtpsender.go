// +build !js

package webrtc

import (
	"io"
	"sync"

	"github.com/pion/randutil"
	"github.com/pion/rtcp"
)

// RTPSender allows an application to control how a given Track is encoded and transmitted to a remote peer
type RTPSender struct {
	track TrackLocal

	transport  *DTLSTransport
	srtpStream *srtpWriterFuture

	payloadType PayloadType
	ssrc        SSRC
	codec       RTPCodecParameters

	// nolint:godox
	// TODO(sgotti) remove this when in future we'll avoid replacing
	// a transceiver sender since we can just check the
	// transceiver negotiation status
	negotiated bool

	// A reference to the associated api object
	api *API
	id  string

	mu                     sync.RWMutex
	sendCalled, stopCalled chan interface{}
}

// NewRTPSender constructs a new RTPSender
func (api *API) NewRTPSender(track TrackLocal, transport *DTLSTransport) (*RTPSender, error) {
	if track == nil {
		return nil, errRTPSenderTrackNil
	} else if transport == nil {
		return nil, errRTPSenderDTLSTransportNil
	}

	id, err := randutil.GenerateCryptoRandomString(32, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if err != nil {
		return nil, err
	}

	r := &RTPSender{
		track:      track,
		transport:  transport,
		api:        api,
		sendCalled: make(chan interface{}),
		stopCalled: make(chan interface{}),
		ssrc:       SSRC(randutil.NewMathRandomGenerator().Uint32()),
		id:         id,
		srtpStream: &srtpWriterFuture{},
	}
	r.srtpStream.rtpSender = r
	return r, nil
}

func (r *RTPSender) isNegotiated() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.negotiated
}

func (r *RTPSender) setNegotiated() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.negotiated = true
}

// Transport returns the currently-configured *DTLSTransport or nil
// if one has not yet been configured
func (r *RTPSender) Transport() *DTLSTransport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.transport
}

// Track returns the RTCRtpTransceiver track, or nil
func (r *RTPSender) Track() TrackLocal {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.track
}

// ReplaceTrack replaces the track currently being used as the sender's source with a new TrackLocal.
// The new track must be of the same media kind (audio, video, etc) and switching the track should not
// require negotiation.
func (r *RTPSender) ReplaceTrack(track TrackLocal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hasSent() {
		if err := r.track.Unbind(TrackLocalContext{
			id:          r.id,
			ssrc:        r.ssrc,
			writeStream: r.srtpStream,
		}); err != nil {
			return err
		}
	}

	if !r.hasSent() || track == nil {
		r.track = track
		return nil
	}

	if _, err := track.Bind(TrackLocalContext{
		id:          r.id,
		codecs:      []RTPCodecParameters{r.codec},
		ssrc:        r.ssrc,
		writeStream: r.srtpStream,
	}); err != nil {
		return err
	}

	r.track = track
	return nil
}

// Send Attempts to set the parameters controlling the sending of media.
func (r *RTPSender) Send(parameters RTPSendParameters) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hasSent() {
		return errRTPSenderSendAlreadyCalled
	}

	var err error
	if r.codec, err = r.track.Bind(TrackLocalContext{
		id:          r.id,
		codecs:      r.api.mediaEngine.getCodecsByKind(r.track.Kind()),
		ssrc:        parameters.Encodings.SSRC,
		writeStream: r.srtpStream,
	}); err != nil {
		return err
	}

	close(r.sendCalled)
	return nil
}

// Stop irreversibly stops the RTPSender
func (r *RTPSender) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-r.stopCalled:
		return nil
	default:
	}
	close(r.stopCalled)

	if !r.hasSent() {
		return nil
	}

	return r.srtpStream.Close()
}

// Read reads incoming RTCP for this RTPReceiver
func (r *RTPSender) Read(b []byte) (n int, err error) {
	select {
	case <-r.sendCalled:
		return r.srtpStream.Read(b)
	case <-r.stopCalled:
		return 0, io.ErrClosedPipe
	}
}

// ReadRTCP is a convenience method that wraps Read and unmarshals for you
func (r *RTPSender) ReadRTCP() ([]rtcp.Packet, error) {
	b := make([]byte, receiveMTU)
	i, err := r.Read(b)
	if err != nil {
		return nil, err
	}

	return rtcp.Unmarshal(b[:i])
}

// hasSent tells if data has been ever sent for this instance
func (r *RTPSender) hasSent() bool {
	select {
	case <-r.sendCalled:
		return true
	default:
		return false
	}
}
