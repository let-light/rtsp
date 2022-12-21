package rtsp

import "sync"

type IServSessionEventListener interface {
	OnPlay(ss *IServSession) error
	OnPublish(ss *IServSession) error
	OnPause(ss *IServSession) error
	OnResume(ss *IServSession) error
	OnStream(ss *IServSession) error
}

type IServSession interface {
	AddParams(k, v interface{})
	GetParams(k interface{}) (interface{}, bool)
	DeleteParams(k interface{})
	EvehtHandler() IServSessionEventListener
}

type ServSession struct {
	params   sync.Map
	listener IServSessionEventListener
}

func NewServSession(listener IServSessionEventListener) *ServSession {
	return &ServSession{
		listener: listener,
	}
}

func (ss *ServSession) AddParams(k, v interface{}) {
	ss.params.Store(k, v)
}

func (ss *ServSession) GetParams(k interface{}) (interface{}, bool) {
	return ss.params.Load(k)
}

func (ss *ServSession) DeleteParams(k interface{}) {
	ss.params.Delete(k)
}

func (ss *ServSession) EvehtHandler() IServSessionEventListener {
	return ss.listener
}
