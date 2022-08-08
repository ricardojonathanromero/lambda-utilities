package mongo

import (
	"context"
	"errors"
	"github.com/ricardojonathanromero/lambda-utilities/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const timeoutIn = 10 * time.Second

var (
	InvalidMongoDBConn = errors.New("connection to db not initialized correctly")
	DBNoConnected      = errors.New("ping to db no generate any response")
)

// NewConn creates a new mongodb connection/*
func NewConn() (*mongo.Client, error) {
	// get environment variable
	uri := utils.GetEnv("DB_URI", "")
	// create context with timeout to validate db connection
	ctx, cancel := utils.NewContextWithTimeout(timeoutIn)
	defer cancel()

	// create client connection
	log.Infof("conencting to %s", uri)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Errorf("error connecting to db. reason\n%v", err)
		return client, InvalidMongoDBConn
	}

	// confirm connection making ping to db server
	log.Info("doing ping to db")
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		log.Errorf("error ping to db. reason \n%v", err)
		return client, DBNoConnected
	}

	return client, nil
}
