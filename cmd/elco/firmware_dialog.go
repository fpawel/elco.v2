package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type FirmwareDialog struct {
	productID int64
	w *walk.Dialog
	rbStored, rbCalculated *walk.RadioButton
	neSerial,
	neScaleBegin,
	neScaleEnd,
	neSens *walk.NumberEdit
	edDate *walk.DateEdit
	cbPlace,
	cbType,
	cbGas,
	cbUnits *walk.ComboBox


}

func (x *FirmwareDialog) run() {

	if err := (Dialog{
		AssignTo:&x.w,
		Layout:VBox{},
		Children:[]Widget{

			ScrollView{
				Layout:HBox{},
				VerticalFixed:true,
				Children:[]Widget{
					Label{
						Text:"Место",
					},
					ComboBox{
						AssignTo:&x.cbPlace,
					},
					RadioButtonGroup{
						Buttons:[]RadioButton{
							{
								Text:"Записано",
								AssignTo:&x.rbStored,
							},
							{
								AssignTo:&x.rbCalculated,
								Text:"Расчитано",
							},
						},
					},
				},
			},

			Composite{
				Layout:HBox{},
				Children:[]Widget{
					Composite{
						Layout:VBox{},
						Children:[]Widget{

							Label{
								Text:"Серийный номер",
							},
							NumberEdit{
								AssignTo:&x.neSerial,
							},

							Label{
								Text:"Дата",
							},
							DateEdit{
								AssignTo:&x.edDate,
							},

							Label{
								Text:"Исполнение",
							},
							ComboBox{
								AssignTo:&x.cbType,
							},

							Label{
								Text:"Газ",
							},
							ComboBox{
								AssignTo:&x.cbGas,
							},

							Label{
								Text:"Единицы измерения",
							},
							ComboBox{
								AssignTo:&x.cbUnits,
							},

							Label{
								Text:"Начало шкалы",
							},
							NumberEdit{
								AssignTo:&x.neScaleBegin,
							},

							Label{
								Text:"Конец шкалы",
							},
							NumberEdit{
								AssignTo:&x.neScaleEnd,
							},

							Label{
								Text:"Чувствительность",
							},
							NumberEdit{
								AssignTo:&x.neSens,
							},
						},
					},

					ScrollView{
						HorizontalFixed:true,
						Layout:VBox{},
						Children:[]Widget{
							PushButton{
								Text:"Записать",
							},
							PushButton{
								Text:"Считать",
							},
							PushButton{
								Text:"График",
							},
						},
					},
				},
			},
		},
	}).Create(mw.w); err != nil{
		panic(err)
	}
	x.w.SetFont(mw.w.Font())

	x.w.Run()

}