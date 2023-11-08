package importer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/RobinThrift/stuff/entities"
)

func ImportFromSnipeITJSONExport(r *http.Request) ([]*entities.Asset, map[string]string, error) {
	_, hasFileUpload := r.MultipartForm.File["import_file"]
	if !hasFileUpload {
		err := errors.New("missing import file")
		return nil, map[string]string{"import_file": err.Error()}, err
	}

	uploaded, header, err := r.FormFile("import_file")
	if err != nil {
		return nil, map[string]string{"import_file": err.Error()}, err
	}

	err = checkFileType(header, []string{"application/json"})
	if err != nil {
		return nil, map[string]string{"import_file": err.Error()}, err
	}

	importData, err := io.ReadAll(uploaded)
	if err != nil {
		return nil, map[string]string{"general": err.Error()}, err
	}

	assets, err := mapSnipeITJSONExport(importData)
	if err != nil {
		return nil, nil, err
	}

	return assets, map[string]string{}, nil
}

type snipeITJSONExport struct {
	Data []struct {
		AssetName       string  `json:"Asset Name"`
		AssetTag        string  `json:"Asset Tag"`
		Serial          string  `json:"Serial"`
		Model           string  `json:"Model"`
		ModelNo         string  `json:"Model No."`
		Category        string  `json:"Category"`
		Status          string  `json:"Status"`
		CheckedOutTo    string  `json:"Checked Out To"`
		Location        string  `json:"Location"`
		Manufacturer    string  `json:"Manufacturer"`
		Supplier        string  `json:"Supplier"`
		PurchaseDate    string  `json:"Purchase Date"`
		PurchaseCost    float64 `json:"Purchase Cost"`
		CurrentValue    int     `json:"Current Value"`
		OrderNumber     string  `json:"Order Number"`
		WarrantyExpires string  `json:"Warranty Expires"`
		Notes           string  `json:"Notes"`
	} `json:"data"`
}

func mapSnipeITJSONExport(data []byte) ([]*entities.Asset, error) {
	var imported snipeITJSONExport
	err := json.Unmarshal(data, &imported)
	if err != nil {
		return nil, err
	}

	assets := make([]*entities.Asset, 0, len(imported.Data))

	for _, data := range imported.Data {
		if data.AssetName == "" {
			continue
		}

		var purchaseDate time.Time
		if data.PurchaseDate != "" {
			purchaseDate, err = time.Parse("01.02.2006", data.PurchaseDate)
			if err != nil {
				return nil, err
			}
		}

		assets = append(assets, &entities.Asset{
			Status:       mapSnipeITStatus(data.Status),
			Tag:          fmt.Sprint(data.AssetTag),
			Name:         data.AssetName,
			Category:     data.Category,
			Model:        data.Model,
			ModelNo:      data.ModelNo,
			SerialNo:     data.Serial,
			Manufacturer: data.Manufacturer,
			Location:     data.Location,
			Notes:        data.Notes,
			Purchases: []*entities.Purchase{{
				Amount:   entities.MonetaryAmount(data.PurchaseCost),
				Supplier: data.Supplier,
				OrderNo:  data.OrderNumber,
				Date:     purchaseDate,
			}},
		})
	}

	return assets, nil
}

func mapSnipeITStatus(s string) entities.Status {
	switch s {
	case "Ready to Deploy Deployed", "Ready to Deploy":
		return entities.StatusInUse
	case "In Storage":
		return entities.StatusInStorage
	}

	return entities.StatusInStorage
}

const snipeITAPIPath = "/api/v1"

func ImportFromSnipeITAPI(ctx context.Context, serverURL string, apiKey string) ([]*entities.Asset, error) {
	if serverURL == "" {
		return nil, errors.New("missing Snipe-IT server URL")
	}

	if apiKey == "" {
		return nil, errors.New("missing Snipe-IT API key")
	}

	assets, err := getAssetsFromSnipeITAPI(ctx, serverURL, apiKey)
	if err != nil {
		return nil, err
	}

	components, err := getComponentsFromSnipeITAPI(ctx, serverURL, apiKey)
	if err != nil {
		return nil, err
	}

	assets = append(assets, components...)

	consumables, err := getConsumablesFromSnipeITAPI(ctx, serverURL, apiKey)
	if err != nil {
		return nil, err
	}

	assets = append(assets, consumables...)

	return assets, nil
}

type snipeITHardwarePage struct {
	Total int                   `json:"total"`
	Rows  []snipeITHardwareItem `json:"rows"`
}

type snipeITHardwareItem struct {
	Name     string `json:"name"`
	AssetTag string `json:"asset_tag"`
	Serial   string `json:"serial"`
	Model    struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"model"`
	ModelNumber string `json:"model_number"`
	StatusLabel struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		StatusType string `json:"status_type"`
		StatusMeta string `json:"status_meta"`
	} `json:"status_label"`
	Category struct {
		Name string `json:"name"`
	} `json:"category"`
	Manufacturer struct {
		Name string `json:"name"`
	} `json:"manufacturer"`
	Supplier struct {
		Name string `json:"name"`
	} `json:"supplier"`
	Notes       string `json:"notes"`
	OrderNumber string `json:"order_number"`
	Location    struct {
		Name string `json:"name"`
	} `json:"location"`
	Image           string `json:"image"`
	WarrantyExpires struct {
		Datetime string `json:"datetime"`
	} `json:"warranty_expires"`
	PurchaseDate struct {
		Date string `json:"date"`
	} `json:"purchase_date"`
	PurchaseCost string `json:"purchase_cost"`
	Quantity     int    `json:"qty"`
	ItemNo       string `json:"item_no"`

	CustomFields map[string]struct {
		Value string `json:"value"`
		Type  string `json:"element"`
	} `json:"custom_fields"`
}

func getAssetsFromSnipeITAPI(ctx context.Context, serverURL string, apiKey string) ([]*entities.Asset, error) {
	var assets []*entities.Asset

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	u.Path += snipeITAPIPath + "/hardware"

	fetched := 0

	for {
		q := u.Query()
		q.Add("limit", "100")
		q.Add("offset", fmt.Sprint(fetched))
		q.Add("sort", "created")
		q.Add("order", "desc")
		u.RawQuery = q.Encode()

		page, err := execSnipeITAPIRequest[snipeITHardwarePage](ctx, u.String(), apiKey)
		if err != nil {
			return nil, err
		}

		for i := range page.Rows {
			a, err := mapSnipeITAPIToAsset(&page.Rows[i])
			if err != nil {
				return nil, err
			}

			assets = append(assets, a)
		}

		fetched += len(page.Rows)
		if fetched >= page.Total {
			break
		}
	}

	return assets, nil
}

func getComponentsFromSnipeITAPI(ctx context.Context, serverURL string, apiKey string) ([]*entities.Asset, error) {
	var assets []*entities.Asset

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	u.Path += snipeITAPIPath + "/components"

	fetched := 0

	for {
		q := u.Query()
		q.Add("limit", "100")
		q.Add("offset", fmt.Sprint(fetched))
		q.Add("sort", "created")
		q.Add("order", "desc")
		u.RawQuery = q.Encode()

		page, err := execSnipeITAPIRequest[snipeITHardwarePage](ctx, u.String(), apiKey)
		if err != nil {
			return nil, err
		}

		for i := range page.Rows {
			a, err := mapSnipeITAPIToAsset(&page.Rows[i])
			if err != nil {
				return nil, err
			}

			a.Type = entities.AssetTypeComponent
			a.Quantity = uint64(page.Rows[i].Quantity)

			assets = append(assets, a)
		}

		fetched += len(page.Rows)
		if fetched >= page.Total {
			break
		}
	}

	return assets, nil
}

func getConsumablesFromSnipeITAPI(ctx context.Context, serverURL string, apiKey string) ([]*entities.Asset, error) {
	var assets []*entities.Asset

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	u.Path += snipeITAPIPath + "/consumables"

	fetched := 0

	for {
		q := u.Query()
		q.Add("limit", "100")
		q.Add("offset", fmt.Sprint(fetched))
		q.Add("sort", "created")
		q.Add("order", "desc")
		u.RawQuery = q.Encode()

		page, err := execSnipeITAPIRequest[snipeITHardwarePage](ctx, u.String(), apiKey)
		if err != nil {
			return nil, err
		}

		for i := range page.Rows {
			a, err := mapSnipeITAPIToAsset(&page.Rows[i])
			if err != nil {
				return nil, err
			}

			a.Type = entities.AssetTypeConsumable
			a.Model = page.Rows[i].ItemNo
			a.Quantity = uint64(page.Rows[i].Quantity)

			assets = append(assets, a)
		}

		fetched += len(page.Rows)
		if fetched >= page.Total {
			break
		}
	}

	return assets, nil
}

func execSnipeITAPIRequest[R any](ctx context.Context, url string, apiKey string) (result *R, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	var r R
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	result = &r

	return
}

func mapSnipeITAPIToAsset(item *snipeITHardwareItem) (*entities.Asset, error) {
	var err error
	var purchaseDate time.Time
	if item.PurchaseDate.Date != "" {
		purchaseDate, err = time.Parse("2006-01-02", item.PurchaseDate.Date)
		if err != nil {
			return nil, err
		}
	}

	var cost entities.MonetaryAmount
	if item.PurchaseCost != "" {
		costStr := strings.ReplaceAll(strings.ReplaceAll(item.PurchaseCost, ".", ""), ",", "")
		costInt, err := strconv.ParseInt(costStr, 10, 64)
		if err != nil {
			return nil, err
		}
		cost = entities.MonetaryAmount(costInt)
	}

	return &entities.Asset{
		Status:       mapSnipeITStatus(item.StatusLabel.Name),
		Type:         entities.AssetTypeAsset,
		Tag:          fmt.Sprint(item.AssetTag),
		Name:         item.Name,
		Category:     item.Category.Name,
		Model:        item.Model.Name,
		ModelNo:      item.ModelNumber,
		SerialNo:     item.Serial,
		Manufacturer: item.Manufacturer.Name,
		Location:     item.Location.Name,
		Notes:        item.Notes,
		ImageURL:     item.Image,
		Purchases: []*entities.Purchase{{
			Amount:   cost,
			Supplier: item.Supplier.Name,
			OrderNo:  item.OrderNumber,
			Date:     purchaseDate,
		}},
	}, nil

}

var errFileTypeNotAllowed = errors.New("file type not allowed")

func checkFileType(header *multipart.FileHeader, allowlist []string) error {
	return checkContentTypeAllowed(header.Header.Get("content-type"), allowlist)
}

func checkContentTypeAllowed(ct string, allowlist []string) error {
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}

	allowed := false
	for _, m := range allowlist {
		if mt == m {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("%w: %s", errFileTypeNotAllowed, mt)
	}

	return nil

}
