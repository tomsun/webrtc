// +build !js

package interceptor

// NoOp is an Interceptor that does not modify any packets. It can embedded in other interceptors, so it's
// possible to implement only a subset of the methods.
type NoOp struct{}

// BindRTCPReader lets you modify any incoming RTCP packets. It is called once per sender/receiver, however this might
// change in the future. The returned method will be called once per packet batch.
func (i *NoOp) BindRTCPReader(reader RTCPReader) RTCPReader {
	return reader
}

// BindRTCPWriter lets you modify any outgoing RTCP packets. It is called once per PeerConnection. The returned method
// will be called once per packet batch.
func (i *NoOp) BindRTCPWriter(writer RTCPWriter) RTCPWriter {
	return writer
}

// BindLocalTrack lets you modify any outgoing RTP packets. It is called once for per LocalTrack. The returned method
// will be called once per rtp packet.
func (i *NoOp) BindLocalTrack(_ *TrackInfo, writer RTPWriter) RTPWriter {
	return writer
}

// UnbindLocalTrack is called when the Track is removed. It can be used to clean up any data related to that track.
func (i *NoOp) UnbindLocalTrack(_ *TrackInfo) {}

// BindRemoteTrack lets you modify any incoming RTP packets. It is called once for per RemoteTrack. The returned method
// will be called once per rtp packet.
func (i *NoOp) BindRemoteTrack(_ *TrackInfo, reader RTPReader) RTPReader {
	return reader
}

// UnbindRemoteTrack is called when the Track is removed. It can be used to clean up any data related to that track.
func (i *NoOp) UnbindRemoteTrack(_ *TrackInfo) {}

// Close closes the Interceptor, cleaning up any data if necessary.
func (i *NoOp) Close() error {
	return nil
}
