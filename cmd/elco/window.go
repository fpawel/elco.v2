package main

import (
	"fmt"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type AppMainWindow struct {
	w           *walk.MainWindow
	tblProducts *walk.TableView
	lblWork,
	lblWorkTime *walk.Label
	DelayHelp *delayHelp
}

func runMainWindow() {
	var (
		tbStop, tbStart *walk.ToolButton
		cbWorks         *walk.ComboBox
		panelTools      *walk.ScrollView
	)

	log.Debug("create main window")

	lastPartyProducts.Setup()

	var columns []TableViewColumn
	for _, c := range data.NotEmptyProductsFields(lastPartyProducts.Products()) {
		precision, _ := productsColPrecision[c]
		columns = append(columns, TableViewColumn{
			Title:     productColName[c],
			Width:     80,
			Precision: precision,
		})
	}

	if err := (MainWindow{
		AssignTo: &mw.w,
		Title: "Партия ЭХЯ " + (func() string {
			p := data.GetLastParty()
			return fmt.Sprintf("№%d %s", p.PartyID, p.CreatedAt.Format("02.01.2006"))
		}()),
		Name:       "MainWindow",
		Font:       Font{PointSize: 12, Family: "Segoe UI"},
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
		Size:       Size{800, 600},
		Layout:     VBox{},

		Children: []Widget{

			ScrollView{
				AssignTo:      &panelTools,
				VerticalFixed: true,
				Layout:        HBox{SpacingZero: true},
				Children: []Widget{
					ToolButton{
						Text:        "Новая загрузка",
						Image:       "img/new25.png",
						ToolTipText: "Создать новую загрузку",
						OnClicked: func() {
							if walk.MsgBox(mw.w, "Новая партия",
								"Подтвердите необходимость создания новой партии",
								walk.MsgBoxIconQuestion|walk.MsgBoxYesNo) != win.IDYES {
								return
							}
							data.CreateNewParty()
							lastPartyProducts.Invalidate()
						},
					},

					ToolButton{
						Text:        "Выбрать годные ЭХЯ",
						Image:       "img/check25m.png",
						ToolTipText: "Выбрать годные ЭХЯ",
						OnClicked: func() {
							data.SetOnlyOkProductsProduction()
							lastPartyProducts.Invalidate()
						},
					},

					ToolButton{
						Text:        "Паспорта и итоговая таблица",
						Image:       "img/pdf25.png",
						ToolTipText: "Паспорта и итоговая таблица",
						OnClicked: func() {
							data.SetOnlyOkProductsProduction()
							lastPartyProducts.Invalidate()
						},
					},

					VSpacer{MinSize: Size{10, 0}},

					ComboBox{
						AssignTo: &cbWorks,
						Model: []string{
							"Опрос",
							"Термокомпенсация",
							"Погрешность",
							"Прошивка",
						},
						CurrentIndex: 0,
					},

					VSpacer{MinSize: Size{3, 0}},

					ToolButton{
						AssignTo:    &tbStop,
						Visible:     false,
						Text:        "Прервать выполнение операции",
						Image:       "img/stop25.png",
						ToolTipText: "Прервать выполнение операции",
						OnClicked:   func() {},
					},

					ToolButton{
						AssignTo:    &tbStart,
						Text:        "Начать выполнение выбранной операции",
						Image:       "img/start25.png",
						ToolTipText: "Начать выполнение выбранной операции",
						OnClicked:   func() {},
					},

					Label{
						AssignTo:  &mw.lblWorkTime,
						TextColor: walk.RGB(0, 128, 0),
					},
					Label{
						AssignTo: &mw.lblWork,
					},
					mw.DelayHelp.Widget(),
					ScrollView{
						VerticalFixed: true,
						Layout:        Grid{},
					},
					ToolButton{
						Text:        "Ввод серийных номеров ЭХЯ",
						Image:       "img/edit25.png",
						ToolTipText: "Ввод серийных номеров ЭХЯ",
						OnClicked: func() {
							formPartySerialsSetVisible(true)
						},
					},
					ToolButton{
						Text:        "Параметры загрузки",
						Image:       "img/sets25g.png",
						ToolTipText: "Параметры загрузки",
						OnClicked:   runPartyDialog,
					},
					ToolButton{
						Text:        "Настройки",
						Image:       "img/sets25b.png",
						ToolTipText: "Настройки",
						OnClicked: func() {
							tbStart.SetVisible(!tbStart.Visible())
							tbStop.SetVisible(!tbStop.Visible())
							if err := panelTools.Invalidate(); err != nil {
								panic(err)
							}
							if err := tbStop.Invalidate(); err != nil {
								panic(err)
							}
							if err := tbStart.Invalidate(); err != nil {
								panic(err)
							}
						},
					},
				},
			},
			ScrollView{
				Layout: HBox{},
				Children: []Widget{

					GroupBox{
						Layout: Grid{},
						Title:  "Настраиваемые ЭХЯ",
						Children: []Widget{

							TableView{
								AssignTo:                 &mw.tblProducts,
								NotSortableByHeaderClick: true,
								LastColumnStretched:      true,
								CheckBoxes:               true,
								MultiSelection:           true,
								Model:                    lastPartyProducts.ProductsTable(),
								Columns:                  columns,
								OnItemActivated: func() {
									p := lastPartyProducts.ProductsTable().ProductAt(mw.tblProducts.CurrentIndex())
									if p.ProductID != 0 {
										runFirmwareDialog(p)
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
					GroupBox{
						Layout: Grid{},
						Title:  "Опрос",
						Children: []Widget{
							TableView{
								Model:      lastPartyProducts.PlacesTable(),
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
				},
			},
		},
	}).Create(); err != nil {
		panic(err)
	}
	log.Debug("run main window")
	mw.w.Run()

	if err := settings.Save(); err != nil {
		panic(err)
	}
}

func (x AppMainWindow) invalidateProductsColumns() {
	_ = mw.tblProducts.Columns().Clear()
	for _, c := range data.NotEmptyProductsFields(lastPartyProducts.Products()) {
		col := walk.NewTableViewColumn()
		_ = col.SetTitle(productColName[c])
		_ = col.SetWidth(80)
		_ = mw.tblProducts.Columns().Add(col)

		if precision, f := productsColPrecision[c]; f {
			_ = col.SetPrecision(precision)
		} else {
			_ = col.SetPrecision(3)
		}
	}
}

func newComboBoxComport(comboBox **walk.ComboBox, key string) ComboBox {
	return ComboBox{
		AssignTo:     comboBox,
		Model:        getComports(),
		CurrentIndex: comportIndex(getIniStr(key)),
		OnMouseDown: func(_, _ int, _ walk.MouseButton) {
			cb := *comboBox
			n := cb.CurrentIndex()
			_ = cb.SetModel(getComports())
			_ = cb.SetCurrentIndex(n)
		},
		OnCurrentIndexChanged: func() {
			putIniStr(key, (*comboBox).Text())
		},
	}
}

func getIniStr(key string) string {
	s, _ := settings.Get(key)
	return s
}

func putIniStr(key, value string) {
	if err := settings.Put(key, value); err != nil {
		panic(err)
	}
}

func getComports() []string {
	ports, _ := comport.Ports()
	return ports
}

func comportIndex(portName string) int {
	ports, _ := comport.Ports()
	for i, s := range ports {
		if s == portName {
			return i
		}
	}
	return -1
}

var (
	productColName = map[data.ProductField]string{
		data.ProductFieldPlace:        "№",
		data.ProductFieldSerial:       "Зав.№",
		data.ProductFieldFon20:        "фон.20",
		data.ProductField2Fon20:       "фон.20.2",
		data.ProductFieldSens20:       "ч.20",
		data.ProductFieldKSens20:      "Кч.20",
		data.ProductFieldFonMinus20:   "фон.-20",
		data.ProductFieldSensMinus20:  "ч.-20",
		data.ProductFieldFon50:        "фон.50",
		data.ProductFieldSens50:       "ч.50",
		data.ProductFieldKSens50:      "Кч.50",
		data.ProductFieldI24:          "ПГС2",
		data.ProductFieldI35:          "ПГС3",
		data.ProductFieldI26:          "ПГС2",
		data.ProductFieldI17:          "ПГС1",
		data.ProductFieldNotMeasured:  "неизмеряемый",
		data.ProductFieldType:         "ИБЯЛ",
		data.ProductFieldPointsMethod: "метод",
		data.ProductFieldNote:         "примечание",
	}
	productsColPrecision = map[data.ProductField]int{
		data.ProductFieldKSens20: 1,
		data.ProductFieldKSens50: 1,
	}
)
