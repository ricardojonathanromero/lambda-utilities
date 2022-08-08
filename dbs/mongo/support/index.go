package support

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ricardojonathanromero/lambda-utilities/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io"
	"os"
	"time"
)

type IMongo interface {
	StartMongoDBLocal() error
	PauseContainer()
	ReRunContainer()
	ShutDownMongoLocal()
	GetMongoDBClient() *mongo.Client
}

type mongoSup struct {
	suite.Suite
	mongoURI string
}

const (
	mongoImage string = "mongo:latest"
	timeoutIn         = 30 * time.Second
)

var _ IMongo = (*mongoSup)(nil)

func (suite *mongoSup) StartMongoDBLocal() error {

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
			if mongoImage == repoTag {
				imageExist = true
				break
			}
		}
		if imageExist {
			break
		}
	}

	if !imageExist {
		out, err := cli.ImagePull(ctx, mongoImage, types.ImagePullOptions{})

		if err != nil {
			return err
		}
		_, _ = io.Copy(os.Stdout, out)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        mongoImage,
		ExposedPorts: nat.PortSet{"27017": struct{}{}},
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{"27017": {{HostIP: "0.0.0.0", HostPort: "27017"}}},
	}, nil, nil, "mongolocal")
	if err != nil {
		return err
	}

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	r, _ := cli.ContainerInspect(ctx, resp.ID)
	_ = r.Config

	if os.Getenv("CI") == "CI" {
		suite.mongoURI = "mongodb://" + r.NetworkSettings.IPAddress + ":27017/"
	} else {
		suite.mongoURI = "mongodb://localhost:27017/"
	}
	return nil
}

func (suite *mongoSup) PauseContainer() {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	ctx := context.Background()
	if err != nil {
		panic(err)
	}

	err = cli.ContainerPause(ctx, "mongolocal")
	if err != nil {
		log.Printf("container no active: %v", err)
	}
}

func (suite *mongoSup) ReRunContainer() {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	ctx := context.Background()
	if err != nil {
		panic(err)
	}

	err = cli.ContainerUnpause(ctx, "mongolocal")
	if err != nil {
		log.Printf("container no active: %v", err)
	}
}

func (suite *mongoSup) ShutDownMongoLocal() {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	ctx := context.Background()
	if err != nil {
		panic(err)
	}

	if err = cli.ContainerStop(ctx, "mongolocal", nil); err != nil {
		log.Printf("Unable to stop container %s: %s", "mongolocal", err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err = cli.ContainerRemove(ctx, "mongolocal", removeOptions); err != nil {
		fmt.Printf("Unable to remove container: %v\n", err)
	}
}

func (suite *mongoSup) GetMongoDBClient() *mongo.Client {
	// create context with timeout to validate db connection
	ctx, cancel := utils.NewContextWithTimeout(timeoutIn)
	defer cancel()

	// create client connection
	log.Infof("conencting to %s", suite.mongoURI)
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(suite.mongoURI))
	if err != nil {
		log.Fatalf("error connecting to db. reason\n%v", err)
	}

	// confirm connection making ping to db server
	log.Info("doing ping to db")
	if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		_ = mongoClient.Disconnect(context.Background())
		log.Fatalf("error ping to db. reason \n%v", err)
	}

	return mongoClient
}

func New() IMongo {
	return &mongoSup{}
}
