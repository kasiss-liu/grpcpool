package pool

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

func TestFunctions(t *testing.T) {

	opt, _ := NewOptions(10, []string{"127.0.0.1:9999"})
	pool, _ := NewGRPCPool(opt, grpc.WithInsecure())

	createFn := defaultFactoryCreateConn()
	conn, err := createFn(pool)
	assert.Nil(t, err)
	assert.EqualValues(t, connectivity.Idle, conn.GetState())

	closeFn := defaultCloseConn()
	err = closeFn(conn)
	assert.Nil(t, err)
	assert.EqualValues(t, connectivity.Shutdown, conn.GetState())

}

func fakeConnFactory(p *GRPCPool) (*grpc.ClientConn, error) {
	return nil, nil
}

func fakeConnClose(conn *grpc.ClientConn) error {
	return nil
}

func TestNewGRPCPool(t *testing.T) {
	opt, _ := NewOptions(10, []string{"127.0.0.1:8899"})

	pool, err := NewGRPCPool(opt)
	assert.Nil(t, err)
	assert.Equal(t, 10, pool.Cap())
	assert.Equal(t, 0, pool.Len())
}

func CreateTestingGrpcPool() *GRPCPool {
	opt, _ := NewOptions(10, []string{"127.0.0.1:8899"})
	pool, _ := NewGRPCPool(opt)
	return pool
}

func CreateFakeGrpcPool() *GRPCPool {
	opt, _ := NewOptions(10, []string{"127.0.0.1:8899"})
	pool, _ := NewGRPCPool(opt, grpc.WithInsecure())
	return pool
}

func TestGRPCPool_SetConnFactory(t *testing.T) {
	pool := CreateTestingGrpcPool()
	pool.SetConnFactory(fakeConnFactory)
	assert.NotNil(t, pool.connFactory)
}

func TestGRPCPool_SetDoConnClose(t *testing.T) {
	pool := CreateTestingGrpcPool()
	pool.SetDoConnClose(fakeConnClose)
	assert.NotNil(t, pool.connDoClose)
}

func TestGRPCPool_GetDialOptions(t *testing.T) {
	pool := CreateTestingGrpcPool()
	dialOptions := pool.GetDialOptions()
	assert.Equal(t, 1, len(dialOptions))
}

func TestGRPCPool_GetOptions(t *testing.T) {
	sec5 := 5 * time.Second
	pool := CreateTestingGrpcPool()
	options := pool.GetOptions()
	assert.Equal(t, 10, options.Cap)
	assert.Equal(t, sec5, options.DialTimeout)
	assert.Equal(t, sec5, options.IdleTimeout)
	assert.Equal(t, sec5, options.PingTimeout)
	assert.True(t, options.ClientKeepAlive)
	assert.False(t, options.ForcePermit)
}

func TestGRPCPool_InitConnections(t *testing.T) {

	pool := CreateFakeGrpcPool()
	defer pool.Close()
	assert.Equal(t, 0, pool.Len())
	err := pool.InitConnections()
	assert.Nil(t, err)
	assert.Equal(t, pool.Cap(), pool.Len())

}

func TestGRPCPool_Get(t *testing.T) {

	pool := CreateFakeGrpcPool()
	_ = pool.InitConnections()
	wg := sync.WaitGroup{}
	counts := [3]int{0, 0, 0}
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(i int) {
			gConn, err := pool.Get()
			assert.Nil(t, err)
			counts[int(gConn.RefCount())]++
			wg.Done()
		}(i)
	}
	wg.Wait()
	assert.EqualValues(t, 5, pool.connNext)
	assert.EqualValues(t, 10, counts[1])
	assert.EqualValues(t, 5, counts[2])
}

func TestGRPCPool_Close(t *testing.T) {
	pool := CreateFakeGrpcPool()
	_ = pool.InitConnections()

	pool.Close()

	assert.Nil(t, pool.connPool)
	assert.Nil(t, pool.connFactory)
	assert.Nil(t, pool.connDoClose)
}

func TestGrpcConn_Conn(t *testing.T) {
	pool := CreateFakeGrpcPool()
	_ = pool.InitConnections()

	GrpcConn, _ := pool.Get()

	GrpcConn.Conn()
	assert.IsType(t, (*grpc.ClientConn)(nil), GrpcConn.Conn())
	assert.EqualValues(t, 1, GrpcConn.refcount)
}

func TestGrpcConn_Release(t *testing.T) {
	pool := CreateFakeGrpcPool()
	gConn, _ := pool.Get()

	assert.EqualValues(t, 1, gConn.RefCount())
	gConn.Release()
	assert.EqualValues(t, 0, gConn.RefCount())
}

func TestGrpcConn_Close(t *testing.T) {
	pool := CreateFakeGrpcPool()
	gConn, _ := pool.Get()
	err := gConn.Close()
	assert.Nil(t, err)
}
