package view

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/fpawel/gohelp"
	"github.com/hako/durafmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/powerman/must"
	"github.com/powerman/structlog"
	"strings"
	"sync"
	"time"
)

type AppWindow struct {
	w *walk.MainWindow
	tblProducts,
	tblJournal *walk.TableView
	gbBlocks *walk.GroupBox

	delayHelp      *delayHelp
	productsTblMdl *ProductsTable
	blocksTblMdl   *BlocksTable
	journal        *Journal
	cancelWork     context.CancelFunc

	enableOnWork, visibleOnWork []visWidget

	workStarted, closing bool

	tbStop,
	tbStart, tbNewParty *walk.ToolButton
	cbWorks *walk.ComboBox

	menuProductTypes *walk.Menu

	ctxWork, ctxDelay context.Context
	works             []NamedWork
	doWork            func(NamedWork) error
}

type NamedWork struct {
	Name string
	Work Work
}

type Work func() error

type CtxType int

const (
	CtxWork CtxType = iota
	CtxDelay
)

type visWidget struct {
	*walk.WindowBase
	V bool
}

func NewAppMainWindow(doWork func(NamedWork) error, works []NamedWork) *AppWindow {
	x := &AppWindow{
		delayHelp:  &delayHelp{skip: func() {}},
		cancelWork: func() {},
		journal:    new(Journal),
		works:      works,
		doWork:     doWork,
	}
	x.productsTblMdl, x.blocksTblMdl = newProductsModels()
	return x
}

func (x *AppWindow) showErr(title string, err error) {

	if merry.Is(err, context.Canceled) {
		log.Warn("выполнение прервано")
		return
	}

	walk.MsgBox(x.w, title,
		fmt.Sprintf("выполнение завершилось с ошибкой: \n\n%v", err), walk.MsgBoxIconError|walk.MsgBoxOK)
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

func (x *AppWindow) Ctx(ctx CtxType) context.Context {
	ch := make(chan context.Context)
	x.w.Synchronize(func() {
		switch ctx {
		case CtxWork:
			ch <- x.ctxWork
		case CtxDelay:
			ch <- x.ctxDelay
		default:
			panic(ctx)
		}
	})
	return <-ch
}

func (x *AppWindow) SynchronizeStrong(f func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	x.w.Synchronize(func() {
		f()
		wg.Done()
	})
	wg.Wait()
}

func (x *AppWindow) SkipDelay() {
	x.SynchronizeStrong(x.delayHelp.skip)
}

func (x *AppWindow) RunDelay(what string, duration time.Duration) {
	log := gohelp.LogWithKeys(log, "delay", what, "total_delay_duration", durafmt.Parse(duration))
	log.Info("begin", structlog.KeyTime, now())
	x.SynchronizeStrong(func() {
		x.ctxDelay, x.delayHelp.skip = context.WithTimeout(x.ctxWork, duration)
		x.delayHelp.show(what, duration)
		go func() {
			x.delayHelp.run(x.ctxDelay.Done())
			log.Debug("end")
		}()
	})
}

func (x *AppWindow) window() MainWindow {

	var works []string
	for _, y := range x.works {
		works = append(works, y.Name)
	}

	return MainWindow{
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

			Composite{
				Layout: HBox{Spacing: 5},
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
							x.ResetProductsView()
						},
					},

					ToolButton{
						Text:        "Выбрать годные ЭХЯ",
						Image:       "img/check25m.png",
						ToolTipText: "Выбрать годные ЭХЯ",
						OnClicked: func() {
							data.SetOnlyOkProductsProduction()
							x.ResetProductsView()
						},
					},

					ToolButton{
						Text:        "Паспорта и итоговая таблица",
						Image:       "img/pdf25.png",
						ToolTipText: "Паспорта и итоговая таблица",
						OnClicked: func() {

						},
					},

					ToolButton{
						AssignTo:    &x.tbStart,
						Text:        "Начать выполнение выбранной операции",
						Image:       "img/start25.png",
						ToolTipText: "Начать выполнение выбранной операции",
						OnClicked:   x.runMainWork,
					},
					ToolButton{
						AssignTo:    &x.tbStop,
						Visible:     false,
						Text:        "Прервать выполнение операции",
						Image:       "img/stop25.png",
						ToolTipText: "Прервать выполнение операции",
						OnClicked: func() {
							log.Info("Пользователь прервал выполнение работы")
							x.cancelWork()
						},
					},

					ComboBox{
						AssignTo:     &x.cbWorks,
						Model:        works,
						CurrentIndex: 0,
					},

					x.delayHelp.Widget(),

					ScrollView{
						AssignTo:      &x.delayHelp.spacer,
						Layout:        HBox{MarginsZero: true},
						MaxSize:       Size{0, 1},
						VerticalFixed: true,
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
			TabWidget{
				Pages: []TabPage{
					{
						Title:  "Партия",
						Layout: Grid{MarginsZero: true, SpacingZero: true},
						Children: []Widget{
							TableView{
								AssignTo:                 &x.tblProducts,
								NotSortableByHeaderClick: true,
								LastColumnStretched:      true,
								CheckBoxes:               true,
								MultiSelection:           true,
								Model:                    x.productsTblMdl,
								Columns:                  AllProductsTableColumns(),
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
								ContextMenuItems: []MenuItem{
									Action{
										Text: "Выбрать",
										OnTriggered: func() {
											for _, place := range x.tblProducts.SelectedIndexes() {
												place := place
												must.AbortIf(x.productsTblMdl.SetChecked(place, true))
												x.productsTblMdl.PublishRowChanged(place)
											}
										},
									},
									Action{
										Text: "Отменить",
										OnTriggered: func() {
											for _, place := range x.tblProducts.SelectedIndexes() {
												place := place
												must.AbortIf(x.productsTblMdl.SetChecked(place, false))
												x.productsTblMdl.PublishRowChanged(place)
											}
										},
									},
									Separator{},
									Menu{
										Text:     "Исполнение",
										AssignTo: &x.menuProductTypes,
									},
									Menu{
										Text: "Метод расчёта",
										Items: []MenuItem{
											Action{
												Text: "как во всей партии",
												OnTriggered: x.updateSelectedProduct(func(p *data.Product) {
													p.PointsMethod = sql.NullInt64{}
												}),
											},
											Action{
												Text: "по двум точкам",
												OnTriggered: x.updateSelectedProduct(func(p *data.Product) {
													p.PointsMethod = sql.NullInt64{2, true}
												}),
											},
											Action{
												Text: "по трём точкам",
												OnTriggered: x.updateSelectedProduct(func(p *data.Product) {
													p.PointsMethod = sql.NullInt64{3, true}
												}),
											},
										},
									},
								},
							},
						},
					},
					{
						Title:  "Работа",
						Layout: VBox{},
						Children: []Widget{
							GroupBox{
								MaxSize:  Size{0, 370},
								AssignTo: &x.gbBlocks,
								Layout:   Grid{},
								Title:    "Опрос",
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
			Composite{
				Layout: HBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{

					Composite{
						Layout: VBox{MarginsZero: true, SpacingZero: true},
					},
				},
			},
		},
	}
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

	x.w.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if !x.workStarted {
			return
		}
		*canceled = true
		x.closing = true
		x.cancelWork()
		log.Info("работа прервана, пользователь закрыл приложение")
	})

	x.productsTblMdl.SetTableView(x.tblProducts)
	x.setupProductTypesMenu()
	x.w.Run()
}

func (x *AppWindow) setupProductTypesMenu() {
	add := func(s string, f func(*data.Product)) {
		a := walk.NewAction()
		must.AbortIf(a.SetText(s))
		a.Triggered().Attach(x.updateSelectedProduct(f))
		must.AbortIf(x.menuProductTypes.Actions().Add(a))
	}
	add("как во всей партии", func(p *data.Product) {
		p.ProductTypeName = sql.NullString{}
	})

	for _, s := range data.ProductTypeNames() {
		s := s
		add(s, func(p *data.Product) {
			p.ProductTypeName = sql.NullString{s, true}
		})
	}
}

func (x *AppWindow) updateSelectedProduct(f func(p *data.Product)) func() {
	return func() {
		for _, place := range x.tblProducts.SelectedIndexes() {
			p := data.GetProductAtPlace(place)
			f(&p)
			must.AbortIf(data.DB.Save(&p))
			x.productsTblMdl.ResetProductRow(place)
		}

	}
}

func (x *AppWindow) runMainWork() {

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
	x.ctxWork, x.cancelWork = context.WithCancel(context.Background())

	workIndex := x.cbWorks.CurrentIndex()
	work := x.works[workIndex]

	what := strings.ToLower(fmt.Sprintf("Работа: %v", work.Name))

	x.AddJournalRecord(INF, fmt.Sprintf("%s: начало выполнения", what))

	go func() {

		err := x.doWork(work)

		x.w.Synchronize(func() {

			x.workStarted = false

			if x.closing {
				log.ErrIfFail(x.w.Close)
				return
			}

			for _, y := range x.enableOnWork {
				y.WindowBase.SetEnabled(!y.V)
			}
			for _, y := range x.visibleOnWork {
				y.WindowBase.SetVisible(!y.V)
			}
			if err == nil {
				x.AddJournalRecord(INF, fmt.Sprintf("%s: выполнено", what))
				return
			}
			if merry.Is(err, context.Canceled) {
				x.AddJournalRecord(WRN, fmt.Sprintf("%s: выполнение прервано", what))
				return
			}

			x.AddJournalRecord(ERR, fmt.Sprintf("%s: выполнение завершилось с ошибкой: %v", what, err))

			x.showErr(what, err)

		})
	}()
}

func (x *AppWindow) ResetProductsView() {
	x.productsTblMdl.Reset()
}

func (x *AppWindow) ResetProductRow(place int) {
	x.productsTblMdl.ResetProductRow(place)
}

func (x *AppWindow) SetInterrogateBlockValues(block int, values []float64) {
	x.w.Synchronize(func() {
		s := fmt.Sprintf("Опрос: %s блок %d : %v",
			time.Now().Format("15:04:05"),
			block,
			values)
		if err := x.gbBlocks.SetTitle(s); err != nil {
			panic(err)
		}
		for n := 0; n < 8; n++ {
			x.blocksTblMdl.values[block*8+n] = &values[n]
		}
		x.blocksTblMdl.PublishRowsReset()
	})
}

var log = structlog.New()
