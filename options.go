package pool

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

var (
	//ErrPoolClosed error when pool is closed
	ErrPoolClosed = errors.New("pool is closed")
	//ErrOptionValid error when pool option is invalid
	ErrOptionValid = errors.New("options is invalid")
	//ErrTargetEmpty error when pool target address is empty
	ErrTargetEmpty = errors.New("target is empty")
	//ErrPoolInitialized error when pool is not initialized
	ErrPoolInitialized = errors.New("pool has been initialized")
	//ErrClientBuilderNil error when no client builder set
	ErrClientBuilderNil = errors.New("servercluster client builder is nil")
	//ErrConnConnect error when connection failed
	ErrConnConnect = errors.New("failed to get connection too many times")
	//ErrServerBuilderNil error when server build is nil
	ErrServerBuilderNil = errors.New("server client builder is nil")
)

//Options is for GRPCPool
type Options struct {
	Cap             int
	Targets         []string
	ClientKeepAlive bool
	DialTimeout     time.Duration
	IdleTimeout     time.Duration
	PingTimeout     time.Duration
	ForcePermit     bool
}

//validate option if available
func (o Options) validate() error {
	if o.Targets == nil ||
		o.Cap <= 0 ||
		o.DialTimeout == 0 {
		return ErrOptionValid
	}

	if o.ClientKeepAlive && (o.PingTimeout == 0 ||
		o.IdleTimeout == 0) {
		return ErrOptionValid
	}

	return nil
}

//getTarget return a rand target from Options.Targets
func (o Options) getTarget() string {
	l := len(o.Targets)
	if l == 0 {
		return ""
	}
	if l == 1 {
		return o.Targets[0]
	}
	//随机获取目标地址
	return o.Targets[rand.Int()%l]
}

//NewOptions  return a *Options instance
//the timeouts is set 5s and ForcePermit is set false by default
func NewOptions(cap int, targets []string) (*Options, error) {
	opt := &Options{
		Cap:             cap,
		Targets:         targets,
		ClientKeepAlive: true,
		DialTimeout:     5 * time.Second,
		IdleTimeout:     5 * time.Second,
		PingTimeout:     5 * time.Second,
		ForcePermit:     false,
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return opt, nil
}
