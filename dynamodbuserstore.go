package main

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DynamoDBUserStore struct {
	Client *dynamodb.DynamoDB
	Table  string
}

func (ddbus *DynamoDBUserStore) Create(entry *UserEntry) error {
	if _, err := ddbus.Client.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String(ddbus.Table),
		Item:                userToAttributes(entry),
		ConditionExpression: aws.String("attribute_not_exists(user)"),
	}); err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (ddbus *DynamoDBUserStore) Update(entry *UserEntry) error {
	if _, err := ddbus.Client.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String(ddbus.Table),
		Item:                userToAttributes(entry),
		ConditionExpression: aws.String("attribute_exists(user)"),
	}); err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

func (ddbus *DynamoDBUserStore) Get(user UserID) (*UserEntry, error) {
	rsp, err := ddbus.Client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(ddbus.Table),
		Key: map[string]*dynamodb.AttributeValue{
			"user": &dynamodb.AttributeValue{S: aws.String(string(user))},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return attributesToUser(rsp.Item), nil
}

func userToAttributes(entry *UserEntry) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"user": &dynamodb.AttributeValue{
			S: aws.String(string(entry.User)),
		},
		"email": &dynamodb.AttributeValue{S: aws.String(entry.Email)},
		"passwordHash": &dynamodb.AttributeValue{
			S: aws.String(base64.RawStdEncoding.EncodeToString(
				entry.PasswordHash,
			)),
		},
	}
}

func attributesToUser(attrs map[string]*dynamodb.AttributeValue) *UserEntry {
	user := UserID(*attrs["user"].S)
	email := *attrs["email"].S
	data, err := base64.RawStdEncoding.DecodeString(*attrs["passwordHash"].S)
	if err != nil {
		panic(fmt.Sprintf(
			"non-base64-encoded `passwordHash` value for user `%s`: %v",
			user,
			err,
		))
	}
	return &UserEntry{User: user, Email: email, PasswordHash: data}
}
