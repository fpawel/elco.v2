package old

import (
	"encoding/json"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/fpawel/elco/pkg/winapp"
	"gopkg.in/reform.v1"
	"io/ioutil"
	"path/filepath"
	"time"
)

func ExportLastParty(db *reform.DB) error {
	var party data.Party
	if err := data.GetLastParty(db, &party); err != nil {
		return err
	}
	products, err := data.GetLastPartyProducts(db, data.ProductsFilter{})
	if err != nil {
		return err
	}
	oldParty := NewOldParty(party, products)
	b, err := json.MarshalIndent(&oldParty, "", "    ")
	if err != nil {
		return err
	}
	importFileName, err := importFileName()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(importFileName, b, 0666)

}

func ImportLastParty(db *reform.DB) error {

	importFileName, err := importFileName()
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(importFileName)
	if err != nil {
		return err
	}
	var oldParty OldParty
	if err := json.Unmarshal(b, &oldParty); err != nil {
		return err
	}
	party, products := oldParty.Party()

	if err := data.EnsureProductTypeName(db, party.ProductTypeName); err != nil {
		return err
	}
	party.CreatedAt = time.Now().Add(-3 * time.Hour)
	if err := db.Save(&party); err != nil {
		return err
	}
	for _, p := range products {
		p.PartyID = party.PartyID
		if p.ProductTypeName.Valid {
			if err := data.EnsureProductTypeName(db, p.ProductTypeName.String); err != nil {
				return err
			}
		}
		if err := db.Save(&p); err != nil {
			return merry.Appendf(err, "product: serial: %v place: %d",
				p.Serial, p.Place)
		}
	}
	return nil
}

func importFileName() (string, error) {
	appDataFolderPath, err := winapp.AppDataFolderPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataFolderPath, "elco", "export-party.json"), nil
}
