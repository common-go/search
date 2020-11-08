package search

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type SearchHandler struct {
	searchService             SearchService
	searchModelType           reflect.Type
	LogError                  func(context.Context, string)
	LogWriter                 SearchLogWriter
	Config                    SearchResultConfig
	quickSearch               bool
	isExtendedSearchModelType bool
	Resource                  string
	Action                    string
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
	UserId             = "userId"
	Uid                = "uid"
	Username           = "username"
	Search             = "search"
)

func NewSearchHandler(searchService SearchService, searchModelType reflect.Type, resource string, logError func(context.Context, string), logService SearchLogWriter) *SearchHandler {
	return NewSearchHandlerWithParameters(searchService, searchModelType, logError, nil, logService, true, resource, Search, UserId, "")
}
func NewSearchHandlerWithUserId(searchService SearchService, searchModelType reflect.Type, resource string, logError func(context.Context, string), logService SearchLogWriter, userId string) *SearchHandler {
	return NewSearchHandlerWithParameters(searchService, searchModelType, logError, nil, logService, true, resource, Search, userId, "")
}
func NewJSONSearchHandler(searchService SearchService, searchModelType reflect.Type, resource string, logError func(context.Context, string), logService SearchLogWriter) *SearchHandler {
	return NewSearchHandlerWithParameters(searchService, searchModelType, logError, nil, logService, false, resource, Search, UserId, "")
}
func NewDefaultSearchHandler(searchService SearchService, searchModelType reflect.Type, resource string, logError func(context.Context, string), logService SearchLogWriter, quickSearch bool, userId string) *SearchHandler {
	return NewSearchHandlerWithParameters(searchService, searchModelType, logError, nil, logService, quickSearch, resource, Search, userId, "")
}
func NewSearchHandlerWithParameters(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), config *SearchResultConfig, logService SearchLogWriter, quickSearch bool, resource string, action string, userId string, embedField string) *SearchHandler {
	var c SearchResultConfig
	if len(action) == 0 {
		action = Search
	}
	if config != nil {
		c = *config
	} else {
		c.LastPage = "last"
		c.Results = "results"
		c.Total = "total"
	}
	isExtendedSearchModelType := IsExtendedFromSearchModel(searchModelType)
	if isExtendedSearchModelType == false {
		panic(errors.New(searchModelType.Name() + " isn't SearchModel struct nor extended from SearchModel struct!"))
	}

	paramIndex := BuildParamIndex(searchModelType)
	searchModelParamIndex := BuildParamIndex(reflect.TypeOf(SearchModel{}))
	searchModelIndex := FindSearchModelIndex(searchModelType)

	return &SearchHandler{searchService: searchService, searchModelType: searchModelType, Config: c, LogWriter: logService, quickSearch: quickSearch, isExtendedSearchModelType: isExtendedSearchModelType, Resource: resource, Action: action, paramIndex: paramIndex, searchModelIndex: searchModelIndex, searchModelParamIndex: searchModelParamIndex, userId: userId, embedField: embedField, LogError: logError}
}

func BuildSearchModel(r *http.Request, searchModelType reflect.Type, isExtendedSearchModelType bool, userIdName string, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) (interface{}, int, error) {
	var searchModel = CreateSearchModelObject(searchModelType, isExtendedSearchModelType)
	method := r.Method
	x := 1
	if method == http.MethodGet {
		ps := r.URL.Query()
		fs := ps.Get("fields")
		if len(fs) == 0 {
			x = -1
		}
		MapParamsToSearchModel(searchModel, ps, searchModelParamIndex, searchModelIndex, paramIndex)
	} else if method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&searchModel); err != nil {
			return nil, x, err
		}
	}
	userId := ""
	if len(userId) == 0 {
		u := r.Context().Value(userIdName)
		if u != nil {
			u2, ok2 := u.(string)
			if ok2 {
				userId = u2
			}
		}
	}
	ProcessSearchModel(searchModel, userId)
	return searchModel, x, nil
}
func BuildResultMap(models interface{}, count int64, m *SearchModel, config SearchResultConfig) (map[string]interface{}, bool) {
	result := make(map[string]interface{})
	isLastPage := IsLastPage(models, count, m.PageIndex, m.PageSize, m.FirstPageSize)

	result[config.Total] = count
	if isLastPage {
		result[config.LastPage] = isLastPage
	}
	result[config.Results] = models
	return result, isLastPage
}
func ResultToCsv(searchModel interface{}, m *SearchModel, models interface{}, count int64, isLastPage bool, embedField string) (string, bool) {
	value := reflect.Indirect(reflect.ValueOf(searchModel))
	numField := value.NumField()
	for i := 0; i < numField; i++ {
		field := value.Field(i)
		interfaceOfField := field.Interface()
		if v, ok := interfaceOfField.(*SearchModel); ok {
			if len(v.Fields) > 0 {
				result1 := ToCsv(*m, models, count, isLastPage, embedField)
				return result1, true
			}
		}
	}
	return "", false
}
func BuildParamIndex(searchModelType reflect.Type) map[string]int {
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

func FindSearchModelIndex(searchModelType reflect.Type) int {
	numField := searchModelType.NumField()
	for i := 0; i < numField; i++ {
		if searchModelType.Field(i).Type == reflect.TypeOf(&SearchModel{}) {
			return i
		}
	}
	return -1
}

// Check valid and change value of pagination to correct
func RepairSearchModel(searchModel *SearchModel, currentUserId string) {
	searchModel.CurrentUserId = currentUserId

	pageSize := searchModel.PageSize
	if pageSize > MaxPageSizeDefault || pageSize < 1 {
		pageSize = PageSizeDefault
	}
	pageIndex := searchModel.PageIndex
	if searchModel.PageIndex < 1 {
		pageIndex = 1
	}

	if searchModel.PageSize != pageSize {
		searchModel.PageSize = pageSize
	}

	if searchModel.PageIndex != pageIndex {
		searchModel.PageIndex = pageIndex
	}
}

func ProcessSearchModel(sm interface{}, currentUserId string) {
	if s, ok := sm.(*SearchModel); ok { // Is SearchModel struct
		RepairSearchModel(s, currentUserId)
	} else { // Is extended from SearchModel struct
		value := reflect.Indirect(reflect.ValueOf(sm))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find SearchModel field of extended struct
			if s, ok := value.Field(i).Interface().(*SearchModel); ok {
				RepairSearchModel(s, currentUserId)
				break
			}
		}
	}
}
func CreateSearchModelObject(searchModelType reflect.Type, isExtendedSearchModelType bool) interface{} {
	var searchModel = reflect.New(searchModelType).Interface()
	if isExtendedSearchModelType {
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
func MapParamsToSearchModel(searchModel interface{}, params url.Values, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) interface{} {
	value := reflect.Indirect(reflect.ValueOf(searchModel))
	if value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	for paramKey, valueArr := range params {
		paramValue := ""
		if len(valueArr) > 0 {
			paramValue = valueArr[0]
		}
		if err, field := FindField(value, paramKey, searchModelParamIndex, searchModelIndex, paramIndex); err == nil {
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

func FindField(value reflect.Value, paramKey string, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) (error, reflect.Value) {
	if index, ok := searchModelParamIndex[paramKey]; ok {
		searchModelField := value.Field(searchModelIndex)
		if searchModelField.Kind() == reflect.Ptr {
			searchModelField = reflect.Indirect(searchModelField)
		}
		return nil, searchModelField.Field(index)
	} else if index, ok := paramIndex[paramKey]; ok {
		return nil, value.Field(index)
	}
	return errors.New("can't find field " + paramKey), value
}
