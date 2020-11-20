// +build !js

package webrtc

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/transport/test"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/stretchr/testify/assert"
)

type testInterceptor struct {
	t           *testing.T
	extensionID uint8
	writeRTCP   atomic.Value
	lastRTCP    atomic.Value
	NoOpInterceptor
}

func (t *testInterceptor) BindLocalTrack(_ *TrackLocalContext, write WriteRTP) WriteRTP {
	return func(p *rtp.Packet, attributes map[interface{}]interface{}) (int, error) {
		// set extension on outgoing packet
		p.Header.Extension = true
		p.Header.ExtensionProfile = 0xBEDE
		assert.NoError(t.t, p.Header.SetExtension(t.extensionID, []byte("write")))

		return write(p, attributes)
	}
}

func (t *testInterceptor) BindRemoteTrack(ctx *TrackRemoteContext, read ReadRTP) ReadRTP {
	return func() (*rtp.Packet, map[interface{}]interface{}, error) {
		p, attributes, err := read()
		if err != nil {
			return nil, nil, err
		}
		// set extension on incoming packet
		p.Header.Extension = true
		p.Header.ExtensionProfile = 0xBEDE
		assert.NoError(t.t, p.Header.SetExtension(t.extensionID, []byte("read")))

		// write back a pli
		writeRTCP := t.writeRTCP.Load().(WriteRTCP)
		pli := &rtcp.PictureLossIndication{SenderSSRC: uint32(ctx.SSRC()), MediaSSRC: uint32(ctx.SSRC())}
		_, err = writeRTCP([]rtcp.Packet{pli}, make(map[interface{}]interface{}))
		assert.NoError(t.t, err)

		return p, attributes, nil
	}
}

func (t *testInterceptor) BindReadRTCP(read ReadRTCP) ReadRTCP {
	return func() ([]rtcp.Packet, map[interface{}]interface{}, error) {
		pkts, attributes, err := read()
		if err != nil {
			return nil, nil, err
		}

		t.lastRTCP.Store(pkts[0])

		return pkts, attributes, nil
	}
}

func (t *testInterceptor) lastReadRTCP() rtcp.Packet {
	p, _ := t.lastRTCP.Load().(rtcp.Packet)
	return p
}

func (t *testInterceptor) BindWriteRTCP(write WriteRTCP) WriteRTCP {
	t.writeRTCP.Store(write)
	return write
}

func TestPeerConnection_Interceptor(t *testing.T) {
	to := test.TimeOut(time.Second * 20)
	defer to.Stop()

	report := test.CheckRoutines(t)
	defer report()

	createPC := func(interceptor Interceptor) *PeerConnection {
		m := &MediaEngine{}
		err := m.RegisterDefaultCodecs()
		if err != nil {
			t.Fatal(err)
		}
		ir := &InterceptorRegistry{}
		ir.Add(interceptor)
		pc, err := NewAPI(WithMediaEngine(m), WithInterceptorRegistry(ir)).NewPeerConnection(Configuration{})
		if err != nil {
			t.Fatal(err)
		}

		return pc
	}

	sendInterceptor := &testInterceptor{t: t, extensionID: 1}
	senderPC := createPC(sendInterceptor)
	receiverPC := createPC(&testInterceptor{t: t, extensionID: 2})

	track, err := NewTrackLocalStaticSample(RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion")
	if err != nil {
		t.Fatal(err)
	}

	sender, err := senderPC.AddTrack(track)
	if err != nil {
		t.Fatal(err)
	}

	pending := new(int32)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	*pending++
	receiverPC.OnTrack(func(track *TrackRemote, receiver *RTPReceiver) {
		p, readErr := track.ReadRTP()
		if readErr != nil {
			t.Fatal(readErr)
		}
		assert.Equal(t, p.Extension, true)
		assert.Equal(t, "write", string(p.GetExtension(1)))
		assert.Equal(t, "read", string(p.GetExtension(2)))
		atomic.AddInt32(pending, -1)
		wg.Done()

		for {
			_, readErr = track.ReadRTP()
			if readErr != nil {
				return
			}
		}
	})

	wg.Add(1)
	*pending++
	go func() {
		_, readErr := sender.ReadRTCP()
		assert.NoError(t, readErr)
		atomic.AddInt32(pending, -1)
		wg.Done()

		for {
			_, readErr = sender.ReadRTCP()
			if readErr != nil {
				return
			}
		}
	}()

	err = signalPair(senderPC, receiverPC)
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(time.Millisecond * 100)
			if routineErr := track.WriteSample(media.Sample{Data: []byte{0x00}, Duration: time.Second}); routineErr != nil {
				t.Error(routineErr)
				return
			}

			if atomic.LoadInt32(pending) == 0 {
				return
			}
		}
	}()

	wg.Wait()
	assert.NoError(t, senderPC.Close())
	assert.NoError(t, receiverPC.Close())

	pli, _ := sendInterceptor.lastReadRTCP().(*rtcp.PictureLossIndication)
	if pli == nil || pli.SenderSSRC == 0 {
		t.Errorf("pli not found by send interceptor")
	}
}
