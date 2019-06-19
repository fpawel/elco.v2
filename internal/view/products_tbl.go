package view

import (
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/powerman/must"
	"os"
	"path/filepath"
)

type ProductsTable struct {
	walk.TableModelBase
	products    []data.ProductInfo
	blocksTable *BlocksTable
	tableView   *walk.TableView
}

func newProductsModels() (*ProductsTable, *BlocksTable) {
	x, y := new(ProductsTable), new(BlocksTable)
	x.blocksTable = y
	y.productsTable = x

	return x, y
}

func (x *ProductsTable) setup() {
	x.products = make([]data.ProductInfo, 96)
	for _, p := range data.GetLastPartyProductsInfo() {
		x.products[p.Place] = p
	}
}

func (x *ProductsTable) ProductAtPlace(place int) data.ProductInfo {
	return x.products[place]
}

func AllProductsTableColumns() (columns []declarative.TableViewColumn) {
	for _, c := range data.AllProductsFields {
		precision, _ := productsColPrecision[c]
		columns = append(columns, declarative.TableViewColumn{
			Title:     productColName[c],
			Width:     80,
			Precision: precision,
		})
	}
	return
}

func (x *ProductsTable) SetTableView(tableView *walk.TableView) {
	x.tableView = tableView
	x.setup()
}

func (x *ProductsTable) ResetProductRow(place int) {
	x.tableView.Synchronize(func() {
		x.products[place] = data.GetProductInfoAtPlace(place)
		x.PublishRowChanged(place)
	})
}

func setupProductsColumns(tableView *walk.TableView, fields []data.ProductField) {
	tableView.SetVisible(false)
	defer tableView.SetVisible(true)
	if err := tableView.Columns().Clear(); err != nil {
		panic(err)
	}
	for _, c := range fields {
		col := walk.NewTableViewColumn()
		must.AbortIf(col.SetTitle(productColName[c]))
		must.AbortIf(col.SetWidth(80))
		must.AbortIf(tableView.Columns().Add(col))
		precision, f := productsColPrecision[c]
		if !f {
			precision = 3
		}
		must.AbortIf(col.SetPrecision(precision))
	}
}

func (x *ProductsTable) Reset() {
	x.setup()
	x.PublishRowsReset()
	x.blocksTable.PublishRowsReset()
}

func (x *ProductsTable) RowCount() int {
	return 96
}

func (x *ProductsTable) Value(row, col int) interface{} {

	p := x.products[row]
	if p.ProductID == 0 {
		if col == 0 {
			return data.FormatPlace(row)
		}
		return ""
	}

	if v := p.FieldValue(data.AllProductsFields[col]); v != nil {
		return v
	}
	return ""
}

func (x *ProductsTable) Checked(row int) bool {

	p := x.products[row]
	if p.ProductID == 0 {
		return false
	}
	return p.Production
}

func (x *ProductsTable) SetChecked(row int, checked bool) error {

	p := data.GetProductAtPlace(row)
	p.Production = checked
	if err := data.DB.Save(&p); err != nil {
		panic(err)
	}

	x.products[row].ProductID = p.ProductID
	x.products[row].Production = p.Production
	x.products[row].Place = row

	x.blocksTable.PublishRowChanged(row / 8)
	return nil
}

func (x *ProductsTable) StyleCell(c *walk.CellStyle) {

	if (c.Row()/8)%2 != 0 {
		c.BackgroundColor = walk.RGB(245, 245, 245)
	}

	if c.Col() < 0 || c.Col() >= len(data.AllProductsFields) {
		return
	}

	p := x.products[c.Row()]
	if p.ProductID == 0 {
		return
	}

	field := data.AllProductsFields[c.Col()]
	c.Font = fontDefault
	switch field {
	case data.ProductFieldPlace:
		if p.HasFirmware {
			c.Image = check16Png
		}
	case data.ProductFieldSerial:
		c.Font = fontSerial
		c.TextColor = walk.RGB(128, 0, 0)
	}

	chk := p.OkFieldValue(field)
	if chk.Valid {
		if chk.Bool {
			c.TextColor = walk.RGB(0, 0, 0xFF)
		} else {
			c.TextColor = walk.RGB(0xFF, 0, 0)
		}
	}
}

func newFont(family string, pointSize int, style walk.FontStyle) *walk.Font {
	font, err := walk.NewFont(family, pointSize, style)
	if err != nil {
		panic(err)
	}
	return font
}

func newBitmapFromFile(filename string) walk.Image {
	img, err := walk.NewImageFromFile(filepath.Join(filepath.Dir(os.Args[0]), "img", filename))
	if err != nil {
		panic(err)
	}
	return img
}

var (
	fontSerial  = newFont("Segoe UI", 12, walk.FontItalic)
	fontDefault = newFont("Segoe UI", 12, 0)

	check16Png = newBitmapFromFile("check16.png")

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
