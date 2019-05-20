package main

import (
	"fmt"
	"github.com/hako/durafmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/powerman/structlog"
	"time"
)

type delayHelp struct {
	*walk.Composite
	pb     *walk.ProgressBar
	lbl    *walk.Label
	ticker *time.Ticker
	done   chan struct{}
}

func (x *delayHelp) Show(what string, total time.Duration) {

	log := structlog.New("delay", what)

	startMoment := time.Now()
	x.done = make(chan struct{}, 1)
	x.ticker = time.NewTicker(time.Millisecond * 500)
	x.Composite.Synchronize(func() {
		x.SetVisible(true)
		x.pb.SetRange(0, int(total.Nanoseconds()/1000000))
		x.pb.SetValue(0)
		_ = x.lbl.SetText(fmt.Sprintf("%s: %s", what, durafmt.Parse(total)))
	})

	log.Info("begin", structlog.KeyTime, now())
	go func() {
		defer func() {
			log.Info("end", structlog.KeyTime, now())
		}()
		for {
			select {
			case <-x.ticker.C:
				x.Composite.Synchronize(func() {
					x.pb.SetValue(int(time.Since(startMoment).Nanoseconds() / 1000000))
				})
			case <-x.done:
				return
			}
		}
	}()
}

func (x *delayHelp) Hide() {
	x.ticker.Stop()
	close(x.done)
	x.Composite.Synchronize(func() {
		x.SetVisible(false)
	})

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
					skipDelay()
				},
			},
		},
	}
}

func now() string {
	return time.Now().Format("15:04:05")
}