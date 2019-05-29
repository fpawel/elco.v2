package main

import (
	"database/sql"
	"github.com/fpawel/elco.v2/internal/data"
	"strconv"
	"strings"
)

type ProductsSvc struct {
}

func (x ProductsSvc) GetSerialAtPlace(place [1]int, serial *string) error {
	product := data.GetProductAtPlace(place[0])
	if product.Serial.Valid {
		*serial = strconv.Itoa(int(product.Serial.Int64))
	}
	return nil
}

func (x ProductsSvc) SetSerialAtPlace(p struct {
	Place  int
	Serial string
}, _ *struct{}) (err error) {

	product := data.GetProductAtPlace(p.Place)

	if len(strings.TrimSpace(p.Serial)) == 0 {
		product.Serial.Valid = false
	} else {
		if product.Serial.Int64, err = strconv.ParseInt(p.Serial, 10, 64); err != nil {
			return
		}
		product.Serial.Valid = true
	}
	if err = data.DB.Save(&product); err == nil {
		lastPartyProducts.ProductsTable().PublishRowChanged(p.Place)
	}
	return
}

func (x ProductsSvc) SetProductTypeAtPlace(p struct {
	Place           int
	ProductTypeName string
}, _ *struct{}) (err error) {

	product := data.GetProductAtPlace(p.Place)
	product.ProductTypeName.String = strings.TrimSpace(p.ProductTypeName)
	product.ProductTypeName.Valid = len(product.ProductTypeName.String) > 0
	if err = data.DB.Save(&product); err == nil {
		lastPartyProducts.ProductsTable().PublishRowChanged(p.Place)
	}
	return err
}

func (x ProductsSvc) SetPointsMethodAtPlace(a [2]int, _ *struct{}) (err error) {

	product := data.GetProductAtPlace(a[0])
	switch a[1] {
	case 1:
		product.PointsMethod = sql.NullInt64{2, true}
	case 2:
		product.PointsMethod = sql.NullInt64{3, true}
	default:
		product.PointsMethod = sql.NullInt64{}
	}
	if err = data.DB.Save(&product); err == nil {
		lastPartyProducts.ProductsTable().PublishRowChanged(a[0])
	}
	return err
}
