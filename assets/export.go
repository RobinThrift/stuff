package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

func exportAssetsAsJSON(w io.Writer, assets []*Asset) error {
	encoder := json.NewEncoder(w)

	forExport := make([]*apiAsset, 0, len(assets))

	for _, asset := range assets {
		purchases := make([]*apiPurchase, 0, len(asset.Purchases))
		for _, p := range asset.Purchases {
			purchases = append(purchases, &apiPurchase{
				Supplier: p.Supplier,
				OrderNo:  p.OrderNo,
				Date:     p.Date,
				Amount:   p.Amount,
				Currency: p.Currency,
			})
		}

		customAttrs := make([]apiCustomAttr, 0, len(asset.CustomAttrs))
		for _, ca := range asset.CustomAttrs {
			customAttrs = append(customAttrs, apiCustomAttr(ca))
		}

		forExport = append(forExport, &apiAsset{
			Tag:           asset.Tag,
			Status:        asset.Status,
			Name:          asset.Name,
			Category:      asset.Category,
			Model:         asset.Model,
			ModelNo:       asset.ModelNo,
			SerialNo:      asset.SerialNo,
			Manufacturer:  asset.Manufacturer,
			Notes:         asset.Notes,
			WarrantyUntil: asset.WarrantyUntil,
			CustomAttrs:   customAttrs,
			Location:      asset.Location,
			PositionCode:  asset.PositionCode,
			Purchases:     purchases,
		})
	}

	err := encoder.Encode(forExport)
	if err != nil {
		return fmt.Errorf("error exporting assets as JSON: %w", err)
	}

	return nil
}

var csvColumns = []byte(strings.Join([]string{
	"Tag", "Status", "Name", "Category", "Model", "Model No", "Serial No",
	"Manufacturer", "Notes", "Warranty Until",
	"Location", "Position Code",
	"Purchase Supplier", "Purchase OrderNo", "Purchase Date", "Purchase Amount", "Purchase Currency",
}, ","))

func newCSVWriteLineErr(err error) error {
	return fmt.Errorf("error exporting assets to csv: error writing line to buffer: %w", err)
}

func exportAssetsAsCSV(w io.Writer, assets []*Asset) error {
	var line bytes.Buffer

	_, err := line.Write(csvColumns)
	if err != nil {
		return newCSVWriteLineErr(err)
	}

	_, err = line.WriteTo(w)
	if err != nil {
		return fmt.Errorf("error exporting assets to csv: error writing line to output: %w", err)
	}

	var values = make([]string, 17)

	for _, asset := range assets {
		line.Reset()

		values[0] = asset.Tag
		values[1] = string(asset.Status)
		values[2] = asset.Name
		values[3] = asset.Category
		values[4] = asset.Model
		values[5] = asset.ModelNo
		values[6] = asset.SerialNo
		values[7] = asset.Manufacturer
		values[8] = asset.Notes

		if asset.WarrantyUntil.IsZero() {
			values[9] = ""
		} else {
			values[9] = asset.WarrantyUntil.Format(time.RFC3339)
		}

		values[10] = asset.Location
		values[11] = asset.PositionCode

		if len(asset.Purchases) != 0 {
			lastPurchase := len(asset.Purchases) - 1
			values[12] = asset.Purchases[lastPurchase].Supplier
			values[13] = asset.Purchases[lastPurchase].OrderNo

			if asset.Purchases[lastPurchase].Date.IsZero() {
				values[14] = ""
			} else {
				values[14] = asset.Purchases[lastPurchase].Date.Format(time.RFC3339)
			}

			values[15] = fmt.Sprint(asset.Purchases[lastPurchase].Amount)
			values[16] = asset.Purchases[lastPurchase].Currency
		}

		_, err = line.WriteRune('\n')
		if err != nil {
			return newCSVWriteLineErr(err)
		}

		_, err = line.WriteString(values[0])
		if err != nil {
			return newCSVWriteLineErr(err)
		}

		for _, s := range values[1:] {
			_, err = line.WriteString(",")
			if err != nil {
				return newCSVWriteLineErr(err)
			}
			_, err = line.WriteString(s)
			if err != nil {
				return newCSVWriteLineErr(err)
			}
		}

		_, err = line.WriteTo(w)
		if err != nil {
			return fmt.Errorf("error exporting assets to csv: error writing line to output: %w", err)
		}
	}

	return nil
}
