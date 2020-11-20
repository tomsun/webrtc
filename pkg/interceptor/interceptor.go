// +build !js

// Package interceptor contains useful default interceptors that should be safe to use in most cases.
package interceptor

import (
	"github.com/pion/webrtc/v3"
)

// RegisterDefaults will register some useful interceptors. If you want to customize which interceptors are loaded,
// you should copy the code from this method and remove unwanted interceptors.
func RegisterDefaults(mediaEngine *webrtc.MediaEngine, interceptorRegistry *webrtc.InterceptorRegistry) error {
	err := ConfigureNack(mediaEngine, interceptorRegistry)
	if err != nil {
		return err
	}

	return nil
}
