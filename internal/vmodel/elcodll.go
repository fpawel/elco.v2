package vmodel

import (
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
)

type LastPartyProductsTable struct {
	walk.ReflectTableModelBase
	party    *data.Party
	fields   []data.ProductField
	m2       *LastPartyPlacesTable
	products productsMap
}

type productsMap = map[int]*data.ProductInfo

func (x *LastPartyProductsTable) ProductAt(place int) *data.ProductInfo {

	p, _ := x.products[place]
	return p
}

func (x *LastPartyProductsTable) RowCount() int {
	return 96
}

func (x *LastPartyProductsTable) Value(row, col int) interface{} {
	p := x.ProductAt(row)
	if p == nil {
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

func (x *LastPartyProductsTable) Checked(row int) bool {
	p := x.ProductAt(row)
	if p == nil {
		return false
	}
	return p.Production
}

func (x *LastPartyProductsTable) SetChecked(row int, checked bool) error {

	p := x.ProductAt(row)
	if p == nil {
		return nil
	}
	p.Production = checked

	product := new(data.Product)
	if err := data.GetProductAtPlace(p.Place, product); err != nil {
		return err
	}
	product.Production = checked
	err := data.DB.Save(product)

	x.m2.PublishRowChanged(row / 8)

	return err
}

func (x *LastPartyProductsTable) StyleCell(c *walk.CellStyle) {

	if (c.Row()/8)%2 != 0 {
		c.BackgroundColor = walk.RGB(245, 245, 245)
	}

	if c.Col() < 0 || c.Col() >= len(x.fields) {
		return
	}

	p := x.ProductAt(c.Row())
	if p == nil {
		return
	}

	field := x.fields[c.Col()]
	c.Font = fontDefault
	switch field {
	case data.ProductFieldPlace:
		if p.HasFirmware {
			c.Image = "assets/png16/check.png"
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
