package vmodel

import (
	"github.com/fpawel/elco.v2/internal/data"
)


type LastPartyProducts struct {
	p data.Party
	m *LastPartyProductsTable
	m2 *LastPartyPlacesTable

}


func NewProducts() *LastPartyProducts {
	x := &LastPartyProducts{}
	x.m = &LastPartyProductsTable{
		party: &x.p,
		products:make(productsMap),
	}
	x.m2 = &LastPartyPlacesTable{
		party: &x.p,
		m:x.m,
	}
	x.m.m2 = x.m2
	return x
}


func (x *LastPartyProducts) Party() data.Party {
	p := x.p
	p.Products = append([]data.ProductInfo{}, x.p.Products...)
	return p
}

func (x *LastPartyProducts) Invalidate() {
	x.p = data.GetLastPartyWithProductsInfo( data.ProductsFilter{} )
	x.m.invalidate()
	x.m2.PublishRowsReset()
}

func (x *LastPartyProducts) ProductsTable() *LastPartyProductsTable {
	return x.m
}

func (x *LastPartyProducts) PlacesTable() *LastPartyPlacesTable {
	return x.m2
}