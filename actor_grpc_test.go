package rungroup

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServer is a mock implementation of grpc.Server.
type MockServer struct {
	mock.Mock
}

func (m *MockServer) Serve(lis net.Listener) error {
	args := m.Called(lis)
	return args.Error(0)
}

func (m *MockServer) GracefulStop() {
	m.Called()
}

func (m *MockServer) Stop() {
	m.Called()
}

func TestGrpcServerActors(t *testing.T) {
	t.Run("start function should call Serve", func(t *testing.T) {
		addr := "127.0.0.1:0" // Use a dynamically assigned port.
		listener, err := net.Listen("tcp", addr)
		assert.NoError(t, err)
		defer listener.Close()

		mockServer := new(MockServer)
		start, _ := GrpcServerActors(mockServer, listener)

		mockServer.On("Serve", mock.Anything).Return(nil)

		err = start()
		assert.NoError(t, err)

		mockServer.AssertCalled(t, "Serve", mock.AnythingOfType("*net.TCPListener"))
	})

	t.Run("stop function should gracefully stop the server", func(t *testing.T) {
		addr := "127.0.0.1:0" // Use a dynamically assigned port.
		listener, err := net.Listen("tcp", addr)
		assert.NoError(t, err)
		defer listener.Close()

		mockServer := new(MockServer)
		_, stop := GrpcServerActors(mockServer, listener)

		mockServer.On("GracefulStop").Return()
		mockServer.On("Stop").Return()

		err = stop()
		assert.NoError(t, err)

		mockServer.AssertCalled(t, "GracefulStop")
		mockServer.AssertNotCalled(t, "Stop") // Stop should only be called if the context times out.
	})

	t.Run("stop function should force stop on timeout", func(t *testing.T) {
		addr := "127.0.0.1:0" // Use a dynamically assigned port.
		listener, err := net.Listen("tcp", addr)
		assert.NoError(t, err)
		defer listener.Close()

		mockServer := new(MockServer)
		_, stop := GrpcServerActors(mockServer, listener, WithShutdownTimeout(200*time.Millisecond))

		// Simulate a scenario where GracefulStop takes too long, forcing a call to Stop.
		mockServer.On("GracefulStop").Run(func(args mock.Arguments) {
			time.Sleep(500 * time.Millisecond) // Simulate a long delay in GracefulStop.
		}).Once()
		mockServer.On("Stop").Once()

		err = stop()
		assert.NoError(t, err)

		mockServer.AssertCalled(t, "GracefulStop")
		mockServer.AssertCalled(t, "Stop") // Stop should be called due to timeout.
	})
}
