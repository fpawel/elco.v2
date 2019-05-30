package view

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"sync"
	"time"
)

type AppWindow struct {
	w *walk.MainWindow
	tblProducts,
	tblJournal *walk.TableView
	lblWork,
	lblWorkTime *walk.Label
	delayHelp      *delayHelp
	productsTblMdl *ProductsTable
	blocksTblMdl   *BlocksTable
	journal        *Journal
	cancelWork     context.CancelFunc

	DoMainWork func(mainWorkIndex int) error

	enableOnWork, visibleOnWork []visWidget

	workStarted bool

	tbStop,
	tbStart, tbNewParty *walk.ToolButton
	cbWorks    *walk.ComboBox
	panelTools *walk.ScrollView

	ctx    context.Context
	works  []Work
	doWork func(Work) error
}

type Work struct {
	Name string
	Func func() error
}

type visWidget struct {
	*walk.WindowBase
	V bool
}

func NewAppMainWindow(doWork func(Work) error, works []Work) *AppWindow {
	x := &AppWindow{
		delayHelp:  new(delayHelp),
		cancelWork: func() {},
		DoMainWork: func(mainWorkIndex int) error {
			time.Sleep(2 * time.Second)
			return errors.New("not implemented")
		},
		journal: new(Journal),
		works:   works,
		doWork:  doWork,
	}
	x.productsTblMdl, x.blocksTblMdl = newProductsModels()
	return x
}

func (x *AppWindow) showErr(title, text string) {
	walk.MsgBox(x.w, title,
		text, walk.MsgBoxIconError|walk.MsgBoxOK)
}

func (x *AppWindow) AddJournalRecord(logLevel LogLevel, text string) {
	x.journal.entries = append(x.journal.entries, JournalEntry{
		time.Now(),
		text,
		logLevel,
	})
	x.journal.PublishRowsReset()
	x.tblJournal.EnsureItemVisible(len(x.journal.entries) - 1)
}

func (x *AppWindow) Ctx() (ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	x.w.Synchronize(func() {
		ctx = x.ctx
		if x.delayHelp.Composite.Visible() {
			ctx = x.delayHelp.ctx
		}
		wg.Done()
	})
	return
}

func (x *AppWindow) StartDelay(what string, duration time.Duration) {

}

func (x *AppWindow) window() MainWindow {

	window := MainWindow{
		AssignTo: &x.w,
		Title: "Партия ЭХЯ " + (func() string {
			p := data.GetLastParty(data.WithoutProducts)
			return fmt.Sprintf("№%d %s", p.PartyID, p.CreatedAt.Format("02.01.2006"))
		}()),
		Name:       "MainWindow",
		Font:       Font{PointSize: 12, Family: "Segoe UI"},
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
		Size:       Size{800, 600},
		Layout:     VBox{MarginsZero: true, SpacingZero: true},

		Children: []Widget{
			ScrollView{
				AssignTo:      &x.panelTools,
				VerticalFixed: true,
				Layout:        HBox{SpacingZero: true},
				Children: []Widget{
					ToolButton{
						AssignTo:    &x.tbNewParty,
						Text:        "Новая загрузка",
						Image:       "img/new25.png",
						ToolTipText: "Создать новую загрузку",
						OnClicked: func() {
							if walk.MsgBox(x.w, "Новая партия",
								"Подтвердите необходимость создания новой партии",
								walk.MsgBoxIconQuestion|walk.MsgBoxYesNo) != win.IDYES {
								return
							}
							data.CreateNewParty()
							x.resetProductsView()
						},
					},

					ToolButton{
						Text:        "Выбрать годные ЭХЯ",
						Image:       "img/check25m.png",
						ToolTipText: "Выбрать годные ЭХЯ",
						OnClicked: func() {
							data.SetOnlyOkProductsProduction()
							x.resetProductsView()
						},
					},

					ToolButton{
						Text:        "Паспорта и итоговая таблица",
						Image:       "img/pdf25.png",
						ToolTipText: "Паспорта и итоговая таблица",
						OnClicked: func() {

						},
					},

					VSpacer{MinSize: Size{10, 0}},

					ToolButton{
						AssignTo:    &x.tbStart,
						Text:        "Начать выполнение выбранной операции",
						Image:       "img/start25.png",
						ToolTipText: "Начать выполнение выбранной операции",
						OnClicked: func() {

							for _, y := range x.enableOnWork {
								y.WindowBase.SetEnabled(y.V)
							}
							for _, y := range x.visibleOnWork {
								y.WindowBase.SetVisible(y.V)
							}

							if x.workStarted {
								panic("already started")
							}
							x.workStarted = true
							x.ctx, x.cancelWork = context.WithCancel(context.Background())

							workIndex := x.cbWorks.CurrentIndex()
							workName := x.cbWorks.Model().([]string)[workIndex]

							x.AddJournalRecord(INF, fmt.Sprintf("%v: начало выполнения", workName))

							go func() {

								err := x.DoMainWork(workIndex)

								x.w.Synchronize(func() {

									x.workStarted = false

									for _, y := range x.enableOnWork {
										y.WindowBase.SetEnabled(!y.V)
									}
									for _, y := range x.visibleOnWork {
										y.WindowBase.SetVisible(!y.V)
									}

									if err != nil {
										if merry.Is(err, context.Canceled) {
											//dafMainWindow.SetWorkStatus(walk.RGB(139, 69, 19), what+": прервано")
											x.AddJournalRecord(WRN, fmt.Sprintf("%v: выполнение прервано", workName))
										} else {
											//dafMainWindow.SetWorkStatus(walk.RGB(255, 0, 0), what+": "+err.Error())
											//log.PrintErr(err)
											x.AddJournalRecord(ERR, fmt.Sprintf("%v: произошла ошибка: %v", workName, err))
											x.showErr(workName, err.Error())
										}

									} else {
										//dafMainWindow.SetWorkStatus(ColorNavy, workName+": выполнено")
									}

								})
							}()
						},
					},
					ToolButton{
						AssignTo:    &x.tbStop,
						Visible:     false,
						Text:        "Прервать выполнение операции",
						Image:       "img/stop25.png",
						ToolTipText: "Прервать выполнение операции",
						OnClicked: func() {
							x.cancelWork()
						},
					},
					VSpacer{MinSize: Size{3, 0}},
					ComboBox{
						AssignTo: &x.cbWorks,
						Model: []string{
							"Опрос",
							"Термокомпенсация",
							"Погрешность",
							"Прошивка",
						},
						CurrentIndex: 0,
					},

					Label{
						AssignTo:  &x.lblWorkTime,
						TextColor: walk.RGB(0, 128, 0),
					},
					Label{
						AssignTo: &x.lblWork,
					},
					x.delay.Widget(),
					ScrollView{
						VerticalFixed: true,
						Layout:        Grid{},
					},
					ToolButton{
						Text:        "Ввод серийных номеров ЭХЯ",
						Image:       "img/edit25.png",
						ToolTipText: "Ввод серийных номеров ЭХЯ",
						OnClicked: func() {

						},
					},
					ToolButton{
						Text:        "Настройки",
						Image:       "img/sets25b.png",
						ToolTipText: "Настройки",
						OnClicked:   x.runSettingsDialog,
					},
				},
			},
			ScrollView{
				Layout: HBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{

					GroupBox{
						Layout: Grid{},
						Title:  "Настраиваемые ЭХЯ",
						Children: []Widget{

							TableView{
								AssignTo:                 &x.tblProducts,
								NotSortableByHeaderClick: true,
								LastColumnStretched:      true,
								CheckBoxes:               true,
								MultiSelection:           true,
								Model:                    x.productsTblMdl,
								Columns:                  x.productsTblMdl.Columns(),
								OnItemActivated: func() {
									p := x.productsTblMdl.ProductAtPlace(x.tblProducts.CurrentIndex())
									if p.ProductID != 0 {
										runFirmwareDialog(x.w, p)
									}
								},
								OnKeyDown: func(key walk.Key) {
									switch key {

									case walk.KeyInsert:

									case walk.KeyDelete:

									}

								},
							},
						},
					},
					Composite{
						Layout: VBox{MarginsZero: true, SpacingZero: true},
						Children: []Widget{
							GroupBox{
								Layout: Grid{},
								Title:  "Опрос",
								Children: []Widget{
									TableView{
										Model:      x.blocksTblMdl,
										CheckBoxes: true,
										Columns: []TableViewColumn{
											{
												Title: "Блок",
											},
											{
												Title: "Место 1",
											},
											{
												Title: "Место 2",
											},
											{
												Title: "Место 3",
											},
											{
												Title: "Место 4",
											},
											{
												Title: "Место 5",
											},
											{
												Title: "Место 6",
											},
											{
												Title: "Место 7",
											},
											{
												Title: "Место 8",
											},
										},
									},
								},
							},
							GroupBox{
								Title:  "Журнал",
								Layout: Grid{},
								Children: []Widget{
									TableView{
										AssignTo: &x.tblJournal,
										Columns: []TableViewColumn{
											{
												Title: "Время",
											},
											{
												Title: "Сообщение",
											},
										},
										Model:               x.journal,
										LastColumnStretched: true,
										HeaderHidden:        true,
										ColumnsSizable:      false,
										ColumnsOrderable:    false,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return window
}

func (x *AppWindow) Run() {
	settings := walk.NewIniFileSettings("settings.ini")
	defer log.ErrIfFail(settings.Save)

	app := walk.App()
	app.SetOrganizationName("analitpribor")
	app.SetProductName("elco")
	app.SetSettings(settings)
	log.ErrIfFail(settings.Load)

	log.Debug("create main window")
	if err := x.window().Create(); err != nil {
		panic(err)
	}
	log.Debug("run main window")

	x.visibleOnWork = []visWidget{
		{&x.tbStart.WindowBase, false},
		{&x.tbStop.WindowBase, true},
	}
	x.enableOnWork = []visWidget{
		{&x.tbNewParty.WindowBase, false},
		{&x.cbWorks.WindowBase, false},
	}

	x.w.Run()
}

func (x *AppWindow) resetProductsView() {
	x.productsTblMdl.Reset(x.tblProducts)
}

var log = structlog.New()
