package support

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	at "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ricardojonathanromero/lambda-utilities/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"time"
)

type IDynamoSupport interface {
	StartDynamoLocal() error
	CreateTable(input *dynamodb.CreateTableInput) error
	GetDynamoDBLocalClient(ctx context.Context) *dynamodb.Client
	PutItem(tableName string, item interface{}) error
	DeleteItem(tableName string, key string, value string) error
	PauseContainer()
	ReRunContainer()
	ShutDownDynamoLocal()
}

type dynamoSup struct {
	suite.Suite
	url string
}

const (
	dynamodbLocalImage string = "amazon/dynamodb-local:latest"
	timeoutIn                 = 30 * time.Second
)

var _ IDynamoSupport = (*dynamoSup)(nil)

// StartDynamoLocal creates a new dynamodb docker container
func (suite *dynamoSup) StartDynamoLocal() error {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	ctx := context.Background()
	listImages, err := cli.ImageList(ctx, types.ImageListOptions{All: false})
	if err != nil {
		return err
	}

	imageExist := false
	for _, image := range listImages {
		for _, repoTag := range image.RepoTags {
			if dynamodbLocalImage == repoTag {
				imageExist = true
				break
			}
		}
		if imageExist {
			break
		}
	}

	if !imageExist {
		out, err := cli.ImagePull(ctx, dynamodbLocalImage, types.ImagePullOptions{})

		if err != nil {
			return err
		}
		_, _ = io.Copy(os.Stdout, out)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        dynamodbLocalImage,
		ExposedPorts: nat.PortSet{"8000": struct{}{}},
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{"8000": {{HostIP: "0.0.0.0", HostPort: "8000"}}},
	}, nil, nil, "dynamodblocal")
	if err != nil {
		return err
	}

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	r, _ := cli.ContainerInspect(ctx, resp.ID)
	_ = r.Config

	if os.Getenv("CI") == "CI" {
		suite.url = "http://" + r.NetworkSettings.IPAddress + ":8000"
	} else {
		suite.url = "http://" + "localhost" + ":8000"
	}
	return nil
}

func (suite *dynamoSup) CreateTable(input *dynamodb.CreateTableInput) error {
	ctx, cancel := utils.NewContextWithTimeout(timeoutIn)
	defer cancel()

	_, err1 := suite.GetDynamoDBLocalClient(ctx).CreateTable(ctx, input)
	if err1 != nil {
		return fmt.Errorf("%v\n", err1)
	}
	return nil
}

func (suite *dynamoSup) GetDynamoDBLocalClient(ctx context.Context) *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: suite.url}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
	)

	if err != nil {
		log.Fatalf("error connecting to DynamoDB: %v", err)
	}

	return dynamodb.NewFromConfig(cfg)
}

func (suite *dynamoSup) PutItem(tableName string, item interface{}) error {
	ctx, cancel := utils.NewContextWithTimeout(timeoutIn)
	defer cancel()

	values := utils.ToDynamoDBMap(item)
	putInput := &dynamodb.PutItemInput{
		Item:      values,
		TableName: aws.String(tableName),
	}
	_, err2 := suite.GetDynamoDBLocalClient(ctx).PutItem(ctx, putInput)
	if err2 != nil {
		return fmt.Errorf("%v\n", err2)
	}
	return nil
}

func (suite *dynamoSup) DeleteItem(tableName string, key string, value string) error {
	ctx, cancel := utils.NewContextWithTimeout(timeoutIn)
	defer cancel()

	input := &dynamodb.DeleteItemInput{
		Key: map[string]at.AttributeValue{
			key: &at.AttributeValueMemberS{Value: value},
		},
		TableName: aws.String(tableName),
	}
	_, err := suite.GetDynamoDBLocalClient(ctx).DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	return nil
}

func (suite *dynamoSup) PauseContainer() {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	ctx := context.Background()
	if err != nil {
		panic(err)
	}

	err = cli.ContainerPause(ctx, "dynamodblocal")
	if err != nil {
		log.Printf("container no active: %v", err)
	}
}

func (suite *dynamoSup) ReRunContainer() {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	ctx := context.Background()
	if err != nil {
		panic(err)
	}

	err = cli.ContainerUnpause(ctx, "dynamodblocal")
	if err != nil {
		log.Printf("container no active: %v", err)
	}
}

func (suite *dynamoSup) ShutDownDynamoLocal() {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	ctx := context.Background()
	if err != nil {
		panic(err)
	}

	if err = cli.ContainerStop(ctx, "dynamodblocal", nil); err != nil {
		log.Printf("Unable to stop container %s: %s", "dynamodblocal", err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err = cli.ContainerRemove(ctx, "dynamodblocal", removeOptions); err != nil {
		fmt.Printf("Unable to remove container: %v\n", err)
	}
}

// New constructor/*
func New() IDynamoSupport {
	return &dynamoSup{}
}
