package vmodel

import (
	"database/sql"
	"github.com/fpawel/elco.v2/internal/data"
)

func Products() []data.ProductInfo {
	products := make([]data.ProductInfo, 96)
	for _, p := range data.GetLastPartyProductsInfo() {
		products[p.Place] = p
	}
	return products
}

func (x *Products) InvalidateAtPlace(place int) {
	x.products[place] = data.GetProductInfoAtPlace(place)
	x.m.PublishRowChanged(place)
	x.m2.PublishRowChanged(place / 8)
}

func (x *Products) Invalidate() {
	x.Setup()
	x.m.PublishRowsReset()
	x.m2.PublishRowsReset()
}

func (x *Products) ProductsTable() *ProductsTable {
	return x.m
}

func (x *Products) PlacesTable() *PlacesTable {
	return x.m2
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
