package view

import (
	"context"
	"fmt"
	"github.com/hako/durafmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/powerman/structlog"
	"sync"
	"time"
)

type delayHelp struct {
	*walk.Composite
	pb     *walk.ProgressBar
	lbl    *walk.Label
	cancel context.CancelFunc
	ctx    context.Context
}

func (x *delayHelp) start(what string, total time.Duration, ctx context.Context) {

	log := log.New("delay", what, "total_delay_duration", durafmt.Parse(total))
	log.Info("begin", structlog.KeyTime, now())

	var wg sync.WaitGroup
	wg.Add(1)
	x.Composite.Synchronize(func() {
		x.ctx, x.cancel = context.WithTimeout(ctx, total)
		x.Composite.SetVisible(true)
		x.pb.SetRange(0, int(total.Nanoseconds()/1000000))
		x.pb.SetValue(0)
		s := fmt.Sprintf("%s: %s", what, durafmt.Parse(total))
		if err := x.lbl.SetText(s); err != nil {
			panic(err)
		}
		wg.Done()
	})
	wg.Wait()

	go func() {
		startMoment := time.Now()
		ticker := time.NewTicker(time.Millisecond * 500)
		defer func() {
			ticker.Stop()
			x.Composite.Synchronize(func() {
				x.SetVisible(false)
			})
			log.Info("end", structlog.KeyTime, now())
		}()
		for {
			select {
			case <-ticker.C:
				x.Composite.Synchronize(func() {
					x.pb.SetValue(int(time.Since(startMoment).Nanoseconds() / 1000000))
				})
			case <-x.ctx.Done():
				return
			}
		}
	}()
}

func (x *delayHelp) Widget() Widget {
	return Composite{
		AssignTo: &x.Composite,
		Layout:   HBox{},
		Visible:  false,
		Children: []Widget{
			Label{AssignTo: &x.lbl},
			ScrollView{
				Layout:        VBox{SpacingZero: true, MarginsZero: true},
				VerticalFixed: true,
				Children: []Widget{
					ProgressBar{
						AssignTo: &x.pb,
						MaxSize:  Size{0, 15},
						MinSize:  Size{0, 15},
					},
				},
			},

			PushButton{
				Text: "Продолжить без задержки",
				OnClicked: func() {
					x.cancel()
					log.Warn("задержка прервана", structlog.KeyTime, now())
				},
			},
		},
	}
}

func now() string {
	return time.Now().Format("15:04:05")
}
