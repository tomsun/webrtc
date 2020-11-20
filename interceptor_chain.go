// +build !js

package webrtc

import (
	"github.com/pion/webrtc/v3/internal/util"
)

type chainInterceptor struct {
	interceptors []Interceptor
}

// BindReadRTCP lets you modify any incoming RTCP packets. It is called once per sender/receiver, however this might
// change in the future. The returned method will be called once per packet batch.
func (i *chainInterceptor) BindReadRTCP(read ReadRTCP) ReadRTCP {
	for _, interceptor := range i.interceptors {
		read = interceptor.BindReadRTCP(read)
	}

	return read
}

// BindWriteRTCP lets you modify any outgoing RTCP packets. It is called once per PeerConnection. The returned method
// will be called once per packet batch.
func (i *chainInterceptor) BindWriteRTCP(write WriteRTCP) WriteRTCP {
	for _, interceptor := range i.interceptors {
		write = interceptor.BindWriteRTCP(write)
	}

	return write
}

// BindLocalTrack lets you modify any outgoing RTP packets. It is called once for per LocalTrack. The returned method
// will be called once per rtp packet.
func (i *chainInterceptor) BindLocalTrack(ctx *TrackLocalContext, write WriteRTP) WriteRTP {
	for _, interceptor := range i.interceptors {
		write = interceptor.BindLocalTrack(ctx, write)
	}

	return write
}

// UnbindLocalTrack is called when the Track is removed. It can be used to clean up any data related to that track.
func (i *chainInterceptor) UnbindLocalTrack(ctx *TrackLocalContext) {
	for _, interceptor := range i.interceptors {
		interceptor.UnbindLocalTrack(ctx)
	}
}

// BindRemoteTrack lets you modify any incoming RTP packets. It is called once for per RemoteTrack. The returned method
// will be called once per rtp packet.
func (i *chainInterceptor) BindRemoteTrack(ctx *TrackRemoteContext, read ReadRTP) ReadRTP {
	for _, interceptor := range i.interceptors {
		read = interceptor.BindRemoteTrack(ctx, read)
	}

	return read
}

// UnbindRemoteTrack is called when the Track is removed. It can be used to clean up any data related to that track.
func (i *chainInterceptor) UnbindRemoteTrack(ctx *TrackRemoteContext) {
	for _, interceptor := range i.interceptors {
		interceptor.UnbindRemoteTrack(ctx)
	}
}

// Close closes the Interceptor, cleaning up any data if necessary.
func (i *chainInterceptor) Close() error {
	var errs []error
	for _, interceptor := range i.interceptors {
		if err := interceptor.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return util.FlattenErrs(errs)
}
