package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewServiceCenterBuilder(t *testing.T) {
	scb := NewServiceCenterBuilder()
	err := scb.SetServerWithDefaultOptions("server1", nil, "127.0.0.1:8899")
	assert.Nil(t, err)

	dialOptions := []grpc.DialOption{grpc.WithInsecure()}

	opt, _ := NewOptions(5, []string{"127.0.0.1:9999"})

	builders := make(map[string]ServerBuilderFunc)
	builders["default"] = clientBuilder
	err = scb.SetServer("server2", builders, *opt, dialOptions)
	assert.Nil(t, err)

	sc := scb.Build()
	_, ok := sc.Get("server")
	assert.False(t, ok)

	s1, ok := sc.Get("server1")
	assert.True(t, ok)
	assert.Equal(t, 0, len(s1.clientBuilder))
	g1, err := s1.GetClient()
	assert.Nil(t, err)
	assert.IsType(t, (*GrpcConn)(nil), g1)

	s2, ok := sc.Get("server2")
	assert.True(t, ok)
	client, release, err := s2.GetServerClient("default")
	assert.Nil(t, err)
	assert.IsType(t, (*testingBuilder)(nil), client)
	assert.IsType(t, func() {}, release)

}
