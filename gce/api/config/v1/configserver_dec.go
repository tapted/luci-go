// Code generated by svcdec; DO NOT EDIT

package config

import (
	"context"

	proto "github.com/golang/protobuf/proto"

	empty "github.com/golang/protobuf/ptypes/empty"
)

type DecoratedConfig struct {
	// Service is the service to decorate.
	Service ConfigServer
	// Prelude is called for each method before forwarding the call to Service.
	// If Prelude returns an error, then the call is skipped and the error is
	// processed via the Postlude (if one is defined), or it is returned directly.
	Prelude func(c context.Context, methodName string, req proto.Message) (context.Context, error)
	// Postlude is called for each method after Service has processed the call, or
	// after the Prelude has returned an error. This takes the the Service's
	// response proto (which may be nil) and/or any error. The decorated
	// service will return the response (possibly mutated) and error that Postlude
	// returns.
	Postlude func(c context.Context, methodName string, rsp proto.Message, err error) error
}

func (s *DecoratedConfig) DeleteVMs(c context.Context, req *DeleteVMsRequest) (rsp *empty.Empty, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(c, "DeleteVMs", req)
		if err == nil {
			c = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteVMs(c, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(c, "DeleteVMs", rsp, err)
	}
	return
}

func (s *DecoratedConfig) EnsureVMs(c context.Context, req *EnsureVMsRequest) (rsp *Block, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(c, "EnsureVMs", req)
		if err == nil {
			c = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.EnsureVMs(c, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(c, "EnsureVMs", rsp, err)
	}
	return
}

func (s *DecoratedConfig) GetVMs(c context.Context, req *GetVMsRequest) (rsp *Block, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(c, "GetVMs", req)
		if err == nil {
			c = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetVMs(c, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(c, "GetVMs", rsp, err)
	}
	return
}