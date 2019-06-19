package view

import (
	"context"
	"github.com/fpawel/gohelp/helpstr"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"time"
)

type delayHelp struct {
	spacer      *walk.ScrollView
	placeholder *walk.Composite
	pb          *walk.ProgressBar
	lblWhat,
	lblTotal *walk.Label
	skip context.CancelFunc
}

func (x *delayHelp) show(what string, total time.Duration) {

	x.spacer.SetVisible(false)
	x.placeholder.SetVisible(true)
	x.pb.SetRange(0, int(total.Nanoseconds()/1000000))
	x.pb.SetValue(0)

	if err := x.lblWhat.SetText(what); err != nil {
		panic(err)
	}
	if err := x.lblTotal.SetText(helpstr.FormatDuration(total)); err != nil {
		panic(err)
	}

}

func (x *delayHelp) run(done <-chan struct{}) {
	startMoment := time.Now()
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
		x.placeholder.Parent().Synchronize(func() {
			x.placeholder.SetVisible(false)
			x.spacer.SetVisible(true)
		})
	}()
	for {
		select {
		case <-ticker.C:
			x.pb.Synchronize(func() {
				x.pb.SetValue(int(time.Since(startMoment).Nanoseconds() / 1000000))

			})
			//x.lblElapsed.Synchronize(func() {
			//	must.AbortIf(x.lblElapsed.SetText(fmtDuration(time.Since(startMoment))))
			//})
		case <-done:
			return
		}
	}
}

func (x *delayHelp) Widget() Widget {
	return Composite{
		AssignTo: &x.placeholder,
		Visible:  false,
		Layout:   HBox{Spacing: 10, Margins: Margins{Left: 10, Right: 2}},
		Children: []Widget{
			Label{
				AssignTo:  &x.lblWhat,
				TextColor: walk.RGB(0, 0, 128),
			},
			Label{
				AssignTo:  &x.lblTotal,
				TextColor: walk.RGB(0, 0, 128),
			},
			Composite{
				Layout: VBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{
					ProgressBar{
						AssignTo: &x.pb,
						MaxSize:  Size{0, 15},
						MinSize:  Size{0, 15},
					},
				},
			},

			ToolButton{
				Image:       "img/skip25.png",
				ToolTipText: "Продолжить без задержки",
				OnClicked: func() {
					log.Info("пользователь прервал задержку")
					x.skip()
				},
			},
		},
	}
}

func now() string {
	return time.Now().Format("15:04:05")
}
