package pool

import (
	"context"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

//ConnFactoryFunc type of function to create grpc conn
type ConnFactoryFunc func(p *GRPCPool) (*grpc.ClientConn, error)

//ConnCloseFunc type of function to close grpc conn
type ConnCloseFunc func(conn *grpc.ClientConn) error

//GrpcConn is a struct warp *grpc.ClientConn
type GrpcConn struct {
	conn     *grpc.ClientConn
	pool     *GRPCPool
	refcount int64
}

//Conn return the *grpc.ClientConn
func (g *GrpcConn) Conn() *grpc.ClientConn {
	return g.conn
}

//to increase num of stream on conn
func (g *GrpcConn) use() {
	atomic.AddInt64(&g.refcount, 1)
}

//Release put conn back pool or close conn when pool is full
func (g *GrpcConn) Release() {
	if g.pool.Len() > g.pool.options.Cap {
		_ = g.pool.connDoClose(g.conn)
		return
	}
	atomic.AddInt64(&g.refcount, -1)
}

//RefCount return stream count on conn
func (g *GrpcConn) RefCount() int64 {
	return g.refcount
}

//Close to close the conn
func (g *GrpcConn) Close() error {
	return g.pool.connDoClose(g.conn)
}

//GRPCPool struct of pool
//pool struct
type GRPCPool struct {
	lock        sync.Mutex
	options     *Options
	dialOptions []grpc.DialOption

	connPool    []*GrpcConn
	connNext    int
	connFactory ConnFactoryFunc
	connDoClose ConnCloseFunc
}

//SetConnFactory set factory func of create conn
func (p *GRPCPool) SetConnFactory(fn ConnFactoryFunc) {
	p.connFactory = fn
}

//SetDoConnClose set func of close conn
func (p *GRPCPool) SetDoConnClose(fn ConnCloseFunc) {
	p.connDoClose = fn
}

//GetOptions return grpc pool options
func (p *GRPCPool) GetOptions() Options {
	return *(p.options)
}

//GetDialOptions return grpc pool dial options
func (p *GRPCPool) GetDialOptions() []grpc.DialOption {
	return p.dialOptions
}

//Len return conn count
func (p *GRPCPool) Len() int {
	return len(p.connPool)
}

//Cap return conn cap
func (p *GRPCPool) Cap() int {
	return p.options.Cap
}

//InitConnections to create connections
func (p *GRPCPool) InitConnections() error {

	poolSize := p.Len()
	if poolSize > 0 {
		return ErrPoolInitialized
	}
	opt := p.options
	//init and put connections into channel
	for i := 0; i < opt.Cap; i++ {
		conn, err := p.connFactory(p)
		if err != nil {
			p.Close()
			return err
		}
		p.connPool = append(p.connPool, &GrpcConn{conn: conn})
	}
	return nil
}

//Get to get one grpc.ConnClient
func (p *GRPCPool) Get() (conn *GrpcConn, err error) {
	//check pool if closed
	if p.connPool == nil {
		return nil, ErrPoolClosed
	}
	p.lock.Lock()
	defer p.lock.Unlock()

	retries := 0
	for {
		//when the number of connections is not reached the cap
		//create new connection
		if len(p.connPool) < p.options.Cap {
			var gconn *grpc.ClientConn
			gconn, err = p.connFactory(p)
			if err != nil {
				return nil, err
			}
			conn = &GrpcConn{conn: gconn, pool: p, refcount: 1}
			p.connPool = append(p.connPool, conn)
			p.connNext = len(p.connPool)
			return
		}

		//adjust slice range to avoid array range out
		if p.connNext >= p.options.Cap {
			p.connNext = 0
		}

		conn = p.connPool[p.connNext]
		//check connection if alive
		if conn.Conn().GetState() != connectivity.Shutdown && conn.Conn().GetState() != connectivity.TransientFailure {
			conn.use()
			p.connNext++
			return
		}
		//if not available remove connection from pool
		p.connPool[p.connNext] = nil
		if p.connNext == p.Cap()-1 {
			p.connPool = p.connPool[:p.connNext]
		} else {
			p.connPool = append(p.connPool[:p.connNext], p.connPool[p.connNext+1:]...)
		}

		retries++
		if retries > 10 {
			break
		}
	}
	return nil, ErrConnConnect
}

// Close to close grpc poll
func (p *GRPCPool) Close() {

	p.lock.Lock()
	defer p.lock.Unlock()

	p.connFactory = nil
	doClose := p.connDoClose
	p.connDoClose = nil
	p.options = nil
	p.dialOptions = nil
	p.connNext = 0

	if p.connPool == nil {
		return
	}
	//to close all connections
	for _, GrpcConn := range p.connPool {
		_ = doClose(GrpcConn.conn)
	}
	//set pool nil
	p.connPool = nil
	return
}

//defaultFactoryCreateConn function to create grpc connections
func defaultFactoryCreateConn() ConnFactoryFunc {
	return func(p *GRPCPool) (*grpc.ClientConn, error) {
		opt := p.options
		dialOptions := p.dialOptions
		ctx, cancel := context.WithTimeout(context.Background(), opt.DialTimeout)
		defer cancel()
		target := opt.getTarget()
		if target == "" {
			return nil, ErrTargetEmpty
		}
		return grpc.DialContext(ctx, target, dialOptions...)
	}
}

//defaultCloseConn function to close grpc connections
func defaultCloseConn() ConnCloseFunc {
	return func(conn *grpc.ClientConn) error {
		return conn.Close()
	}
}

//NewGRPCPool returns a GRPCPool with options and initialization
func NewGRPCPool(opt *Options, dialOptions ...grpc.DialOption) (*GRPCPool, error) {
	if err := opt.validate(); err != nil {
		return nil, err
	}
	if opt.ClientKeepAlive {
		//keepalive config
		kacp := keepalive.ClientParameters{
			Time:                opt.IdleTimeout,
			Timeout:             opt.PingTimeout,
			PermitWithoutStream: opt.ForcePermit,
		}
		keepaliveOption := grpc.WithKeepaliveParams(kacp)
		dialOptions = append(dialOptions, keepaliveOption)
	}

	//pool
	pool := &GRPCPool{}
	pool.options = opt
	pool.dialOptions = dialOptions
	pool.connFactory = defaultFactoryCreateConn()
	pool.connDoClose = defaultCloseConn()
	pool.connPool = make([]*GrpcConn, 0, opt.Cap)
	pool.connNext = 0

	return pool, nil
}
