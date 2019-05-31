package app

import (
	"context"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco.v2/internal/cfg"
	"github.com/fpawel/elco.v2/internal/view"
	"github.com/lxn/walk"
	"github.com/powerman/structlog"
	"time"
)

type reader struct {
	comport     *comport.Reader
	config      comm.Config
	portNameKey string
	ctx         context.Context
}

func readerMeasure(ctx context.Context) reader {
	return reader{
		ctx:         ctx,
		portNameKey: cfg.ComportKey,
		comport:     comportMeasure,
		config: comm.Config{
			ReadByteTimeoutMillis: 15,
			ReadTimeoutMillis:     500,
			MaxAttemptsRead:       10,
		},
	}
}

func readerGas() reader {
	return reader{
		ctx:         MainWindow.Ctx(view.CtxWork),
		portNameKey: cfg.ComportGasKey,
		comport:     comportGas,
		config: comm.Config{
			ReadByteTimeoutMillis: 50,
			ReadTimeoutMillis:     1000,
			MaxAttemptsRead:       3,
		},
	}
}

var (
	comportMeasure = comport.NewReader(comport.Config{
		Baud:        115200,
		ReadTimeout: time.Millisecond,
	})
	comportGas = comport.NewReader(comport.Config{
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	})
)

func (x reader) GetResponse(logger *structlog.Logger, bytes []byte, responseParser comm.ResponseParser) ([]byte, error) {
	if !x.comport.Opened() {
		portName, _ := walk.App().Settings().Get(x.portNameKey)
		if err := x.comport.Open(portName); err != nil {
			return nil, err
		}
	}
	return x.comport.GetResponse(comm.Request{
		Logger:         logger,
		Bytes:          bytes,
		Config:         x.config,
		ResponseParser: responseParser,
	}, x.ctx)
}
