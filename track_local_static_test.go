// +build !js

package webrtc

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/transport/test"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/stretchr/testify/assert"
)

// If a remote doesn't support a Codec used by a `TrackLocalStatic`
// an error should be returned to the user
func Test_TrackLocalStatic_NoCodecIntersection(t *testing.T) {
	lim := test.TimeOut(time.Second * 30)
	defer lim.Stop()

	report := test.CheckRoutines(t)
	defer report()

	track, err := NewTrackLocalStaticSample(RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion")
	assert.NoError(t, err)

	t.Run("Offerer", func(t *testing.T) {
		pc, err := NewPeerConnection(Configuration{})
		assert.NoError(t, err)

		noCodecPC, err := NewAPI().NewPeerConnection(Configuration{})
		assert.NoError(t, err)

		_, err = pc.AddTrack(track)
		assert.NoError(t, err)

		assert.True(t, errors.Is(signalPair(pc, noCodecPC), ErrUnsupportedCodec))

		assert.NoError(t, noCodecPC.Close())
		assert.NoError(t, pc.Close())
	})

	t.Run("Answerer", func(t *testing.T) {
		pc, err := NewPeerConnection(Configuration{})
		assert.NoError(t, err)

		m := &MediaEngine{}
		assert.NoError(t, m.RegisterCodec(RTPCodecParameters{
			RTPCodecCapability: RTPCodecCapability{MimeType: "video/VP9", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        96,
		}, RTPCodecTypeVideo))

		vp9OnlyPC, err := NewAPI(WithMediaEngine(m)).NewPeerConnection(Configuration{})
		assert.NoError(t, err)

		_, err = vp9OnlyPC.AddTransceiverFromKind(RTPCodecTypeVideo)
		assert.NoError(t, err)

		_, err = pc.AddTrack(track)
		assert.NoError(t, err)

		assert.True(t, errors.Is(signalPair(vp9OnlyPC, pc), ErrUnsupportedCodec))

		assert.NoError(t, vp9OnlyPC.Close())
		assert.NoError(t, pc.Close())
	})

	t.Run("Local", func(t *testing.T) {
		offerer, answerer, err := newPair()
		assert.NoError(t, err)

		invalidCodecTrack, err := NewTrackLocalStaticSample(RTPCodecCapability{MimeType: "video/invalid-codec"}, "video", "pion")
		assert.NoError(t, err)

		_, err = offerer.AddTrack(invalidCodecTrack)
		assert.NoError(t, err)

		assert.True(t, errors.Is(signalPair(offerer, answerer), ErrUnsupportedCodec))
		assert.NoError(t, offerer.Close())
		assert.NoError(t, answerer.Close())
	})
}

// Test for states around PeerConnection Media error behavior
//
// * A Track that hasn't been Binded to anything should return an error
// * A Track should be Unbinded when the PeerConnection is closed
func Test_TrackLocalStatic_Closed(t *testing.T) {
	lim := test.TimeOut(time.Second * 30)
	defer lim.Stop()

	report := test.CheckRoutines(t)
	defer report()

	pcOffer, pcAnswer, err := newPair()
	assert.NoError(t, err)

	_, err = pcAnswer.AddTransceiverFromKind(RTPCodecTypeVideo)
	assert.NoError(t, err)

	vp8Writer, err := NewTrackLocalStaticSample(RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion")
	assert.NoError(t, err)

	_, err = pcOffer.AddTrack(vp8Writer)
	assert.NoError(t, err)

	answerChan := make(chan *TrackRemote)
	pcAnswer.OnTrack(func(t *TrackRemote, r *RTPReceiver) {
		answerChan <- t
	})

	assert.NoError(t, signalPair(pcOffer, pcAnswer))

	vp8Reader := func() *TrackRemote {
		for {
			assert.NoError(t, vp8Writer.WriteSample(media.Sample{Data: []byte{0x00}, Duration: time.Second}))
			time.Sleep(time.Millisecond * 25)

			select {
			case track := <-answerChan:
				return track
			default:
				continue
			}
		}
	}()

	closeChan := make(chan error)
	go func() {
		time.Sleep(time.Second)
		closeChan <- pcAnswer.Close()
	}()

	// First read will succeed because first packet is cached
	// for Payload probing
	_, err = vp8Reader.Read(make([]byte, 1))
	assert.NoError(t, err)

	_, err = vp8Reader.Read(make([]byte, 1))
	assert.True(t, errors.Is(err, io.EOF))

	assert.NoError(t, <-closeChan)

	assert.NoError(t, pcOffer.Close())
	assert.NoError(t, pcAnswer.Close())

	if err = vp8Writer.WriteSample(media.Sample{Data: []byte{0x00}, Duration: time.Second}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("Write to TrackLocal with no RTPSenders did not return io.ErrClosedPipe")
	} else if err = pcAnswer.WriteRTCP([]rtcp.Packet{&rtcp.RapidResynchronizationRequest{SenderSSRC: 0, MediaSSRC: 0}}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("WriteRTCP to closed PeerConnection did not return io.ErrClosedPipe")
	}
}

func Test_TrackLocalStatic_PayloadType(t *testing.T) {
	lim := test.TimeOut(time.Second * 30)
	defer lim.Stop()

	report := test.CheckRoutines(t)
	defer report()

	mediaEngineOne := &MediaEngine{}
	assert.NoError(t, mediaEngineOne.RegisterCodec(RTPCodecParameters{
		RTPCodecCapability: RTPCodecCapability{MimeType: "video/VP8", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        100,
	}, RTPCodecTypeVideo))

	mediaEngineTwo := &MediaEngine{}
	assert.NoError(t, mediaEngineTwo.RegisterCodec(RTPCodecParameters{
		RTPCodecCapability: RTPCodecCapability{MimeType: "video/VP8", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        200,
	}, RTPCodecTypeVideo))

	offerer, err := NewAPI(WithMediaEngine(mediaEngineOne)).NewPeerConnection(Configuration{})
	assert.NoError(t, err)

	answerer, err := NewAPI(WithMediaEngine(mediaEngineTwo)).NewPeerConnection(Configuration{})
	assert.NoError(t, err)

	track, err := NewTrackLocalStaticSample(RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion")
	assert.NoError(t, err)

	_, err = offerer.AddTransceiverFromKind(RTPCodecTypeVideo)
	assert.NoError(t, err)

	_, err = answerer.AddTrack(track)
	assert.NoError(t, err)

	onTrackFired, onTrackFiredFunc := context.WithCancel(context.Background())
	offerer.OnTrack(func(track *TrackRemote, r *RTPReceiver) {
		assert.Equal(t, track.PayloadType(), PayloadType(100))
		assert.Equal(t, track.Codec().RTPCodecCapability.MimeType, "video/VP8")

		onTrackFiredFunc()
	})

	assert.NoError(t, signalPair(offerer, answerer))

	sendVideoUntilDone(onTrackFired.Done(), t, []*TrackLocalStaticSample{track})

	assert.NoError(t, offerer.Close())
	assert.NoError(t, answerer.Close())
}
