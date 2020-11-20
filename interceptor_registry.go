// +build !js

package webrtc

// InterceptorRegistry is a collector for interceptors.
type InterceptorRegistry struct {
	interceptors []Interceptor
}

// Add adds a new Interceptor to the registry.
func (i *InterceptorRegistry) Add(interceptor Interceptor) {
	i.interceptors = append(i.interceptors, interceptor)
}

func (i *InterceptorRegistry) build() Interceptor {
	if len(i.interceptors) == 0 {
		return &NoOpInterceptor{}
	}

	return &chainInterceptor{interceptors: i.interceptors}
}
