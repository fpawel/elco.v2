package api

import (
	"database/sql"
	"github.com/fpawel/elco.v2/internal/data"
	"strconv"
	"strings"
)

type Product struct {
	ProductID       int64  `json:"product_id,omitempty"`
	Serial          int64  `json:"serial,omitempty"`
	ProductTypeName string `json:"product_type_name,omitempty"`
	PointsMethod    int64  `json:"points_method,omitempty"`
	Note            string `json:"note,omitempty"`
}

type LastPartySvc struct{}

func (_ LastPartySvc) Products(_ struct{}, products *[]*Product) error {
	*products = make([]*Product, 96)

	for _, p := range data.GetLastPartyProducts(data.WithSerials) {
		(*products)[p.Place] = &Product{
			ProductID:       p.ProductID,
			Serial:          p.Serial.Int64,
			ProductTypeName: p.ProductTypeName.String,
			PointsMethod:    p.PointsMethod.Int64,
			Note:            p.Note.String,
		}
	}
	return nil
}

func (x LastPartySvc) GetSerialAtPlace(place [1]int, serial *string) error {
	product := data.GetProductAtPlace(place[0])
	if product.Serial.Valid {
		*serial = strconv.Itoa(int(product.Serial.Int64))
	}
	return nil
}

func (x LastPartySvc) SetSerialAtPlace(p struct {
	Place  int
	Serial string
}, _ *struct{}) error {
	var serial sql.NullInt64
	if len(strings.TrimSpace(p.Serial)) > 0 {
		n, err := strconv.ParseInt(p.Serial, 10, 64)
		if err != nil {
			return err
		}
		serial = sql.NullInt64{n, true}
	}
	product := data.GetProductAtPlace(p.Place)
	product.Serial = serial
	if err := data.DB.Save(&product); err != nil {
		return err
	}
	//view.ResetProductAtPlace(p.Place)
	return nil
}

func (x LastPartySvc) SetProductTypeAtPlace(p struct {
	Place           int
	ProductTypeName string
}, _ *struct{}) error {

	product := data.GetProductAtPlace(p.Place)
	product.ProductTypeName.String = strings.TrimSpace(p.ProductTypeName)
	product.ProductTypeName.Valid = len(product.ProductTypeName.String) > 0
	if err := data.DB.Save(&product); err != nil {
		return err
	}
	//view.ResetProductAtPlace(p.Place)
	return nil
}

func (x LastPartySvc) SetPointsMethodAtPlace(a [2]int, _ *struct{}) error {

	product := data.GetProductAtPlace(a[0])
	switch a[1] {
	case 1:
		product.PointsMethod = sql.NullInt64{2, true}
	case 2:
		product.PointsMethod = sql.NullInt64{3, true}
	default:
		product.PointsMethod = sql.NullInt64{}
	}
	if err := data.DB.Save(&product); err != nil {
		return err
	}
	//view.ResetProductAtPlace(a[0])
	return nil
}

type EccInfoSvc struct{}

func (_ EccInfoSvc) ProductTypeNames(_ struct{}, productTypeNames *[]string) error {
	*productTypeNames = data.ProductTypeNames()
	return nil
}
