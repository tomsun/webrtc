// +build !js

package interceptor

// NACK interceptor generates/responds to nack messages.
type NACK struct {
	NoOp
}

// BindRemoteTrack lets you modify any incoming RTP packets. It is called once for per RemoteTrack. The returned method
// will be called once per rtp packet.
func (n *NACK) BindRemoteTrack(_ *TrackInfo, reader RTPReader) RTPReader {
	return reader
}
