package view

import (
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/fpawel/elco.v2/internal/imgchart"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/image/bmp"
	"os"
	"path/filepath"
)

type firmwareDialog struct {
	w                                         *walk.Dialog
	storeFirmware, calcFirmware, currFirmware data.FirmwareBytes
	rbStored,
	rbCalculated *walk.RadioButton
	neSerial,
	neScaleBegin,
	neScaleEnd,
	neSens *walk.NumberEdit
	edDate *walk.DateEdit
	cbPlace,
	cbType,
	cbGas,
	cbUnits *walk.ComboBox
	img     *walk.ImageView
	product data.ProductInfo
}

func setComboBoxText(cb *walk.ComboBox, text string) error {
	for n, s := range cb.Model().([]string) {
		if s == text {
			return cb.SetCurrentIndex(n)
		}
	}
	return cb.SetCurrentIndex(-1)
}

func imgChartFileName() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "chart.bmp")
}

func (x *firmwareDialog) saveChartToFile() {
	imgChartFileName := imgChartFileName()
	out, err := os.Create(imgChartFileName)
	if err != nil {
		panic(err)
	}
	imgChart := imgchart.New(x.currFirmware, 600, 350)
	if err := bmp.Encode(out, imgChart); err != nil {
		panic(err)
	}
	if err := out.Close(); err != nil {
		panic(err)
	}
}

func newFirmwareBytes() data.FirmwareBytes {
	x := make(data.FirmwareBytes, data.FirmwareSize)
	for i := range x {
		x[i] = 0xFF
	}
	return x
}

func runFirmwareDialog(owner walk.Form, product data.ProductInfo) {

	x := &firmwareDialog{
		product:       product,
		storeFirmware: newFirmwareBytes(),
		calcFirmware:  newFirmwareBytes(),
	}

	if x.product.HasFirmware {
		x.storeFirmware = data.FirmwareBytes(data.GetProductByProductID(product.ProductID).Firmware)
	}
	if b, err := product.Firmware(); err == nil {
		x.calcFirmware = b.Bytes()
	}

	if x.product.HasFirmware {
		x.currFirmware = append(data.FirmwareBytes{}, x.storeFirmware...)
	} else {
		x.currFirmware = append(data.FirmwareBytes{}, x.calcFirmware...)
	}
	x.saveChartToFile()
	x.run(owner)
}

func (x *firmwareDialog) invalidate() {
	i := x.currFirmware.FirmwareInfo(x.cbPlace.CurrentIndex())
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
	x.saveChartToFile()
	img, err := walk.NewImageFromFile(imgChartFileName())
	if err != nil {
		panic(err)
	}
	if err := x.img.SetImage(img); err != nil {
		panic(err)
	}

}

func (x *firmwareDialog) rbFirmwareInfoSourceClick() {
	if x.rbStored.Checked() {
		x.currFirmware = append(data.FirmwareBytes{}, x.storeFirmware...)
	} else {
		x.currFirmware = append(data.FirmwareBytes{}, x.calcFirmware...)
	}
	x.invalidate()
}

func (x *firmwareDialog) dialog() Dialog {
	var places []string
	for i := 0; i < 96; i++ {
		places = append(places, data.FormatPlace(i))
	}
	return Dialog{
		AssignTo:   &x.w,
		Layout:     VBox{},
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},

		Children: []Widget{
			Composite{
				Layout: HBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 4},
						Children: []Widget{
							Label{
								Text:          "Место",
								TextAlignment: AlignFar,
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
								Text:          "Исполнение",
								TextAlignment: AlignFar,
							},
							ComboBox{
								AssignTo: &x.cbType,
								Model:    data.ProductTypeNames(),
							},

							Label{
								Text:          "Чувст-ть",
								TextAlignment: AlignFar,
							},
							NumberEdit{
								AssignTo: &x.neSens,
								Decimals: 3,
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
							NumberEdit{
								AssignTo: &x.neScaleBegin,
								Decimals: 1,
							},

							Label{
								Text:          "Газ",
								TextAlignment: AlignFar,
							},
							ComboBox{
								AssignTo: &x.cbGas,
								Model:    data.GasesNames(),
							},

							Composite{},
							NumberEdit{
								AssignTo: &x.neScaleEnd,
								Decimals: 1,
							},
						},
					},

					Composite{
						Layout: VBox{},
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

							DateEdit{
								AssignTo: &x.edDate,
							},
							//ScrollView{Layout:VBox{MarginsZero:true}, HorizontalFixed:true},
						},
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{
					RadioButtonGroup{
						Buttons: []RadioButton{
							{
								Text:      "Записано",
								AssignTo:  &x.rbStored,
								OnClicked: x.rbFirmwareInfoSourceClick,
							},
							{
								AssignTo:  &x.rbCalculated,
								Text:      "Расчитано",
								OnClicked: x.rbFirmwareInfoSourceClick,
							},
						},
					},
					ScrollView{Layout: HBox{MarginsZero: true}, VerticalFixed: true},
				},
			},
			ImageView{
				Visible:  true,
				AssignTo: &x.img,
				//Image: "chart.bmp",
			},
		},
	}
}

func (x *firmwareDialog) run(owner walk.Form) {

	if err := x.dialog().Create(owner); err != nil {
		panic(err)
	}
	x.w.SetFont(owner.Font())
	x.rbStored.SetChecked(x.product.HasFirmware)
	x.rbCalculated.SetChecked(!x.product.HasFirmware)
	x.invalidate()
	x.w.Run()
	x.w.Close(0)
}

var firmwareTemperatures = []int{-40, -30, -20, -5, 0, 20, 30, 40, 45, 50}
