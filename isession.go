package rtsp

type ISSEventHandler interface {
	OnPlay(ss *IServSession) error
	OnPublish(ss *IServSession) error
	OnPause(ss *IServSession) error
	OnResume(ss *IServSession) error
	OnStream(ss *IServSession) error
}

type IServSession interface {
	AddContext(k, v interface{})
	GetContext(k interface{}) (interface{}, bool)
	DeleteContext(k interface{})
	EvehtHandler() ISSEventHandler
}
