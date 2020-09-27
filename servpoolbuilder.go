package pool

import (
	"google.golang.org/grpc"
)

//ServiceCenterBuilder builder for ServiceCenter
//help create Service Center
type ServiceCenterBuilder struct {
	defaultOptions     *Options
	defaultGrpcOptions []grpc.DialOption
	clusters           []*ServerCluster
}

//SetDefaultOptions set GRPC Pool Options
func (scb *ServiceCenterBuilder) SetDefaultOptions(opt *Options) {
	opt.Targets = []string{}
	scb.defaultOptions = opt
}

//SetDefaultGrpcDialOptions set grpc dial options
func (scb *ServiceCenterBuilder) SetDefaultGrpcDialOptions(opts []grpc.DialOption) {
	scb.defaultGrpcOptions = opts
}

//SetServer register Server Cluster to Center
func (scb *ServiceCenterBuilder) SetServer(name string, builders map[string]ServerBuilderFunc, opt Options, grpcOptions []grpc.DialOption) error {
	cluster, err := NewServerCluster(name, opt, grpcOptions)
	if err != nil {
		return err
	}
	cluster.SetClientBuilders(builders)
	scb.clusters = append(scb.clusters, cluster)
	return nil
}

//SetServerWithDefaultOptions set Server Cluster use default
func (scb *ServiceCenterBuilder) SetServerWithDefaultOptions(name string, builders map[string]ServerBuilderFunc, targets ...string) error {
	opt := Options{}
	opt.Cap = scb.defaultOptions.Cap
	opt.ForcePermit = scb.defaultOptions.ForcePermit
	opt.ClientKeepAlive = scb.defaultOptions.ClientKeepAlive
	opt.PingTimeout = scb.defaultOptions.PingTimeout
	opt.IdleTimeout = scb.defaultOptions.IdleTimeout
	opt.DialTimeout = scb.defaultOptions.DialTimeout
	opt.Targets = targets

	grpcOptions := make([]grpc.DialOption, 0)
	grpcOptions = append(grpcOptions, scb.defaultGrpcOptions...)

	cluster, err := NewServerCluster(name, opt, grpcOptions)
	if err != nil {
		return err
	}

	cluster.SetClientBuilders(builders)
	scb.clusters = append(scb.clusters, cluster)
	return nil
}

//Build return *ServiceCenter
func (scb *ServiceCenterBuilder) Build() *ServiceCenter {
	sc := &ServiceCenter{}
	for _, c := range scb.clusters {
		sc.Register(c)
	}
	return sc
}

//NewServiceCenterBuilder return
func NewServiceCenterBuilder() *ServiceCenterBuilder {
	scb := &ServiceCenterBuilder{}
	opt, _ := NewOptions(10, []string{""})
	opt.ForcePermit = true
	scb.SetDefaultOptions(opt)

	var dialOptions []grpc.DialOption
	dialOptions = append(dialOptions, grpc.WithInsecure())

	scb.SetDefaultGrpcDialOptions(dialOptions)
	return scb
}
