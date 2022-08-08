package dynamodb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ricardojonathanromero/lambda-utilities/utils"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	localEnv  = "local"
	region    = "us-east-1"
	timeoutIn = 10 * time.Second
)

func localDynamoDB(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: utils.GetEnv("DB_URI", "")}, nil
		})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
	)

	if err != nil {
		log.Errorf("error connecting to DynamoDB: %v", err)
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func NewSess() (*dynamodb.Client, error) {
	ctx, cancel := utils.NewContextWithTimeout(timeoutIn)
	defer cancel()

	if strings.EqualFold(utils.GetEnv("ENV", localEnv), localEnv) {
		return localDynamoDB(ctx)
	}

	// config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(utils.GetEnv("AWS_DEFAULT_REGION", region)))
	if err != nil {
		log.Errorf("unable to load SDK config, %v", err)
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}
