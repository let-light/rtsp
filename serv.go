package rtsp

import (
	"strconv"
	"time"

	goPool "github.com/panjf2000/gnet/pkg/pool/goroutine"
)

const (
	EmptyState State = iota
	OptionsState
	DescribeState
	SetupState
	PlayState
	PauseState
	TeardownState
)

type ServOptions struct {
	IdleTimeout time.Duration `json:"idleTimeout,omitempty" p:"idleTimeout"` // idle timeout
	Logger      Logger
	Write       WriteHandler
}

type Serv struct {
	ss          IServSession
	state       State
	cseqCounter int
	pool        *goPool.Pool
	descChan    chan string
	url         string
	options     ServOptions
}

func NewServ(ss IServSession, options ServOptions) *Serv {

	return &Serv{
		ss:          ss,
		state:       EmptyState,
		cseqCounter: 0,
		pool:        goPool.Default(),
		descChan:    make(chan string, 1),
		url:         "",
		options:     options,
	}
}

func (serv *Serv) State() State {
	return serv.state
}

func (serv *Serv) decodeRtpRtcp(buf []byte) (int, error) {
	return 0, nil
}

func (serv *Serv) Feed(buf []byte) (int, error) {
	if len(buf) == 0 {
		serv.options.Logger.Warnf("rtsp feed empty data")
		return 0, nil
	}

	var err error
	var endOffset int
	if buf[0] != '$' {
		var req *Request
		req, endOffset, err = UnmarshalRequest(buf)
		if err != nil {
			return endOffset, err
		}

		err := serv.handleRequest(req)
		if err != nil {
			return endOffset, err
		}

	} else {
		endOffset, err = serv.decodeRtpRtcp(buf)
		if err != nil {
			return endOffset, err
		}

	}

	return endOffset, nil
}

func (serv *Serv) SetDescribe(desc string) {
	serv.descChan <- desc
}

func (serv *Serv) handleRequest(req *Request) error {
	defer func() {
		if err := recover(); err != nil {
			serv.options.Logger.Errorf("handleRequest panic: %v", err)
		}
	}()

	serv.pool.Submit(func() {
		serv.options.Logger.Debugf("rtsp request: %s", req.String())

		if serv.url == "" {
			serv.url = req.Url()
		}

		var err error
		switch req.Method() {
		case OptionsMethod:
			err = serv.OptionsProcess(req)
		case DescribeMethod:
			err = serv.DescribeProcess(req)
		case AnnounceMethod:
			err = serv.AnnounceProcess(req)
		case SetupMethod:
			err = serv.SetupProcess(req)
		case PlayMethod:
			err = serv.PlayProcess(req)
		case PauseMethod:
			err = serv.PauseProcess(req)
		case TeardownMethod:
			err = serv.TeardownProcess(req)
		case GetParameterMethod:
			err = serv.GetParameterProcess(req)
		case SetParameterMethod:
			err = serv.SetParameterProcess(req)
		default:
			err = serv.WriteResponseStatus(req.CSeq(), StatusMethodNotAllowed)
		}

		if err != nil {
			serv.options.Logger.Errorf("rtsp request error: %s", err.Error())
			return
		}
	})
	return nil
}

func (serv *Serv) OptionsProcess(req *Request) error {
	resp := serv.NewResponse(req.CSeq(), StatusOK).Option()
	resp.SetOptions([]string{
		"OPTIONS",
		"ANNOUNCE",
		"DESCRIBE",
		"SETUP",
		"TEARDOWN",
		"PLAY",
		"PAUSE",
		"GET_PARAMETER",
		"SET_PARAMETER",
	})

	return serv.WriteResponse(resp)
}

func (serv *Serv) DescribeProcess(req *Request) error {
	if serv.ss.GetEventListener() != nil {
		if err := serv.ss.GetEventListener().OnDescribe(serv); err != nil {
			serv.options.Logger.Errorf("rtsp describe error: %s", err.Error())
			return serv.WriteResponseStatus(req.CSeq(), StatusForbidden)
		}
	}

	select {
	case desc := <-serv.descChan:
		serv.options.Logger.Debugf("rtsp describe get desc: %s", desc)
		resp := serv.NewResponse(req.CSeq(), StatusOK).Describe()
		resp.SetContentType("application/sdp")
		resp.SetContentBase(serv.url)
		resp.SetContent(desc)
		return serv.WriteResponse(resp)

	case <-time.After(serv.options.IdleTimeout * time.Second):
		serv.options.Logger.Debugf("rtsp describe timeout")
		return serv.WriteResponseStatus(req.CSeq(), StatusNotFound)
	}
}

func (serv *Serv) AnnounceProcess(req *Request) error {
	contentType := req.Announce().ContentType()
	if contentType != "application/sdp" {
		return serv.WriteResponseStatus(req.CSeq(), StatusUnsupportedMediaType)
	}

	if serv.ss.GetEventListener() != nil {
		if err := serv.ss.GetEventListener().OnAnnounce(serv); err != nil {
			serv.options.Logger.Errorf("rtsp announce error: %s", err.Error())
			return serv.WriteResponseStatus(req.CSeq(), StatusForbidden)
		}
	}

	return serv.WriteResponse(serv.NewResponse(req.CSeq(), StatusOK))
}

func (serv *Serv) SetupProcess(req *Request) error {
	serv.options.Logger.Debugf("rtsp setup")
	return nil
}

func (serv *Serv) PlayProcess(req *Request) error {
	return nil
}

func (serv *Serv) PauseProcess(req *Request) error {
	return nil
}

func (serv *Serv) TeardownProcess(req *Request) error {
	return nil
}

func (serv *Serv) GetParameterProcess(req *Request) error {
	return nil
}

func (serv *Serv) SetParameterProcess(req *Request) error {
	return nil
}

func (serv *Serv) NewResponse(cseq int, status Status) *Response {
	resp := &Response{
		version:   "RTSP/1.0",
		status:    status,
		statusStr: status.String(),
		lines: HeaderLines{
			"CSeq":           strconv.Itoa(cseq),
			"Date":           time.Now().Format(time.RFC1123),
			"Content-Length": strconv.Itoa(0),
			"Server":         "Neon-RTSP",
		},
	}

	return resp
}

func (serv *Serv) WriteResponse(resp IResponse) error {
	return serv.options.Write([]byte(resp.String()))
}

func (serv *Serv) WriteResponseStatus(cseq int, status Status) error {
	return serv.WriteResponse(serv.NewResponse(cseq, status))
}
