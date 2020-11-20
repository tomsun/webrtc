// +build !js

package webrtc

import (
	"io"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

// WriteRTP is used by Interceptor.BindLocalTrack.
type WriteRTP func(p *rtp.Packet, attributes map[interface{}]interface{}) (int, error)

// ReadRTP is used by Interceptor.BindRemoteTrack.
type ReadRTP func() (*rtp.Packet, map[interface{}]interface{}, error)

// WriteRTCP is used by Interceptor.BindWriteRTCP.
type WriteRTCP func(pkts []rtcp.Packet, attributes map[interface{}]interface{}) (int, error)

// ReadRTCP is used by Interceptor.BindReadRTCP.
type ReadRTCP func() ([]rtcp.Packet, map[interface{}]interface{}, error)

// Interceptor can be used to add functionality to you PeerConnections by modifying any incoming/outgoing rtp/rtcp
// packets, or sending your own packets as needed.
type Interceptor interface {

	// BindReadRTCP lets you modify any incoming RTCP packets. It is called once per sender/receiver, however this might
	// change in the future. The returned method will be called once per packet batch.
	BindReadRTCP(read ReadRTCP) ReadRTCP

	// BindWriteRTCP lets you modify any outgoing RTCP packets. It is called once per PeerConnection. The returned method
	// will be called once per packet batch.
	BindWriteRTCP(write WriteRTCP) WriteRTCP

	// BindLocalTrack lets you modify any outgoing RTP packets. It is called once for per LocalTrack. The returned method
	// will be called once per rtp packet.
	BindLocalTrack(ctx *TrackLocalContext, write WriteRTP) WriteRTP

	// UnbindLocalTrack is called when the Track is removed. It can be used to clean up any data related to that track.
	UnbindLocalTrack(ctx *TrackLocalContext)

	// BindRemoteTrack lets you modify any incoming RTP packets. It is called once for per RemoteTrack. The returned method
	// will be called once per rtp packet.
	BindRemoteTrack(ctx *TrackRemoteContext, read ReadRTP) ReadRTP

	// UnbindRemoteTrack is called when the Track is removed. It can be used to clean up any data related to that track.
	UnbindRemoteTrack(ctx *TrackRemoteContext)

	io.Closer
}

// NoOpInterceptor is an Interceptor that does not modify any packets. It can embedded in other interceptors, so it's
// possible to implement only a subset of the methods.
type NoOpInterceptor struct{}

// BindReadRTCP lets you modify any incoming RTCP packets. It is called once per sender/receiver, however this might
// change in the future. The returned method will be called once per packet batch.
func (i *NoOpInterceptor) BindReadRTCP(read ReadRTCP) ReadRTCP {
	return read
}

// BindWriteRTCP lets you modify any outgoing RTCP packets. It is called once per PeerConnection. The returned method
// will be called once per packet batch.
func (i *NoOpInterceptor) BindWriteRTCP(write WriteRTCP) WriteRTCP {
	return write
}

// BindLocalTrack lets you modify any outgoing RTP packets. It is called once for per LocalTrack. The returned method
// will be called once per rtp packet.
func (i *NoOpInterceptor) BindLocalTrack(_ *TrackLocalContext, write WriteRTP) WriteRTP {
	return write
}

// UnbindLocalTrack is called when the Track is removed. It can be used to clean up any data related to that track.
func (i *NoOpInterceptor) UnbindLocalTrack(_ *TrackLocalContext) {}

// BindRemoteTrack lets you modify any incoming RTP packets. It is called once for per RemoteTrack. The returned method
// will be called once per rtp packet.
func (i *NoOpInterceptor) BindRemoteTrack(_ *TrackRemoteContext, read ReadRTP) ReadRTP {
	return read
}

// UnbindRemoteTrack is called when the Track is removed. It can be used to clean up any data related to that track.
func (i *NoOpInterceptor) UnbindRemoteTrack(_ *TrackRemoteContext) {}

// Close closes the Interceptor, cleaning up any data if necessary.
func (i *NoOpInterceptor) Close() error {
	return nil
}
