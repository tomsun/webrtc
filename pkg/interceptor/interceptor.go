// +build !js

// Package interceptor contains the Interceptor interface, with some useful interceptors that should be safe to use
// in most cases.
package interceptor

import (
	"io"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/interceptor/movetopionrtp"
)

// Interceptor can be used to add functionality to you PeerConnections by modifying any incoming/outgoing rtp/rtcp
// packets, or sending your own packets as needed.
type Interceptor interface {

	// BindRTCPReader lets you modify any incoming RTCP packets. It is called once per sender/receiver, however this might
	// change in the future. The returned method will be called once per packet batch.
	BindRTCPReader(reader RTCPReader) RTCPReader

	// BindRTCPWriter lets you modify any outgoing RTCP packets. It is called once per PeerConnection. The returned method
	// will be called once per packet batch.
	BindRTCPWriter(writer RTCPWriter) RTCPWriter

	// BindLocalTrack lets you modify any outgoing RTP packets. It is called once for per LocalTrack. The returned method
	// will be called once per rtp packet.
	BindLocalTrack(info *TrackInfo, writer RTPWriter) RTPWriter

	// UnbindLocalTrack is called when the Track is removed. It can be used to clean up any data related to that track.
	UnbindLocalTrack(info *TrackInfo)

	// BindRemoteTrack lets you modify any incoming RTP packets. It is called once for per RemoteTrack. The returned method
	// will be called once per rtp packet.
	BindRemoteTrack(info *TrackInfo, reader RTPReader) RTPReader

	// UnbindRemoteTrack is called when the Track is removed. It can be used to clean up any data related to that track.
	UnbindRemoteTrack(info *TrackInfo)

	io.Closer
}

// TrackInfo is the Context passed when a TrackLocal has been Binded/Unbinded from a PeerConnection, and used
// in Interceptors.
type TrackInfo struct {
	ID     string
	Params movetopionrtp.RTPParameters
	SSRC   movetopionrtp.SSRC
}

// RTPWriter is used by Interceptor.BindLocalTrack.
type RTPWriter interface {
	// Write a rtp packet
	Write(p *rtp.Packet, attributes Attributes) (int, error)
}

// RTPReader is used by Interceptor.BindRemoteTrack.
type RTPReader interface {
	// Read a rtp packet
	Read() (*rtp.Packet, Attributes, error)
}

// RTCPWriter is used by Interceptor.BindRTCPWriter.
type RTCPWriter interface {
	// Write a batch of rtcp packets
	Write(pkts []rtcp.Packet, attributes Attributes) (int, error)
}

// RTCPReader is used by Interceptor.BindRTCPReader.
type RTCPReader interface {
	// Read a batch of rtcp packets
	Read() ([]rtcp.Packet, Attributes, error)
}

type Attributes map[interface{}]interface{}

// RTPWriterFunc is an adapter for RTPWrite interface
type RTPWriterFunc func(p *rtp.Packet, attributes Attributes) (int, error)

// RTPReaderFunc is an adapter for RTPReader interface
type RTPReaderFunc func() (*rtp.Packet, Attributes, error)

// RTCPWriterFunc is an adapter for RTCPWriter interface
type RTCPWriterFunc func(pkts []rtcp.Packet, attributes Attributes) (int, error)

// RTCPReaderFunc is an adapter for RTCPReader interface
type RTCPReaderFunc func() ([]rtcp.Packet, Attributes, error)

// Write a rtp packet
func (f RTPWriterFunc) Write(p *rtp.Packet, attributes Attributes) (int, error) {
	return f(p, attributes)
}

// Read a rtp packet
func (f RTPReaderFunc) Read() (*rtp.Packet, Attributes, error) {
	return f()
}

// Write a batch of rtcp packets
func (f RTCPWriterFunc) Write(pkts []rtcp.Packet, attributes Attributes) (int, error) {
	return f(pkts, attributes)
}

// Read a batch of rtcp packets
func (f RTCPReaderFunc) Read() ([]rtcp.Packet, Attributes, error) {
	return f()
}

// Get returns the attribute associated with key.
func (a Attributes) Get(key interface{}) interface{} {
	return a[key]
}

// Set sets the attribute associated with key to the given value.
func (a Attributes) Set(key interface{}, val interface{}) {
	a[key] = val
}
