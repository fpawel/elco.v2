package vmodel

import (
	"fmt"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
)

type LastPartyPlacesTable struct {
	walk.ReflectTableModelBase
	party *data.Party
	m *LastPartyProductsTable
}

func (x *LastPartyProductsTable) invalidate(){
	x.products = make(productsMap)
	for i := range x.party.Products{
		p := &x.party.Products[i]
		x.products[p.Place] = p
	}
	x.fields = data.NotEmptyProductsFields(x.party.Products)
	x.PublishRowsReset()
}


func (x *LastPartyPlacesTable) RowCount() int {
	return 12
}

func (x *LastPartyPlacesTable) Value(row, col int) interface{} {
	if col == 0 {
		return fmt.Sprintf("Блок %d", row + 1)
	}
	return ""
}

func (x *LastPartyPlacesTable) Checked(row int) bool {
	for _,p := range x.m.products{
		if p.Place / 8 == row && p.Production {
			return true
		}
	}
	return false
}

func (x *LastPartyPlacesTable) SetChecked(row int, checked bool) error {

	data.SetBlockChecked(row, checked)

	n := row * 8
	for i := n; i < n + 8 ; i++{
		if p,f := x.m.products[i]; f {
			p.Production = checked
			x.m.PublishRowChanged(i)
		}
	}
	return nil
}