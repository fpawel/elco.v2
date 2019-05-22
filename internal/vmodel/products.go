package vmodel

import (
	"database/sql"
	"github.com/fpawel/elco.v2/internal/data"
)

type Products struct {
	products []data.ProductInfo
	m        *ProductsTable
	m2       *PlacesTable
}

func NewProducts() *Products {
	x := &Products{
		products: make([]data.ProductInfo, 96),
	}

	x.m = &ProductsTable{
		products: x.products,
	}
	x.m2 = &PlacesTable{
		products: x.products,
		m:        x.m,
	}
	x.m.m2 = x.m2
	return x
}

func (x *Products) Invalidate() {
	for _, p := range data.GetLastPartyWithProductsInfo(data.ProductsFilter{}).Products {
		x.products[p.Place] = p
	}
	x.m.invalidate()
	x.m2.PublishRowsReset()
}

func (x *Products) ProductsTable() *ProductsTable {
	return x.m
}

func (x *Products) PlacesTable() *PlacesTable {
	return x.m2
}

func (x *Products) Products() []data.ProductInfo {
	return append([]data.ProductInfo{}, x.products...)
}

func (x *Products) SetProductSerialAt(place int, serial sql.NullInt64) error {
	if err := data.SetProductSerialAtPlace(place, serial); err != nil {
		return err
	}
	x.products[place] = data.GetProductInfoAtPlace(place)
	x.m.PublishRowChanged(place)
	x.m2.PublishRowChanged(place / 8)
	return nil
}
