package search

import (
	"context"
	"errors"
	"reflect"
	"strings"
)

type SearchHandler struct {
	searchService             SearchService
	searchModelType           reflect.Type
	LogError                  func(context.Context, string)
	WriteLog                  func(ctx context.Context, resource string, action string, success bool, desc string) error
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

func NewSearchHandler(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(searchService, searchModelType, logError, writeLog, true, options...)
}
func NewJSONSearchHandler(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(searchService, searchModelType, logError, writeLog, true, options...)
}
func NewSearchHandlerWithQuickSearch(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
	var resource, action, user string
	if len(options) >= 1 {
		user = options[0]
	} else {
		user = UserId
	}
	if len(options) >= 2 {
		resource = options[1]
	} else {
		name := searchModelType.Name()
		if len(name) >= 3 && strings.HasSuffix(name, "SM") {
			name = name[0 : len(name)-2]
		}
		resource = BuildResourceName(name)
	}
	if len(options) >= 3 {
		action = options[2]
	} else {
		action = Search
	}
	return NewSearchHandlerWithConfig(searchService, searchModelType, logError, nil, writeLog, quickSearch, resource, action, user, "")
}
func NewSearchHandlerWithUserId(searchService SearchService, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(searchService, searchModelType, userId, logError, writeLog, true, options...)
}
func NewJSONSearchHandlerWithUserId(searchService SearchService, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(searchService, searchModelType, userId, logError, writeLog, false, options...)
}
func NewSearchHandlerWithUserIdAndQuickSearch(searchService SearchService, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
	var resource, action string
	if len(options) >= 1 {
		resource = options[0]
	} else {
		name := searchModelType.Name()
		if len(name) >= 3 && strings.HasSuffix(name, "SM") {
			name = name[0 : len(name)-2]
		}
		resource = BuildResourceName(name)
	}
	if len(options) >= 2 {
		action = options[1]
	} else {
		action = Search
	}
	return NewSearchHandlerWithConfig(searchService, searchModelType, logError, nil, writeLog, quickSearch, resource, action, userId, "")
}
func NewDefaultSearchHandler(searchService SearchService, searchModelType reflect.Type, resource string, logError func(context.Context, string), userId string, quickSearch bool, writeLog func(context.Context, string, string, bool, string) error) *SearchHandler {
	return NewSearchHandlerWithConfig(searchService, searchModelType, logError, nil, writeLog, quickSearch, resource, Search, userId, "")
}
func NewSearchHandlerWithConfig(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), config *SearchResultConfig, writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, resource string, action string, userId string, embedField string) *SearchHandler {
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

	return &SearchHandler{searchService: searchService, searchModelType: searchModelType, Config: c, WriteLog: writeLog, quickSearch: quickSearch, isExtendedSearchModelType: isExtendedSearchModelType, Resource: resource, Action: action, paramIndex: paramIndex, searchModelIndex: searchModelIndex, searchModelParamIndex: searchModelParamIndex, userId: userId, embedField: embedField, LogError: logError}
}
