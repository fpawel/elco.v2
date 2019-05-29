package main

import (
	"database/sql"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"math"
)

func runPartyDialog() {
	var (
		dlg        *walk.Dialog
		cbType     *walk.ComboBox
		btn        *walk.PushButton
		saveOnEdit bool
		edNote     *walk.TextEdit
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
		lastPartyProducts.Invalidate()
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

		TextEdit{
			AssignTo:   &edNote,
			ColumnSpan: 4,
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

		Composite{ColumnSpan: 3},
		PushButton{
			AssignTo: &btn,
			Text:     "Закрыть",
			OnClicked: func() {
				dlg.Accept()
			},
		})

	dialog := Dialog{
		Title:         "Параметры партии",
		Font:          Font{PointSize: 12, Family: "Segoe UI"},
		AssignTo:      &dlg,
		Layout:        Grid{Columns: 4},
		DefaultButton: &btn,
		CancelButton:  &btn,
		Children:      widgets,
	}
	if err := dialog.Create(mw.w); err != nil {
		walk.MsgBox(mw.w, "Параметры партии", err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
		return
	}
	saveOnEdit = true
	dlg.Run()

}
