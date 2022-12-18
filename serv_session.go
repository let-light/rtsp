package rtsp

import (
	"sync"

	"github.com/panjf2000/gnet"
)

type ServSession struct {
	*Serv
	c       gnet.Conn
	handler ISSEventHandler
	ctxs    sync.Map
}

func NewServSession(handler ISSEventHandler, c gnet.Conn, logger Logger) *ServSession {
	ss := &ServSession{
		c:       c,
		handler: handler,
	}

	ss.Serv = NewServ(
		ss,
		logger,
		func(data []byte) error {
			return ss.c.AsyncWrite(data)
		})

	c.SetContext(ss)

	return ss
}

func (ss *ServSession) AddContext(k, v interface{}) {
	ss.ctxs.Store(k, v)
}

func (ss *ServSession) GetContext(k interface{}) (interface{}, bool) {
	return ss.ctxs.Load(k)
}

func (ss *ServSession) DeleteContext(k interface{}) {
	ss.ctxs.Delete(k)
}

func (ss *ServSession) EvehtHandler() ISSEventHandler {
	return ss.handler
}
