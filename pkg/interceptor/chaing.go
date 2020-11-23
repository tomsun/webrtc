// +build !js

package interceptor

import (
	"github.com/pion/webrtc/v3/internal/util"
)

// Chain is an interceptor that runs all child interceptors in order.
type Chain struct {
	interceptors []Interceptor
}

// NewChain returns a new Chain interceptor.
func NewChain(interceptors []Interceptor) *Chain {
	return &Chain{interceptors: interceptors}
}

// BindRTCPReader lets you modify any incoming RTCP packets. It is called once per sender/receiver, however this might
// change in the future. The returned method will be called once per packet batch.
func (i *Chain) BindRTCPReader(reader RTCPReader) RTCPReader {
	for _, interceptor := range i.interceptors {
		reader = interceptor.BindRTCPReader(reader)
	}

	return reader
}

// BindRTCPWriter lets you modify any outgoing RTCP packets. It is called once per PeerConnection. The returned method
// will be called once per packet batch.
func (i *Chain) BindRTCPWriter(writer RTCPWriter) RTCPWriter {
	for _, interceptor := range i.interceptors {
		writer = interceptor.BindRTCPWriter(writer)
	}

	return writer
}

// BindLocalTrack lets you modify any outgoing RTP packets. It is called once for per LocalTrack. The returned method
// will be called once per rtp packet.
func (i *Chain) BindLocalTrack(ctx *TrackInfo, writer RTPWriter) RTPWriter {
	for _, interceptor := range i.interceptors {
		writer = interceptor.BindLocalTrack(ctx, writer)
	}

	return writer
}

// UnbindLocalTrack is called when the Track is removed. It can be used to clean up any data related to that track.
func (i *Chain) UnbindLocalTrack(ctx *TrackInfo) {
	for _, interceptor := range i.interceptors {
		interceptor.UnbindLocalTrack(ctx)
	}
}

// BindRemoteTrack lets you modify any incoming RTP packets. It is called once for per RemoteTrack. The returned method
// will be called once per rtp packet.
func (i *Chain) BindRemoteTrack(ctx *TrackInfo, reader RTPReader) RTPReader {
	for _, interceptor := range i.interceptors {
		reader = interceptor.BindRemoteTrack(ctx, reader)
	}

	return reader
}

// UnbindRemoteTrack is called when the Track is removed. It can be used to clean up any data related to that track.
func (i *Chain) UnbindRemoteTrack(ctx *TrackInfo) {
	for _, interceptor := range i.interceptors {
		interceptor.UnbindRemoteTrack(ctx)
	}
}

// Close closes the Interceptor, cleaning up any data if necessary.
func (i *Chain) Close() error {
	var errs []error
	for _, interceptor := range i.interceptors {
		if err := interceptor.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return util.FlattenErrs(errs)
}
