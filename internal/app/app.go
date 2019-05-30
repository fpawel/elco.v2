package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/fpawel/elco.v2/internal/view"
	"github.com/powerman/structlog"
	"time"
)

func Run() {
	MainWindow = view.NewAppMainWindow(doWork, []view.Work{
		{"Опрос", interrogate},
	})

	MainWindow.Run()
}

func interrogate() error {
	for {
		checkedBlocks := data.GetCheckedBlocks()
		if len(checkedBlocks) == 0 {
			return merry.New("необходимо выбрать блок для опроса")
		}
		for _, block := range checkedBlocks {

			if _, err := readBlockMeasure(log, block, MainWindow.Ctx); err != nil {
				return err
			}
		}
	}
}

func readBlockMeasure(log *structlog.Logger, block int, ctx context.Context) ([]float64, error) {

	log = comm.LogWithKeys(log, "блок", block)

	values, err := modbus.Read3BCDs(log, portMeasure, modbus.Addr(block+101), 0, 8)

	switch err {

	case nil:
		//notify.ReadCurrent(x.notifyWindow, api.ReadCurrent{
		//	Block:  block,
		//	Values: values,
		//})
		return values, nil

	default:
		return nil, merry.WithValue(err, "block", block)
	}
}

func doWork(w view.Work) error {
	log = comm.NewLogWithKeys("работа", w.Name)
	log.ErrIfFail(portMeasure.Close)
	log.ErrIfFail(portGas.Close)
	return w.Func()
}

var (
	log = structlog.New()

	portMeasure = reader{
		Reader: comport.NewReader(comport.Config{
			Baud:        115200,
			ReadTimeout: time.Millisecond,
		}),
		Config: comm.Config{
			ReadByteTimeoutMillis: 15,
			ReadTimeoutMillis:     500,
			MaxAttemptsRead:       10,
		},
	}

	portGas = reader{
		Reader: comport.NewReader(comport.Config{
			Baud:        9600,
			ReadTimeout: time.Millisecond,
		}),
		Config: comm.Config{
			ReadByteTimeoutMillis: 50,
			ReadTimeoutMillis:     1000,
			MaxAttemptsRead:       3,
		},
	}

	MainWindow *view.AppWindow
)
