package view

import (
	"context"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"time"
)

type delayHelp struct {
	placeholder                   *walk.ScrollView
	pb                            *walk.ProgressBar
	lblWhat, lblTotal, lblElapsed *walk.Label
	skip                          context.CancelFunc
}

func (x *delayHelp) show(what string, total time.Duration) {

	x.placeholder.SetVisible(true)
	x.pb.SetRange(0, int(total.Nanoseconds()/1000000))
	x.pb.SetValue(0)

	if err := x.lblWhat.SetText(what); err != nil {
		panic(err)
	}
	if err := x.lblTotal.SetText(fmtDuration(total)); err != nil {
		panic(err)
	}
	if err := x.lblElapsed.SetText("00:00:00"); err != nil {
		panic(err)
	}

}

func (x *delayHelp) run(done <-chan struct{}) {
	startMoment := time.Now()
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
		x.placeholder.Synchronize(func() {
			x.placeholder.SetVisible(false)
		})
	}()
	for {
		select {
		case <-ticker.C:
			x.placeholder.Synchronize(func() {
				x.pb.SetValue(int(time.Since(startMoment).Nanoseconds() / 1000000))
				if err := x.lblElapsed.SetText(fmtDuration(time.Since(startMoment))); err != nil {
					log.PrintErr(err)
					return
				}
			})
		case <-done:
			return
		}
	}
}

func (x *delayHelp) Widget() Widget {
	return ScrollView{
		Layout:        HBox{SpacingZero: true, MarginsZero: true},
		VerticalFixed: true,
		Children: []Widget{

			ScrollView{
				AssignTo:      &x.placeholder,
				Visible:       false,
				Layout:        HBox{Spacing: 10, Margins: Margins{Left: 10, Right: 2}},
				VerticalFixed: true,
				Children: []Widget{
					Label{
						AssignTo:  &x.lblWhat,
						TextColor: walk.RGB(0, 0, 128),
					},
					Label{
						AssignTo:  &x.lblElapsed,
						TextColor: walk.RGB(139, 0, 0),
					},
					Label{
						Text:      ":",
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
			},
		},
	}
}

func now() string {
	return time.Now().Format("15:04:05")
}
