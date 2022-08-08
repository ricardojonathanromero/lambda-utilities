package utils

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
)

// ToString converts interface to string/*
func ToString(i interface{}) string {
	m, _ := json.Marshal(i)
	return string(m)
}

// StringToInt converts string to int64/*
func StringToInt(value string) int64 {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return int64(v)
}

// EncodeStr encodes string to base64/*
func EncodeStr(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// ToDynamoDBMap converts interface to map[string]*dynamodb.Attribute.
// Just use this function when your struct has configured json tags/*
func ToDynamoDBMap(i interface{}) map[string]types.AttributeValue {
	v, _ := attributevalue.MarshalMap(i)
	return v
}

// DynamoListToInterface unmarshal items map to model based on dynamoav tags/*
func DynamoListToInterface(items []map[string]types.AttributeValue, i interface{}) {
	_ = attributevalue.UnmarshalListOfMaps(items, &i)
}

// DynamoMapToInterface unmarshal items map to model based on dynamoav tags/*
func DynamoMapToInterface(item map[string]types.AttributeValue, i interface{}) {
	_ = attributevalue.UnmarshalMap(item, &i)
}

// CopyStruct copy source to target based on json tags/*
func CopyStruct(source, target interface{}) {
	s, _ := json.Marshal(source)
	_ = json.Unmarshal(s, &target)
}

// StringToStruct parse string to struct/*
func StringToStruct(s string, i interface{}) {
	_ = json.Unmarshal([]byte(s), &i)
}
