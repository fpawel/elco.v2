package main

import (
	"fmt"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"strconv"
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
		pbCancelWork *walk.PushButton
		btnRun       *walk.SplitButton
	)

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
				VerticalFixed: true,
				Layout:        HBox{},
				Children: []Widget{
					SplitButton{
						Text: "Партия",
						MenuItems: []MenuItem{
							Action{
								Text: "Создать новую",
								OnTriggered: func() {
									if walk.MsgBox(mw.w, "Новая партия",
										"Подтвердите необходимость создания новой партии",
										walk.MsgBoxIconQuestion|walk.MsgBoxYesNo) != win.IDYES {
										return
									}
									data.CreateNewParty()
									lastPartyProducts.Invalidate()
								},
							},
							Action{
								Text: "Параметры",
								OnTriggered: func() {
									runPartyDialog()
								},
							},
							Action{
								Text: "Ввод",
								OnTriggered: func() {
									for place := 0; place < 96; place++ {
										s := ""
										p := lastPartyProducts.ProductsTable().ProductAt(place)
										if p != nil && p.Serial.Valid {
											s = strconv.FormatInt(p.Serial.Int64, 10)
										}
										formPartySerialsSetCell(place/8+1, place%8+1, s)
									}
									formPartySerialsShow()
								},
							},
							Action{
								Text: "Выбрать годные ЭХЯ",
								OnTriggered: func() {
									data.SetOnlyOkProductsProduction()
									lastPartyProducts.Invalidate()
								},
							},
						},
					},
					SplitButton{
						Text:      "Управление",
						AssignTo:  &btnRun,
						MenuItems: []MenuItem{},
					},
					PushButton{
						AssignTo: &pbCancelWork,
						Text:     "Прервать",
						OnClicked: func() {
							cancelComport()
						},
					},

					Label{
						AssignTo:  &mw.lblWorkTime,
						TextColor: walk.RGB(0, 128, 0),
					},
					Label{
						AssignTo: &mw.lblWork,
					},
					mw.DelayHelp.Widget(),
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
								Model:                    lastPartyProducts.ProductsTable(),
								OnItemActivated: func() {

									p := lastPartyProducts.ProductsTable().ProductAt(mw.tblProducts.CurrentIndex())
									if p == nil {
										return
									}
									//mw.w.SetVisible(false)
									x := &FirmwareDialog{productID: p.ProductID}
									x.run()
									//mw.w.SetVisible(true)

								},
								OnKeyDown: func(key walk.Key) {
									switch key {

									case walk.KeyInsert:

									case walk.KeyDelete:

									}

								},
								OnCurrentIndexChanged: func() {

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

	pbCancelWork.SetVisible(false)
	mw.invalidateProductsColumns()
	mw.w.Run()

	if err := settings.Save(); err != nil {
		panic(err)
	}
}

func (x AppMainWindow) invalidateProductsColumns() {
	_ = mw.tblProducts.Columns().Clear()
	for _, c := range data.NotEmptyProductsFields(lastPartyProducts.Party().Products) {
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
