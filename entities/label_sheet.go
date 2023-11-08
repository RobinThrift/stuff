package entities

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"

	"github.com/RobinThrift/stuff/frontend"
	"github.com/signintech/gopdf"
)

type PageSize string

const (
	PageSizeA4 PageSize = "A4"
)

type LabelSize struct {
	FontSize float64

	// Height of a single label in mm including padding.
	Height float64
	// Width of a single label in mm including padding.
	Width float64

	// VerticalPadding applied to the inside of the label in mm.
	VerticalPadding float64
	// HorizontalPadding applied to the inside of the label in mm.
	HorizontalPadding float64

	// VerticalSpacing between each label in mm.
	VerticalSpacing float64
	// HorizontalSpacing between each label in mm.
	HorizontalSpacing float64
}

type PageLayout struct {
	// Cols per page.
	Cols int
	// Rows per page.
	Rows int

	// MarginLeft of the page in mm.
	MarginLeft float64
	// MarginTop of the page in mm.
	MarginTop float64
	// MarginRight of the page in mm.
	MarginRight float64
	// MarginBottom of the page in mm.
	MarginBottom float64
}

type Sheet struct {
	PageSize      PageSize
	Labels        []Label
	LabelSize     LabelSize
	PageLayout    PageLayout
	SkipNumLabels int
	PrintBorders  bool
}

func (s PageSize) rect() gopdf.Rect {
	if s == PageSizeA4 {
		return *gopdf.PageSizeA4
	}
	return gopdf.Rect{}
}

func (s *Sheet) Generate() ([]byte, error) { //nolint gocognit Will refactor later
	sheet, err := s.newPDF()
	if err != nil {
		return nil, err
	}

	labels := s.Labels
	if s.SkipNumLabels != 0 {
		skipLabels := make([]Label, s.SkipNumLabels)
		labels = append(skipLabels, s.Labels...) //nolint gocritic this is fine
	}

	labelWidth := s.LabelSize.Width
	labelHeight := s.LabelSize.Height
	innerWidth := s.LabelSize.Width - s.LabelSize.HorizontalPadding
	innerHeight := s.LabelSize.Height - s.LabelSize.VerticalPadding

	usableWidth, usableHeight := s.usableSize(sheet)

	if s.PageLayout.Cols == 0 {
		s.PageLayout.Cols = int(usableWidth / (labelWidth + s.LabelSize.HorizontalSpacing))
	}

	if s.PageLayout.Rows == 0 {
		s.PageLayout.Rows = int(usableHeight / (labelHeight + s.LabelSize.VerticalSpacing))
	}

	perPage := s.PageLayout.Cols * s.PageLayout.Rows
	if perPage == 0 {
		return nil, errors.New("perPageCalculation returned 0")
	}
	numPages := max(len(s.Labels)/perPage, 1)

	lineHeight, err := sheet.MeasureCellHeightByText("measured text")
	if err != nil {
		return nil, err
	}

	barcodeSize := min((innerWidth/3), innerHeight) - lineHeight
	textWidth := (innerWidth / 3) * 2

	for i := 0; i <= numPages-1; i++ {
		sheet.AddPage()

		sheet.SetXY(0, 0)

		for r := 0; r < s.PageLayout.Rows; r++ {
			for c := 0; c < s.PageLayout.Cols; c++ {
				labelIndex := i*perPage + r*s.PageLayout.Cols + c
				if labelIndex >= len(labels) {
					break
				}

				xPos := float64(c)*(labelWidth+s.LabelSize.HorizontalSpacing) + s.PageLayout.MarginLeft + s.LabelSize.HorizontalPadding/2
				yPos := float64(r)*(labelHeight+s.LabelSize.VerticalSpacing) + s.PageLayout.MarginTop + s.LabelSize.VerticalPadding/2

				label := labels[labelIndex]
				if label.Tag == "" {
					continue
				}

				if s.PrintBorders {
					sheet.RectFromUpperLeft(xPos-s.LabelSize.HorizontalPadding, yPos-s.LabelSize.VerticalPadding, labelWidth, labelHeight)
				}

				barcode := convertTo8BitPNG(&label)

				err = sheet.ImageFrom(barcode, xPos, yPos, &gopdf.Rect{
					W: barcodeSize,
					H: barcodeSize,
				})
				if err != nil {
					return nil, fmt.Errorf("error adding barcode to PDF: %w", err)
				}

				sheet.SetXY(xPos+s.LabelSize.HorizontalPadding/2, yPos+barcodeSize+lineHeight)

				err = sheet.Text(label.Tag)
				if err != nil {
					return nil, err
				}

				sheet.SetXY(xPos+barcodeSize+1, yPos)

				nameRect := &gopdf.Rect{W: textWidth, H: labelHeight}
				_, nameTextHeight, err := sheet.IsFitMultiCell(nameRect, label.Name)
				if err != nil {
					return nil, err
				}

				nameRect.H = nameTextHeight
				err = sheet.MultiCellWithOption(nameRect, label.Name, gopdf.CellOption{
					Align: gopdf.Left,
				})
				if err != nil {
					return nil, err
				}

				sheet.SetXY(xPos+barcodeSize+1, yPos+nameTextHeight+lineHeight)

				err = sheet.Text(label.LocationCode)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	var output bytes.Buffer
	_, err = sheet.WriteTo(&output)
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

func (s *Sheet) newPDF() (*gopdf.GoPdf, error) {
	sheet := gopdf.GoPdf{}
	sheet.Start(gopdf.Config{
		PageSize: s.PageSize.rect(),
		Unit:     gopdf.UnitMM,
	})
	sheet.SetMargins(s.PageLayout.MarginLeft, s.PageLayout.MarginTop, s.PageLayout.MarginRight, s.PageLayout.MarginBottom)

	fontData, err := frontend.PDFFont()
	if err != nil {
		return nil, err
	}

	err = sheet.AddTTFFontData("OpenSans", fontData)
	if err != nil {
		return nil, err
	}

	err = sheet.SetFont("OpenSans", "", s.LabelSize.FontSize)
	if err != nil {
		return nil, err
	}

	return &sheet, nil
}

func (s *Sheet) usableSize(sheet *gopdf.GoPdf) (width float64, height float64) {
	r := s.PageSize.rect()
	width = sheet.PointsToUnits(r.W) - s.PageLayout.MarginLeft - s.PageLayout.MarginRight
	height = sheet.PointsToUnits(r.H) - s.PageLayout.MarginTop - s.PageLayout.MarginBottom
	return
}

// Inspired by https://github.com/transcom/mymove/blob/9aeb2ec733fde04aa104c545d5b689a06faa0989/pkg/paperwork/generator.go#L67
// LICENSE MIT
func convertTo8BitPNG(label *Label) image.Image {
	b := label.Barcode.Image.Bounds()
	eightBitImg := image.NewRGBA(b)
	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			newPixel := color.RGBAModel.Convert(label.Barcode.Image.At(x, y))
			eightBitImg.Set(x, y, newPixel)
		}
	}

	return eightBitImg
}
