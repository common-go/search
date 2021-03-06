package search

import (
	"errors"
	"reflect"
)

type Extractor struct {
	Page string
	Limit string
	FirstLimit string
}

func NewExtractor(options...string) *Extractor {
	var page, limit, firstLimit string
	if len(options) >= 1 && len(options[0]) > 0 {
		page = options[0]
	} else {
		page = "Page"
	}
	if len(options) >= 2 && len(options[1]) > 0 {
		limit = options[1]
	} else {
		limit = "Limit"
	}
	if len(options) >= 3 && len(options[2]) > 0 {
		firstLimit = options[2]
	} else {
		firstLimit = "FirstLimit"
	}
	return &Extractor{Page: page, Limit: limit, FirstLimit: firstLimit}
}

func (e *Extractor) Extract(m interface{}) (int64, int64, int64, error) {
	if sModel, ok0 := m.(*SearchModel); ok0 {
		return sModel.Page, sModel.Limit, sModel.FirstLimit, nil
	}
	var page, limit, firstLimit int64
	page = -1
	limit = -1
	firstLimit = -1
	value := reflect.Indirect(reflect.ValueOf(m))
	t := value.Type()
	numField := t.NumField()
	// numField := value.NumField()
	for i := 0; i < numField; i++ {
		if sModel1, ok1 := value.Field(i).Interface().(*SearchModel); ok1 {
			return sModel1.Page, sModel1.Limit, sModel1.FirstLimit, nil
		} else {
			n := t.Field(i).Name
			if n == e.Page {
				if p, ok := value.Field(i).Interface().(int64); ok {
					page = p
				}
			} else if n == e.Limit {
				if p, ok := value.Field(i).Interface().(int64); ok {
					limit = p
				}
			} else if n == e.FirstLimit {
				if p, ok := value.Field(i).Interface().(int64); ok {
					firstLimit = p
				}
			}
			if page >= 0 && limit >=0 && firstLimit >= 0 {
				return page, limit, firstLimit, nil
			}
		}
	}
	return page, limit, firstLimit, nil
}

func ExtractSearch(m interface{}) (int64, int64, int64, error) {
	if sModel, ok := m.(*SearchModel); ok {
		return sModel.Page, sModel.Limit, sModel.FirstLimit, nil
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*SearchModel); ok {
				return sModel1.Page, sModel1.Limit, sModel1.FirstLimit, nil
			}
		}
		return 0, 0, 0, errors.New("cannot extract sort, pageIndex, pageSize, firstPageSize from model")
	}
}
