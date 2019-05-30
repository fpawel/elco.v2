package view

import (
	"fmt"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/lxn/walk"
)

type BlocksTable struct {
	walk.TableModelBase
	productsTable *ProductsTable
}

func (x *BlocksTable) RowCount() int {
	return 12
}

func (x *BlocksTable) Value(row, col int) interface{} {
	if col == 0 {
		return fmt.Sprintf("Блок %d", row+1)
	}
	return ""
}

func (x *BlocksTable) Checked(row int) bool {
	for _, p := range x.productsTable.products {
		if p.Place/8 == row && p.Production {
			return true
		}
	}
	return false
}

func (x *BlocksTable) SetChecked(row int, checked bool) error {

	data.SetBlockChecked(row, checked)

	n := row * 8
	for i := n; i < n+8; i++ {
		if x.productsTable.products[i].ProductID != 0 {
			x.productsTable.products[i].Production = checked
		}
	}

	x.productsTable.PublishRowsReset()
	return nil
}
