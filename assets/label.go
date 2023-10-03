package assets

import (
	"bytes"
	"fmt"
	"image/png"
	"net/url"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func (a *Asset) Labels(baseURL *url.URL, barcodeSize int) ([]Label, error) {
	labels := make([]Label, 0, len(a.Children)+1)

	var barcode Barcode
	assetURL := ""
	if baseURL != nil {
		u := *baseURL
		u.Path = fmt.Sprintf("/assets/%v", a.Tag)
		assetURL = u.String()

		image, err := generateBarcode(assetURL, barcodeSize)
		if err != nil {
			return nil, err
		}
		barcode = Barcode{Size: barcodeSize, Value: assetURL, Image: image}
	} else {
		image, err := generateBarcode(a.Tag, barcodeSize)
		if err != nil {
			return nil, err
		}
		barcode = Barcode{Size: barcodeSize, Value: a.Tag, Image: image}
	}

	labels = append(labels, Label{
		Tag:          a.Tag,
		Name:         a.Name,
		URL:          assetURL,
		LocationCode: fmtLocationCode(a.Location, a.PositionCode),
		Barcode:      barcode,
	})

	for _, part := range a.Parts {
		labels = append(labels, Label{
			Tag:          part.Tag,
			Name:         part.Name,
			URL:          assetURL,
			LocationCode: fmtLocationCode(part.Location, part.PositionCode),
			Barcode:      barcode,
		})
	}

	return labels, nil
}

type Label struct {
	Tag          string
	Name         string
	LocationCode string
	URL          string
	Barcode      Barcode
}

type Barcode struct {
	Size  int
	Value string
	Image []byte
}

func generateBarcode(value string, size int) ([]byte, error) {
	var b bytes.Buffer
	code, err := qr.Encode(value, qr.M, qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("error encoding url as QR code: %w", err)
	}

	qrCode, err := barcode.Scale(code, size, size)
	if err != nil {
		return nil, fmt.Errorf("error scaling url QR code image to %dx%[1]dpx: %w", size, err)
	}

	err = png.Encode(&b, qrCode)
	if err != nil {
		return nil, fmt.Errorf("error encoding url as QR code as PNG: %w", err)
	}

	return b.Bytes(), nil

}

func fmtLocationCode(loc string, posCode string) string {
	if posCode != "" {
		return loc + "(" + posCode + ")"
	}
	return loc
}
