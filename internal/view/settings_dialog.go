package view

import (
	"database/sql"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"math"
)

func (x *AppWindow) runSettingsDialog() {
	var (
		dlg                                 *walk.Dialog
		cbType                              *walk.ComboBox
		btn                                 *walk.PushButton
		saveOnEdit                          bool
		edNote                              *walk.LineEdit
		cbComport, cbComportGas, cbChipType *walk.ComboBox
	)

	party := data.GetLastParty(data.WithProducts)

	saveParty := func() {
		if !saveOnEdit {
			return
		}
		if err := data.DB.Save(&party); err != nil {
			walk.MsgBox(dlg, "Ошибка данных", err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
			return
		}
		x.tblProducts.Model().(*ProductsTable).PublishRowsReset()
	}

	types := data.ProductTypeNames()

	widgets := []Widget{
		Label{Text: "Исполнение:", TextAlignment: AlignFar},
		ComboBox{
			Model:    types,
			AssignTo: &cbType,
			CurrentIndex: func() int {
				for i, x := range types {
					if x == party.ProductTypeName {
						return i
					}
				}
				return -1
			}(),
			OnCurrentIndexChanged: func() {
				party.ProductTypeName = types[cbType.CurrentIndex()]
				saveParty()
			},
		},
	}

	neField := func(what string, decimals int, p *float64) {
		var ne *walk.NumberEdit
		widgets = append(widgets,
			Label{Text: what, TextAlignment: AlignFar},
			NumberEdit{
				Decimals: decimals,
				Value:    *p,
				AssignTo: &ne,
				MinValue: 0,
				MaxValue: math.MaxFloat64,
				OnValueChanged: func() {
					*p = ne.Value()
					saveParty()
				},
			})
	}

	neField2 := func(what string, decimals int, p *sql.NullFloat64) {
		var (
			ne *walk.NumberEdit
			cb *walk.CheckBox
		)
		widgets = append(widgets,
			Label{Text: what, TextAlignment: AlignFar},
			Composite{
				Layout: HBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{

					CheckBox{
						MaxSize:  Size{15, 0},
						AssignTo: &cb,
						Checked:  p.Valid,
						OnCheckedChanged: func() {
							p.Valid = cb.Checked()
							ne.SetEnabled(p.Valid)
							saveParty()
						},
					},

					NumberEdit{
						Enabled:  p.Valid,
						Decimals: decimals,
						Value:    p.Float64,
						AssignTo: &ne,
						OnValueChanged: func() {
							p.Float64 = ne.Value()
							saveParty()
						},
					},
				},
			},
		)
	}

	neField("ПГС1:", 1, &party.Concentration1)
	neField("ПГС2:", 1, &party.Concentration2)
	neField("ПГС3:", 1, &party.Concentration3)
	neField2("Фон.мин.", 2, &party.MinFon)
	neField2("Фон.мaкс.", 2, &party.MaxFon)
	neField2("Dфон.мaкс.", 2, &party.MaxDFon)
	neField2("Кч20.мин", 2, &party.MinKSens20)
	neField2("Кч20.макс", 2, &party.MaxKSens20)
	neField2("Кч50.мин.", 2, &party.MinKSens50)
	neField2("Кч50.макс", 2, &party.MaxKSens50)
	neField2("Dt.мин.", 2, &party.MinDTemp)
	neField2("Dt.мaкс", 2, &party.MaxDTemp)
	neField2("Dn.мaкс", 2, &party.MaxDNotMeasured)

	widgets = append(widgets,
		Label{
			Text: "Примечание",
		},
		LineEdit{
			AssignTo:   &edNote,
			ColumnSpan: 3,
			Text:       party.Note.String,
			OnTextChanged: func() {
				if len(edNote.Text()) == 0 {
					party.Note.Valid = false
				} else {
					party.Note.Valid = true
					party.Note.String = edNote.Text()
				}
				saveParty()
			},
		},
	)

	dialog := Dialog{
		Title:         "Параметры",
		Font:          Font{PointSize: 12, Family: "Segoe UI"},
		AssignTo:      &dlg,
		Layout:        VBox{},
		DefaultButton: &btn,
		CancelButton:  &btn,
		Children: []Widget{

			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Children: []Widget{
					Composite{
						Layout:   Grid{Columns: 4},
						Children: widgets,
					},

					Composite{
						Layout: VBox{},
						Children: []Widget{
							GroupBox{
								Layout: VBox{},
								Title:  "СОМ порт",
								Children: []Widget{
									Label{Text: "Блок измерения"},
									comboBoxComport(&cbComport, "Comport"),
									Label{Text: "Газовый блок"},
									comboBoxComport(&cbComportGas, "ComportGas"),
								},
							},

							Label{Text: "Тип микросхемы"},
							comboBox(&cbChipType, "CHIP_TYPE", []string{
								"24LC16",
								"24LC64",
								"24LC256",
							}),

							Label{Text: "Т\"С окр.среды"},
							NumberEdit{},
						},
					},
				},
			},

			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Children: []Widget{
					ScrollView{
						VerticalFixed: true,
						Layout:        Grid{SpacingZero: true, MarginsZero: true},
					},
					PushButton{
						AssignTo: &btn,
						Text:     "Закрыть",
						OnClicked: func() {
							dlg.Accept()
						},
					},
				},
			},
		},
	}
	if err := dialog.Create(x.w); err != nil {
		walk.MsgBox(x.w, "Параметры партии", err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
		return
	}
	saveOnEdit = true
	dlg.Run()
}

func comboBoxComport(comboBox **walk.ComboBox, key string) ComboBox {
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

func comboBox(comboBox **walk.ComboBox, key string, model []string) ComboBox {
	return ComboBox{
		AssignTo:     comboBox,
		Model:        model,
		CurrentIndex: comboBoxIndex(getIniStr(key), model),
		OnCurrentIndexChanged: func() {
			putIniStr(key, (*comboBox).Text())
		},
	}
}

func getIniStr(key string) string {

	s, _ := walk.App().Settings().Get(key)
	return s
}

func putIniStr(key, value string) {
	if err := walk.App().Settings().Put(key, value); err != nil {
		panic(err)
	}
}

func comboBoxIndex(s string, m []string) int {
	for i, x := range m {
		if s == x {
			return i
		}
	}
	return -1
}

func comportIndex(portName string) int {
	ports, _ := comport.Ports()
	return comboBoxIndex(portName, ports)
}

func getComports() []string {
	ports, _ := comport.Ports()
	return ports
}
