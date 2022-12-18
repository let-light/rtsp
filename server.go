package rtsp

import (
	"context"
	"fmt"

	"github.com/panjf2000/gnet"
)

type IServEventHandler interface {
	ISSEventHandler
	OnShutdown(s *Server)
	OnConnect(ss IServSession)
	OnDisconnect(ss IServSession)
}

type Options struct {
	gnet.Options
	Addr string `mapstructure:"addr"`
}

type Server struct {
	gnet.EventServer
	handler IServEventHandler
	opt     Options
}

func NewServer(handler IServEventHandler, opt Options) (*Server, error) {
	s := &Server{
		handler: handler,
		opt:     opt,
	}

	s.opt.Options.Codec = s
	if opt.Logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	return s, nil
}

func (s *Server) Run() error {
	err := gnet.Serve(s, s.opt.Addr, gnet.WithOptions(s.opt.Options))
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return gnet.Stop(ctx, s.opt.Addr)
}

func (s *Server) OnInitComplete(gs gnet.Server) (action gnet.Action) {
	return
}

func (s *Server) OnShutdown(gs gnet.Server) {
	if s.handler != nil {
		s.handler.OnShutdown(s)
	}
}

func (s *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	ss := NewServSession(s.handler, c, s.opt.Options.Logger)

	if s.handler != nil {
		s.handler.OnConnect(ss)
	}

	return
}

func (s *Server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	if s.handler != nil {
		ss, err := s.getServConnection(c)
		if err == nil {
			s.handler.OnDisconnect(ss)
		}
	}

	return
}

func (s *Server) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

func (s *Server) Decode(c gnet.Conn) ([]byte, error) {
	ss, err := s.getServConnection(c)
	if err != nil {
		s.opt.Logger.Errorf("getServConnection error: %v", err)
		return nil, err
	}

	offset, err := ss.Serv.Feed(c.Read())
	if offset > 0 && offset < c.BufferLength() {
		c.ShiftN(offset)
	} else if offset >= c.BufferLength() {
		c.ResetBuffer()
	}

	if err != nil {
		s.opt.Logger.Errorf("serv feed error: %v", err)
		return nil, err
	}

	return nil, nil
}

func (s *Server) getServConnection(c gnet.Conn) (*ServSession, error) {
	cctx := c.Context()
	if cctx == nil {
		s.opt.Logger.Errorf("connection context is nil")
		return nil, fmt.Errorf("connection context is nil")
	}

	conn := cctx.(*ServSession)

	return conn, nil
}
