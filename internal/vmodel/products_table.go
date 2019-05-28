package vmodel

import (
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
)

type ProductsTable struct {
	walk.ReflectTableModelBase
	fields   []data.ProductField
	m2       *PlacesTable
	products []data.ProductInfo
}

func (x *ProductsTable) Fields() []data.ProductField {
	return append([]data.ProductField{}, x.fields...)
}

func (x *ProductsTable) ProductAt(place int) data.ProductInfo {
	return x.products[place]
}

func (x *ProductsTable) RowCount() int {
	return 96
}

func (x *ProductsTable) Value(row, col int) interface{} {
	p := x.ProductAt(row)
	if p.ProductID == 0 {
		if col == 0 {
			return data.FormatPlace(row)
		}
		return ""
	}

	if v := p.FieldValue(x.fields[col]); v != nil {
		return v
	}
	return ""
}

func (x *ProductsTable) Checked(row int) bool {
	p := x.ProductAt(row)
	if p.ProductID == 0 {
		return false
	}
	return p.Production
}

func (x *ProductsTable) SetChecked(row int, checked bool) error {

	p := data.GetProductAtPlace(row)
	p.Production = checked
	if err := data.DB.Save(&p); err != nil {
		return err
	}
	x.products[row].ProductID = p.ProductID
	x.products[row].Production = p.Production
	x.products[row].Place = row
	x.m2.PublishRowChanged(row / 8)
	return nil
}

func (x *ProductsTable) StyleCell(c *walk.CellStyle) {

	if (c.Row()/8)%2 != 0 {
		c.BackgroundColor = walk.RGB(245, 245, 245)
	}

	if c.Col() < 0 || c.Col() >= len(x.fields) {
		return
	}

	p := x.ProductAt(c.Row())
	if p.ProductID == 0 {
		return
	}

	field := x.fields[c.Col()]
	c.Font = fontDefault
	switch field {
	case data.ProductFieldPlace:
		if p.HasFirmware {
			c.Image = "img/check16.png"
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

var fontSerial, fontDefault *walk.Font

func init() {
	fontSerial, _ = walk.NewFont("Segoe UI", 12, walk.FontItalic)
	fontDefault, _ = walk.NewFont("Segoe UI", 12, 0)
}
