package app

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/fpawel/elco.v2/internal/view"
	"github.com/hako/durafmt"
	"github.com/powerman/structlog"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

func Run() {
	MainWindow = view.NewAppMainWindow(doWork, []view.NamedWork{
		{"Опрос", interrogate},
		{"Задержка", func() error {
			return delay("Задержка", 2*time.Minute)
		}},
	})
	MainWindow.Run()
}

func interrogate() error {
	ctxWork := MainWindow.Ctx(view.CtxWork)
	for {
		checkedBlocks := data.GetCheckedBlocks()
		if len(checkedBlocks) == 0 {
			return merry.New("необходимо выбрать блок для опроса")
		}
		for _, block := range checkedBlocks {
			if _, err := readBlock(log, block, ctxWork); err != nil {
				return err
			}
			pause(time.Second, ctxWork.Done())
		}
	}
}

func pause(duration time.Duration, done <-chan struct{}) {
	timer := time.NewTimer(duration)
	for {
		select {
		case <-timer.C:
			return
		case <-done:
			timer.Stop()
			return
		}
	}
}

func readBlock(log *structlog.Logger, block int, ctx context.Context) ([]float64, error) {

	log = comm.LogWithKeys(log, "блок", block)

	values, err := modbus.Read3BCDs(log, readerMeasure(ctx), modbus.Addr(block+101), 0, 8)
	if err != nil {
		return nil, merry.Appendf(err, "блок %d", block)
	}
	MainWindow.SetInterrogateBlockValues(block, values)
	return values, nil
}

func delay(what string, duration time.Duration) error {
	MainWindow.RunDelay(what, duration)

	origLog := log
	defer func() {
		log = origLog
		MainWindow.SkipDelay()
	}()
	log = comm.LogWithKeys(log,
		"фоновый_опрос", what,
		"total_delay_duration", durafmt.Parse(duration),
	)

	ctxDelay := MainWindow.Ctx(view.CtxDelay)
	for {
		products := data.GetLastPartyProducts(data.WithProduction)

		if len(products) == 0 {
			return merry.New("фоновый опрос: не выбрано ни одного прибора")
		}
		for _, products := range groupProductsByBlocks(products) {
			block := products[0].Place / 8
			_, err := readBlock(log, block, ctxDelay)
			if ctxDelay.Err() != nil {
				return nil
			}
			if err != nil {
				return err
			}
			pause(time.Second, ctxDelay.Done())
		}
	}
}

func groupProductsByBlocks(ps []data.Product) (gs [][]*data.Product) {
	m := make(map[int][]*data.Product)
	for i := range ps {
		p := &ps[i]
		v, _ := m[p.Place/8]
		m[p.Place/8] = append(v, p)
	}
	for _, v := range m {
		gs = append(gs, v)
	}
	sort.Slice(gs, func(i, j int) bool {
		return gs[i][0].Place/8 < gs[j][0].Place
	})
	return
}

func doWork(w view.NamedWork) error {
	log = comm.NewLogWithKeys("работа", w.Name)
	defer func() {
		log.ErrIfFail(comportGas.Close)
		log.ErrIfFail(comportMeasure.Close)
		log = structlog.New()
	}()
	err := w.Work()
	if err != nil && !merry.Is(err, context.Canceled) {
		printWorkErr(err)
		return err
	}
	return err
}

func printWorkErr(err error) {
	var kvs []interface{}
	for k, v := range merry.Values(err) {
		strK := fmt.Sprintf("%v", k)
		if strK != "stack" && strK != "msg" && strK != "message" {
			kvs = append(kvs, k, v)
		}
	}
	kvs = append(kvs, "stack", merryStacktrace(err))
	log.PrintErr(err, kvs...)
}

// stacktrace returns the error's stacktrace as a string formatted
// the same way as golangs runtime package.
// If e has no stacktrace, returns an empty string.
func merryStacktrace(e error) string {

	s := merry.Stack(e)
	if len(s) > 0 {
		buf := bytes.Buffer{}
		for i, fp := range s {
			fnc := runtime.FuncForPC(fp)
			if fnc != nil {
				f, l := fnc.FileLine(fp)
				name := filepath.Base(fnc.Name())
				ident := " "
				if i > 0 {
					ident = "\t"
				}

				buf.WriteString(fmt.Sprintf("%s%s:%d %s\n", ident, f, l, name))
			}
		}
		return buf.String()
	}
	return ""
}

var (
	log = structlog.New()

	MainWindow *view.AppWindow
)
