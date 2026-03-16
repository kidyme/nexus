package httpport

import (
	"bytes"
	"encoding/json"
	"errors"
)

var errJSONObjectOrArrayRequired = errors.New("request body must be a JSON object or array")
var errPageMustBePositive = errors.New("page must be greater than 0")
var errSizeMustBePositive = errors.New("size must be greater than 0")

const (
	defaultPage = 1
	defaultSize = 20
)

func marshalLabels(src *map[string]interface{}) (json.RawMessage, error) {
	if src == nil {
		return nil, nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func unmarshalLabels(src json.RawMessage) (*map[string]interface{}, error) {
	if len(src) == 0 {
		return nil, nil
	}
	var labels map[string]interface{}
	if err := json.Unmarshal(src, &labels); err != nil {
		return nil, err
	}
	return &labels, nil
}

func stringPtr(v string) *string {
	return &v
}

func decodeSingleOrSlice[T any](payload []byte) (*T, []T, error) {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 {
		return nil, nil, errJSONObjectOrArrayRequired
	}

	switch trimmed[0] {
	case '{':
		var item T
		if err := json.Unmarshal(trimmed, &item); err != nil {
			return nil, nil, err
		}
		return &item, nil, nil
	case '[':
		var items []T
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return nil, nil, err
		}
		return nil, items, nil
	default:
		return nil, nil, errJSONObjectOrArrayRequired
	}
}

func normalizePagination(page, size *int) (int, int, error) {
	resultPage := defaultPage
	resultSize := defaultSize

	if page != nil {
		resultPage = *page
	}
	if size != nil {
		resultSize = *size
	}
	if resultPage <= 0 {
		return 0, 0, errPageMustBePositive
	}
	if resultSize <= 0 {
		return 0, 0, errSizeMustBePositive
	}
	return resultPage, resultSize, nil
}
