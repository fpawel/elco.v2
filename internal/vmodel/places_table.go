package vmodel

import (
	"fmt"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
)

type PlacesTable struct {
	walk.ReflectTableModelBase
	products []data.ProductInfo
	m        *ProductsTable
}

func (x *ProductsTable) invalidate() {
	x.fields = data.NotEmptyProductsFields(x.products)
	x.PublishRowsReset()
}

func (x *PlacesTable) RowCount() int {
	return 12
}

func (x *PlacesTable) Value(row, col int) interface{} {
	if col == 0 {
		return fmt.Sprintf("Блок %d", row+1)
	}
	return ""
}

func (x *PlacesTable) Checked(row int) bool {
	for _, p := range x.products {
		if p.Place/8 == row && p.Production {
			return true
		}
	}
	return false
}

func (x *PlacesTable) SetChecked(row int, checked bool) error {

	data.SetBlockChecked(row, checked)

	n := row * 8
	for i := n; i < n+8; i++ {
		if x.products[i].ProductID != 0 {
			x.products[i].Production = checked
			x.m.PublishRowChanged(i)
		}
	}
	return nil
}
