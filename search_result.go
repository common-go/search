package search

type SearchResult struct {
	Results interface{} `mapstructure:"results" json:"results,omitempty" gorm:"column:results" bson:"results,omitempty" dynamodbav:"results,omitempty" firestore:"results,omitempty"`
	Total   int64       `mapstructure:"total" json:"total,omitempty" gorm:"column:total" bson:"total,omitempty" dynamodbav:"total,omitempty" firestore:"total,omitempty"`
	Last    bool        `mapstructure:"last" json:"last,omitempty" gorm:"column:last" bson:"last,omitempty" dynamodbav:"last,omitempty" firestore:"last,omitempty"`
}
