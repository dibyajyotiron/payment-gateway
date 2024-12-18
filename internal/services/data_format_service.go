package services

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"payment-gateway/internal/models"
	"reflect"
)

func GetDataFormat(r *http.Request) models.DataFormat {
	contentTypeHeader := r.Header.Get("Content-Type")

	switch contentTypeHeader {
	case string(models.JSON):
		return models.JSON
	case string(models.XML):
		return models.XML
	default:
		return models.JSON
	}
}

// decodes the incoming request based on content type
func DecodeRequest[T any](r *http.Request, request T) error {
	var err error

	contentType := r.Header.Get("Content-Type")
	dataFormat := models.JSON

	switch models.DataFormat(contentType) {
	case models.JSON:
		err = json.NewDecoder(r.Body).Decode(request)
	case models.XML:
		dataFormat = models.XML
		err = xml.NewDecoder(r.Body).Decode(request)
	default:
		return fmt.Errorf("unsupported content type")
	}

	if err != nil {
		return err
	}

	// Use reflection to set the DataFormat field if it exists
	targetValue := reflect.ValueOf(request).Elem()
	if field := targetValue.FieldByName("DataFormat"); field.IsValid() && field.CanSet() {
		field.SetString(string(dataFormat))
	}
	return nil
}

// EncodeRequestWithHeader will encode the response in the format request was received
//
//	It will also set the appropriate `status code` and header relating to `Content-Type`
func EncodeRequestWithHeader(w http.ResponseWriter, apiResp *models.APIResponse) error {
	// Set the Content-Type based on the requested data format
	switch apiResp.DataFormat {
	case models.JSON:
		w.Header().Set("Content-Type", string(models.JSON))
		w.WriteHeader(apiResp.StatusCode)
		return json.NewEncoder(w).Encode(apiResp)
	case models.XML:
		w.Header().Set("Content-Type", string(models.JSON))
		w.WriteHeader(apiResp.StatusCode)
		return xml.NewEncoder(w).Encode(apiResp)
	default:
		return fmt.Errorf("unsupported content type")
	}
}
