package view

import (
	"context"
	"fmt"
	"github.com/hako/durafmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/powerman/structlog"
	"time"
)

type delayHelp struct {
	*walk.Composite
	pb   *walk.ProgressBar
	lbl  *walk.Label
	skip context.CancelFunc
}

func (x *delayHelp) show(what string, total time.Duration) {

	x.Composite.SetVisible(true)
	x.pb.SetRange(0, int(total.Nanoseconds()/1000000))
	x.pb.SetValue(0)
	s := fmt.Sprintf("%s: %s", what, durafmt.Parse(total))
	if err := x.lbl.SetText(s); err != nil {
		panic(err)
	}
}

func (x *delayHelp) run(done <-chan struct{}) {
	startMoment := time.Now()
	ticker := time.NewTicker(time.Millisecond * 500)
	defer func() {
		ticker.Stop()
		x.Composite.Synchronize(func() {
			x.SetVisible(false)
		})
	}()
	for {
		select {
		case <-ticker.C:
			x.Composite.Synchronize(func() {
				x.pb.SetValue(int(time.Since(startMoment).Nanoseconds() / 1000000))
			})
		case <-done:
			return
		}
	}
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
					x.skip()
					log.Warn("задержка прервана", structlog.KeyTime, now())
				},
			},
		},
	}
}

func now() string {
	return time.Now().Format("15:04:05")
}
