package app

import (
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/lxn/walk"
	"github.com/powerman/structlog"
)

type reader struct {
	*comport.Reader
	comm.Config
	PortNameKey string
}

func (x reader) GetResponse(logger *structlog.Logger, bytes []byte, responseParser comm.ResponseParser) ([]byte, error) {
	if !x.Reader.Opened() {
		portName, _ := walk.App().Settings().Get(x.PortNameKey)
		if err := x.Reader.Open(portName); err != nil {
			return nil, err
		}
	}
	return x.Reader.GetResponse(comm.Request{
		Logger:         logger,
		Bytes:          bytes,
		Config:         x.Config,
		ResponseParser: responseParser,
	}, MainWindow.Ctx)
}
