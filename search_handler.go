package search

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type SearchHandler struct {
	searchService             SearchService
	searchModelType           reflect.Type
	LogWriter                 LogWriter
	quickSearch               bool
	isExtendedSearchModelType bool
	Resource                  string
	embedField                string
	userId                    string

	// Search by GET
	paramIndex            map[string]int
	searchModelParamIndex map[string]int
	searchModelIndex      int
}

const (
	PageSizeDefault    = 10
	MaxPageSizeDefault = 10000
)

func NewSearchHandler(searchService SearchService, searchModelType reflect.Type, logService LogWriter, quickSearch bool, resource string, userId string, embedField string) *SearchHandler {
	isExtendedSearchModelType := IsExtendedFromSearchModel(searchModelType)
	if isExtendedSearchModelType == false {
		panic(errors.New(searchModelType.Name() + " isn't SearchModel struct nor extended from SearchModel struct!"))
	}

	paramIndex := buildParamIndex(searchModelType)
	searchModelParamIndex := buildParamIndex(reflect.TypeOf(SearchModel{}))
	searchModelIndex := findSearchModelIndex(searchModelType)

	return &SearchHandler{searchService: searchService, searchModelType: searchModelType, LogWriter: logService, quickSearch: quickSearch, isExtendedSearchModelType: isExtendedSearchModelType, Resource: resource, paramIndex: paramIndex, searchModelIndex: searchModelIndex, searchModelParamIndex: searchModelParamIndex, userId: userId, embedField: embedField}
}

func buildParamIndex(searchModelType reflect.Type) map[string]int {
	params := map[string]int{}

	numField := searchModelType.NumField()
	for i := 0; i < numField; i++ {
		field := searchModelType.Field(i)
		fullJsonTag := field.Tag.Get("json")
		tagDetails := strings.Split(fullJsonTag, ",")
		if len(tagDetails) > 0 && len(tagDetails[0]) > 0 {
			params[tagDetails[0]] = i
		}
	}

	return params
}

func findSearchModelIndex(searchModelType reflect.Type) int {
	numField := searchModelType.NumField()
	for i := 0; i < numField; i++ {
		if searchModelType.Field(i).Type == reflect.TypeOf(&SearchModel{}) {
			return i
		}
	}
	return -1
}

// Check valid and change value of pagination to correct
func (c *SearchHandler) repairSearchModel(searchModel *SearchModel, currentUserId string) {
	searchModel.CurrentUserId = currentUserId

	pageSize := searchModel.Limit
	if pageSize > MaxPageSizeDefault || pageSize < 1 {
		pageSize = PageSizeDefault
	}
	pageIndex := searchModel.Page
	if searchModel.Page < 1 {
		pageIndex = 1
	}

	if searchModel.Limit != pageSize {
		searchModel.Limit = pageSize
	}

	if searchModel.Page != pageIndex {
		searchModel.Page = pageIndex
	}
}

func (c *SearchHandler) ProcessSearchModel(sm interface{}, currentUserId string) {
	if s, ok := sm.(*SearchModel); ok { // Is SearchModel struct
		c.repairSearchModel(s, currentUserId)
	} else { // Is extended from SearchModel struct
		value := reflect.Indirect(reflect.ValueOf(sm))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find SearchModel field of extended struct
			if s, ok := value.Field(i).Interface().(*SearchModel); ok {
				c.repairSearchModel(s, currentUserId)
				break
			}
		}
	}
}

func (c *SearchHandler) CreateSearchModelObject() interface{} {
	var searchModel = reflect.New(c.searchModelType).Interface()

	if c.isExtendedSearchModelType {
		value := reflect.Indirect(reflect.ValueOf(searchModel))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find SearchModel field of extended struct
			if _, ok := value.Field(i).Interface().(*SearchModel); ok {
				// Init SearchModel to avoid nil value
				value.Field(i).Set(reflect.ValueOf(&SearchModel{}))
				break
			}
		}
	}
	return searchModel
}

func (c *SearchHandler) mapParamsToSearchModel(searchModel interface{}, params url.Values) interface{} {
	value := reflect.Indirect(reflect.ValueOf(searchModel))
	if value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	for paramKey, valueArr := range params {
		paramValue := ""
		if len(valueArr) > 0 {
			paramValue = valueArr[0]
		}
		if err, field := c.findField(value, paramKey); err == nil {
			kind := field.Kind()

			var v interface{}
			// Need handle more case of kind
			if kind == reflect.Int {
				v, _ = strconv.Atoi(paramValue)
			} else if kind == reflect.Int64 {
				v, _ = strconv.ParseInt(paramValue, 10, 64)
			} else if kind == reflect.String {
				v = paramValue
			} else if kind == reflect.Slice {
				sliceKind := reflect.TypeOf(field.Interface()).Elem().Kind()
				if sliceKind == reflect.String {
					v = strings.Split(paramValue, ",")
				} else {
					log.Println("Unhandled slice kind:", kind)
					continue
				}
			} else if kind == reflect.Struct {
				newModel := reflect.New(reflect.Indirect(field).Type()).Interface()
				if errDecode := json.Unmarshal([]byte(paramValue), newModel); errDecode != nil {
					panic(errDecode)
				}
				v = newModel
			} else {
				log.Println("Unhandled kind:", kind)
				continue
			}
			field.Set(reflect.Indirect(reflect.ValueOf(v)))
		} else {
			log.Println(err)
		}
	}
	return searchModel
}

func (c *SearchHandler) findField(value reflect.Value, paramKey string) (error, reflect.Value) {
	if index, ok := c.searchModelParamIndex[paramKey]; ok {
		searchModelField := value.Field(c.searchModelIndex)
		if searchModelField.Kind() == reflect.Ptr {
			searchModelField = reflect.Indirect(searchModelField)
		}
		return nil, searchModelField.Field(index)
	} else if index, ok := c.paramIndex[paramKey]; ok {
		return nil, value.Field(index)
	}
	return errors.New("can't find field " + paramKey), value
}

func (c *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	var searchModel = c.CreateSearchModelObject()

	method := r.Method
	x := 1
	if method == http.MethodGet {
		ps := r.URL.Query()
		fs := ps.Get("fields")
		if len(fs) == 0 {
			x = -1
		}
		c.mapParamsToSearchModel(searchModel, ps)
	} else if method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&searchModel); err != nil {
			http.Error(w, "cannot decode search model: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	userId := ""
	if len(c.userId) == 0 {
		u := r.Context().Value(c.userId)
		if u != nil {
			u2, ok2 := u.(string)
			if ok2 {
				userId = u2
			}
		}
	}
	c.ProcessSearchModel(searchModel, userId)

	result, err := c.searchService.Search(r.Context(), searchModel)
	if err != nil {
		Respond(w, r, http.StatusInternalServerError, InternalServerError, c.LogWriter, c.Resource, "Reject", false, err.Error())
	} else {
		if x == -1 {
			Succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, "Search")
		} else if c.quickSearch && x == 1 {
			value := reflect.Indirect(reflect.ValueOf(searchModel))
			numField := value.NumField()
			for i := 0; i < numField; i++ {
				field := value.Field(i)
				interfaceOfField := field.Interface()
				if v, ok := interfaceOfField.(*SearchModel); ok {
					if len(v.Fields) > 0 {
						result1 := ToCsv(interfaceOfField, result, c.embedField)
						Succeed(w, r, http.StatusOK, result1, c.LogWriter, c.Resource, "Search")
						return
					}
				}
			}
			Succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, "Search")
			// Error(w, r, http.StatusBadRequest, errors.New("Bad request"), c.LogWriter, c.Resource, "Search")
		} else {
			Succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, "Search")
		}
	}
}
