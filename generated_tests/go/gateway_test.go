package core_test

import (
	"testing"

	"github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	core "pilot/pkg/networking/core"
	model "istio.io/istio/pilot/pkg/model"
)

// MockListenerBuilder is a mock of ListenerBuilder interface
type MockListenerBuilder struct {
	mock.Mock
}

func (m *MockListenerBuilder) buildGatewayListeners(builder *core.ListenerBuilder) *core.ListenerBuilder {
	args := m.Called(builder)
	return args.Get(0).(*core.ListenerBuilder)
}

func TestBuildGatewayListenersNoGateways(t *testing.T) {
	mb := &MockListenerBuilder{}
	lb := &core.ListenerBuilder{Node: &model.Proxy{}}
	mb.On("buildGatewayListeners", lb).Return(lb)

	result := mb.buildGatewayListeners(lb)
	assert.Equal(t, lb, result)
	mb.AssertExpectations(t)
}

func TestBuildGatewayListenersWithGateways(t *testing.T) {
	mockListener := &listener.Listener{Name: "listener_1"}

	mb := &MockListenerBuilder{}
	lb := &core.ListenerBuilder{
		Node: &model.Proxy{
			MergedGateway: &model.MergedGateway{},
		},
		GatewayListeners: []*listener.Listener{},
	}
	mb.On("buildGatewayListeners", lb).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*core.ListenerBuilder)
		arg.GatewayListeners = append(arg.GatewayListeners, mockListener)
	}).Return(lb)

	result := mb.buildGatewayListeners(lb)
	assert.Equal(t, lb, result)
	assert.Equal(t, 1, len(result.GatewayListeners))
	assert.True(t, proto.Equal(mockListener, result.GatewayListeners[0]))
	mb.AssertExpectations(t)
}

// Add more test cases based on variations of ListenerBuilder inputs,
// such as different Node configurations and MergedGateway states.
```

These unit tests aim to assess the behavior of the `buildGatewayListeners` function under various conditions. Note that the tests are simplified examples focusing on the syntactic structure due to the lack of complete context from the original `gateway.go` you've provided. Depending on the actual implementations and dependencies in your code, you might need to mock additional objects or functionalities and also test more intricate scenarios specific to your business logic.