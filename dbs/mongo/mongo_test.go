package mongo_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/ricardojonathanromero/lambda-serverless-example/api-gateway-example/infrastructure/mongo"
)

var _ = Describe("unit tests", func() {
	Context("tests", func() {
		_, err := mongo.NewConn()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("connected!")
		}
	})
})
