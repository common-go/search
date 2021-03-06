package search

import (
	"reflect"
)

type SearchModel struct {
	PageIndex     int64 `mapstructure:"page_index" json:"pageIndex,omitempty" gorm:"column:pageindex" bson:"pageIndex,omitempty" dynamodbav:"pageIndex,omitempty" firestore:"pageIndex,omitempty"`
	PageSize      int64 `mapstructure:"page_size" json:"pageSize,omitempty" gorm:"column:pagesize" bson:"pageSize,omitempty" dynamodbav:"pageSize,omitempty" firestore:"pageSize,omitempty"`
	FirstPageSize int64 `mapstructure:"first_page_size" json:"firstPageSize,omitempty" gorm:"column:firstpagesize" bson:"firstPageSize,omitempty" dynamodbav:"firstPageSize,omitempty" firestore:"firstPageSize,omitempty"`

	Page          int64                    `mapstructure:"page" json:"page,omitempty" gorm:"column:pageindex" bson:"page,omitempty" dynamodbav:"page,omitempty" firestore:"page,omitempty"`
	Limit         int64                    `mapstructure:"limit" json:"limit,omitempty" gorm:"column:limit" bson:"limit,omitempty" dynamodbav:"limit,omitempty" firestore:"limit,omitempty"`
	FirstLimit    int64                    `mapstructure:"first_limit" json:"firstLimit,omitempty" gorm:"column:firstlimit" bson:"firstLimit,omitempty" dynamodbav:"firstLimit,omitempty" firestore:"firstLimit,omitempty"`
	Fields        []string                 `mapstructure:"fields" json:"fields,omitempty" gorm:"column:fields" bson:"fields,omitempty" dynamodbav:"fields,omitempty" firestore:"fields,omitempty"`
	Sort          string                   `mapstructure:"sort" json:"sort,omitempty" gorm:"column:sortfield" bson:"sort,omitempty" dynamodbav:"sort,omitempty" firestore:"sort,omitempty"`
	CurrentUserId string                   `mapstructure:"current_user_id" json:"currentUserId,omitempty" gorm:"column:currentuserid" bson:"currentUserId,omitempty" dynamodbav:"currentUserId,omitempty" firestore:"currentUserId,omitempty"`
	Keyword       string                   `mapstructure:"keyword" json:"keyword,omitempty" gorm:"column:keyword" bson:"keyword,omitempty" dynamodbav:"keyword,omitempty" firestore:"keyword,omitempty"`
	Excluding     map[string][]interface{} `mapstructure:"excluding" json:"excluding,omitempty" gorm:"column:excluding" bson:"excluding,omitempty" dynamodbav:"excluding,omitempty" firestore:"excluding,omitempty"`
	RefId         string                   `mapstructure:"refid" json:"refId,omitempty" gorm:"column:refid" bson:"refId,omitempty" dynamodbav:"refId,omitempty" firestore:"refId,omitempty"`
}

func IsExtendedFromSearchModel(searchModelType reflect.Type) bool {
	var searchModel = reflect.New(searchModelType).Interface()
	if _, ok := searchModel.(*SearchModel); ok {
		return false
	} else {
		value := reflect.Indirect(reflect.ValueOf(searchModel))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if _, ok := value.Field(i).Interface().(*SearchModel); ok {
				return true
			}
		}
	}
	return false
}
func GetSearchModel(m interface{}) *SearchModel {
	if sModel, ok := m.(*SearchModel); ok {
		return sModel
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*SearchModel); ok {
				return sModel1
			}
		}
	}
	return nil
}
