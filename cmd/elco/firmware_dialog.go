package main

import (
	"fmt"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type FirmwareDialog struct {
	product                data.ProductInfo
	w                      *walk.Dialog
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
	b data.FirmwareBytes
}

type firmwareTemperatureRow struct {
	t, f, s       float64
	neT, neF, neS *walk.NumberEdit
}

func (x *firmwareTemperatureRow) Row(t, f, s float64) []Widget {
	return []Widget{
		NumberEdit{
			AssignTo: &x.neT,
			Value:    t,
		},
		NumberEdit{
			AssignTo: &x.neF,
			Value:    f,
		},
		NumberEdit{
			AssignTo: &x.neS,
			Value:    s,
		},
		PushButton{Text: "+", MaxSize: Size{Width: 30}},
		PushButton{Text: "-", MaxSize: Size{Width: 30}},
	}
}

func setComboBoxText(cb *walk.ComboBox, text string) error {
	for n, s := range cb.Model().([]string) {
		if s == text {
			return cb.SetCurrentIndex(n)
		}
	}
	return cb.SetCurrentIndex(-1)
}

func (x *FirmwareDialog) changeFirmwareInfoSource() {

	if x.rbStored.Checked() && x.product.HasFirmware {
		p := data.GetProductByProductID(x.product.ProductID)
		x.b = data.FirmwareBytes(p.Firmware)
	} else {
		f, err := x.product.Firmware()
		if err == nil {
			x.b = f.Bytes()
		} else {
			x.b = make(data.FirmwareBytes, data.FirmwareSize)
			for i := range x.b {
				x.b[i] = 0xFF
			}
		}
	}

	i := x.b.FirmwareInfo(x.cbPlace.CurrentIndex())

	for _, err := range []error{
		x.edDate.SetDate(i.Time),
		x.neSerial.SetValue(float64(i.Serial)),
		x.neScaleBegin.SetValue(float64(i.ScaleBeg)),
		x.neScaleEnd.SetValue(float64(i.ScaleEnd)),
		x.neSens.SetValue(float64(i.Sensitivity)),
		setComboBoxText(x.cbGas, i.Gas),
		setComboBoxText(x.cbUnits, i.Units),
		setComboBoxText(x.cbType, i.ProductType),
	} {
		if err != nil {
			panic(err)
		}
	}
}

func (x *FirmwareDialog) run() {

	if x.product.HasFirmware {
		p := data.GetProductByProductID(x.product.ProductID)
		x.b = data.FirmwareBytes(p.Firmware)
	} else {
		f, err := x.product.Firmware()
		if err == nil {
			x.b = f.Bytes()
		} else {
			x.b = make(data.FirmwareBytes, data.FirmwareSize)
			for i := range x.b {
				x.b[i] = 0xFF
			}
		}
	}

	var places []string
	for i := 0; i < 96; i++ {
		places = append(places, data.FormatPlace(i))
	}

	var wPts []Widget
	for _, t := range []float64{-40, -30, -20, -5, 0, 20, 30, 40, 45, 50} {
		wPts = append(wPts,
			Label{
				Text: fmt.Sprintf("%v", t),
			},
			Label{
				Text: fmt.Sprintf("%v", x.b.F(t)),
			},
			Label{
				Text: fmt.Sprintf("%v", x.b.S(t)),
			},
		)
	}

	if err := (Dialog{
		AssignTo: &x.w,
		Layout:   HBox{},
		Children: []Widget{
			Composite{
				Layout: VBox{SpacingZero: true, MarginsZero: true},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 4},
						Children: []Widget{

							Label{
								TextAlignment: AlignFar,
								Text:          "Место",
							},
							ComboBox{
								AssignTo: &x.cbPlace,
								Model:    places,
								Value:    data.FormatPlace(x.product.Place),
							},

							Label{
								Text:          "Серийный №",
								TextAlignment: AlignFar,
							},
							NumberEdit{
								AssignTo: &x.neSerial,
							},

							Label{
								Text:          "Дата",
								TextAlignment: AlignFar,
							},
							DateEdit{
								AssignTo: &x.edDate,
							},

							Label{
								Text:          "Исполнение",
								TextAlignment: AlignFar,
							},
							ComboBox{
								AssignTo: &x.cbType,
								Model:    data.ProductTypeNames(),
							},

							Label{
								Text:          "Газ",
								TextAlignment: AlignFar,
							},
							ComboBox{
								AssignTo: &x.cbGas,
								Model:    data.GasesNames(),
							},

							Label{
								Text:          "Ед.изм.",
								TextAlignment: AlignFar,
							},
							ComboBox{
								AssignTo: &x.cbUnits,
								Model:    data.UnitsNames(),
							},

							Label{
								Text:          "Шкала",
								TextAlignment: AlignFar,
							},
							Composite{
								Layout: HBox{MarginsZero: true, SpacingZero: true},
								Children: []Widget{
									NumberEdit{
										AssignTo: &x.neScaleBegin,
										Decimals: 1,
									},
									NumberEdit{
										AssignTo: &x.neScaleEnd,
										Decimals: 1,
									},
								},
							},

							Label{
								Text:          "Чувст-ть",
								TextAlignment: AlignFar,
							},
							NumberEdit{
								AssignTo: &x.neSens,
								Decimals: 3,
							},
						},
					},
					Composite{
						Layout:   Grid{Columns: 3},
						Children: wPts,
					},
				},
			},

			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{

					PushButton{
						Text: "Записать",
					},
					PushButton{
						Text: "Считать",
					},
					PushButton{
						Text: "График",
					},

					RadioButtonGroup{
						Buttons: []RadioButton{
							{
								Text:      "Записано",
								AssignTo:  &x.rbStored,
								OnClicked: x.changeFirmwareInfoSource,
							},
							{
								AssignTo:  &x.rbCalculated,
								Text:      "Расчитано",
								OnClicked: x.changeFirmwareInfoSource,
							},
						},
					},
				},
			},
		},
	}).Create(mw.w); err != nil {
		panic(err)
	}
	x.w.SetFont(mw.w.Font())

	x.rbStored.SetChecked(x.product.HasFirmware)
	//x.changeFirmwareInfoSource()
	x.w.Run()

}
