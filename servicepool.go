package pool

import (
	"sync"

	"google.golang.org/grpc"
)

//ServerBuilderFunc defined the func of builder
type ServerBuilderFunc func(conn grpc.ClientConnInterface) interface{}

//ServiceCenter service manage center
type ServiceCenter struct {
	clusters sync.Map
}

//Register to put a server cluster into center
func (sc *ServiceCenter) Register(server *ServerCluster) {
	sc.clusters.Store(server.Name, server)
}

//Get return the server with specific server name and a bool to ensure the server exist
func (sc *ServiceCenter) Get(clusterName string) (*ServerCluster, bool) {
	v, ok := sc.clusters.Load(clusterName)
	if !ok {
		return nil, false
	}
	return v.(*ServerCluster), true
}

//UnsafeGet return a server without ensure exist
// ** make sure the server exist before use this method **
func (sc *ServiceCenter) UnsafeGet(clusterName string) *ServerCluster {
	v, _ := sc.clusters.Load(clusterName)
	return v.(*ServerCluster)
}

//ServerCluster is a manager of one physical server
type ServerCluster struct {
	Name          string
	Pool          *GRPCPool
	clientBuilder map[string]ServerBuilderFunc
}

//GetClient return a *GrpcConn
//user can create custom server client with GrpcConn.Conn()
func (server *ServerCluster) GetClient() (*GrpcConn, error) {
	conn, err := server.Pool.Get()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//GetServerClient return grpc server client  release function and error
//client is a interface , it can be available after assert to your own type of client
// release is the GrpcConn.Release function  should be execute after all requests
func (server *ServerCluster) GetServerClient(servname string) (client interface{}, release func(), err error) {

	if len(server.clientBuilder) == 0 {
		return nil, nil, ErrClientBuilderNil
	}

	var builder ServerBuilderFunc
	var ok bool
	if builder, ok = server.clientBuilder[servname]; !ok {
		return nil, nil, ErrServerBuilderNil
	}
	if builder == nil {
		return nil, nil, ErrServerBuilderNil
	}

	conn, err := server.GetClient()
	if err != nil {
		return nil, nil, err
	}

	client = builder(conn.conn)
	release = conn.Release
	return
}

//SetClientBuilder set single server client builder
func (server *ServerCluster) SetClientBuilder(servname string, fn ServerBuilderFunc) {
	server.clientBuilder[servname] = fn
}

//SetClientBuilders set multi server client builders
//if servername exist new builder will instead
func (server *ServerCluster) SetClientBuilders(builders map[string]ServerBuilderFunc) {
	for serv, fn := range builders {
		server.SetClientBuilder(serv, fn)
	}
}

//NewServerCluster return a *ServerCluster
func NewServerCluster(serverName string, opt Options, dialOptions []grpc.DialOption) (*ServerCluster, error) {
	server := &ServerCluster{
		Name:          serverName,
		clientBuilder: make(map[string]ServerBuilderFunc),
	}
	gPool, err := NewGRPCPool(&opt, dialOptions...)
	if err != nil {
		return nil, err
	}
	server.Pool = gPool
	return server, nil
}

//NewServerClusterWithBuilders return a *ServerCluster with builders set
func NewServerClusterWithBuilders(serverName string, opt Options, dialOptions []grpc.DialOption, builders map[string]ServerBuilderFunc) (*ServerCluster, error) {
	server, err := NewServerCluster(serverName, opt, dialOptions)
	if err != nil {
		return nil, err
	}
	server.SetClientBuilders(builders)
	return server, nil
}
