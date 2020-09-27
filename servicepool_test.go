package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewServerCluster(t *testing.T) {

	opt, _ := NewOptions(10, []string{"127.0.0.1:9999"})
	sc, err := NewServerClusterWithBuilders("server1", *opt, []grpc.DialOption{grpc.WithInsecure()}, nil)
	assert.Nil(t, err)

	assert.Equal(t, "server1", sc.Name)
	assert.Equal(t, 10, sc.Pool.Cap())
	assert.Equal(t, 2, len(sc.Pool.dialOptions))
}

func TestServerCluster_GetClient(t *testing.T) {
	opt, _ := NewOptions(10, []string{"127.0.0.1:9999"})
	sc, err := NewServerCluster("server1", *opt, []grpc.DialOption{grpc.WithInsecure()})
	assert.Nil(t, err)

	_, _, err = sc.GetServerClient("default")
	assert.Equal(t, ErrClientBuilderNil, err)

	sc.SetClientBuilder("default", clientBuilder)

	client, err := sc.GetClient()

	assert.IsType(t, (*GrpcConn)(nil), client)

	serverClient, release, err := sc.GetServerClient("default")

	assert.Nil(t, err)
	assert.IsType(t, (*testingBuilder)(nil), serverClient)
	assert.IsType(t, func() {}, release)

	_, _, err = sc.GetServerClient("default0")

	assert.Equal(t, ErrServerBuilderNil, err)

	builders := make(map[string]ServerBuilderFunc)
	builders["default0"] = clientBuilder
	builders["default1"] = nil

	sc.SetClientBuilders(builders)

	_, _, err = sc.GetServerClient("default0")
	assert.Nil(t, err)

	_, _, err = sc.GetServerClient("default1")
	assert.Equal(t, ErrServerBuilderNil, err)

}

type iTestingBuilder interface {
	Read()
	Write()
}

type testingBuilder struct{}

func (tb *testingBuilder) Read()  {}
func (tb *testingBuilder) Write() {}

func clientBuilder(conn grpc.ClientConnInterface) interface{} {
	return &testingBuilder{}
}

func TestServiceCenter(t *testing.T) {
	opt, _ := NewOptions(10, []string{"127.0.0.1:9999"})
	sc, _ := NewServerCluster("server1", *opt, []grpc.DialOption{grpc.WithInsecure()})
	builders := make(map[string]ServerBuilderFunc)
	builders["default0"] = clientBuilder
	builders["default1"] = nil
	sc.SetClientBuilders(builders)

	sCenter := &ServiceCenter{}
	sCenter.Register(sc)

	cluster, ok := sCenter.Get("server1")
	assert.True(t, ok)
	assert.Equal(t, "server1", cluster.Name)
	assert.IsType(t, (*GRPCPool)(nil), cluster.Pool)
	GrpcConn, err := cluster.GetClient()
	assert.Nil(t, err)
	conn := GrpcConn.Conn()
	assert.IsType(t, (*grpc.ClientConn)(nil), conn)

	client, release, err := cluster.GetServerClient("default0")
	assert.Nil(t, err)
	client.(iTestingBuilder).Read()
	assert.IsType(t, (*testingBuilder)(nil), client)
	assert.IsType(t, func() {}, release)
}

func BenchmarkServerCluster_GetClient(b *testing.B) {
	opt, _ := NewOptions(10, []string{"127.0.0.1:9999"})
	sc, _ := NewServerCluster("server1", *opt, []grpc.DialOption{grpc.WithInsecure()})

	builders := make(map[string]ServerBuilderFunc)
	builders["default0"] = clientBuilder
	builders["default1"] = nil

	sc.SetClientBuilders(builders)
	for i := 0; i < b.N; i++ {
		_, r, e := sc.GetServerClient("default0")
		if e != nil {
			b.Error(e)
		}
		defer r()
	}

}
