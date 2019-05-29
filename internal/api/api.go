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

type LastPartySvc struct {
}

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

type EccInfoSvc struct{}

func (_ EccInfoSvc) ProductTypeNames(_ struct{}, productTypeNames *[]string) error {
	*productTypeNames = data.ProductTypeNames()
	return nil
}
