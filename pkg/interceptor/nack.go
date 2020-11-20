// +build !js

package interceptor

import (
	"github.com/pion/webrtc/v3"
)

// NACK interceptor generates/responds to nack messages.
type NACK struct {
	webrtc.NoOpInterceptor
}

// ConfigureNack will setup everything necessary for handling generating/responding to nack messages.
func ConfigureNack(mediaEngine *webrtc.MediaEngine, interceptorRegistry *webrtc.InterceptorRegistry) error {
	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack"}, webrtc.RTPCodecTypeVideo)
	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack", Parameter: "pli"}, webrtc.RTPCodecTypeVideo)
	interceptorRegistry.Add(&NACK{})
	return nil
}

// BindRemoteTrack lets you modify any incoming RTP packets. It is called once for per RemoteTrack. The returned method
// will be called once per rtp packet.
func (n *NACK) BindRemoteTrack(ctx *webrtc.TrackRemoteContext, read webrtc.ReadRTP) webrtc.ReadRTP {
	return read
}
