package utils_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ricardojonathanromero/lambda-serverless-example/api-gateway-example/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func doSomethingCool(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("timed out")
			return
		default:
			log.Info("doing something cool")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

var _ = Describe("context unit tests", func() {
	Context("happy path", func() {
		It("create context", func() {
			ctx, cancel := utils.NewContextWithTimeout(2)
			defer cancel()
			go doSomethingCool(ctx)
			select {
			case <-ctx.Done():
				log.Info("context done by timeout")
			}

			time.Sleep(4 * time.Second)
		})
	})
})

var _ = Describe("env unit tests", func() {
	Context("happy path", func() {
		BeforeEach(func() {
			_ = os.Setenv("SOME_VALUE", "test")
		})
		AfterEach(func() {
			_ = os.Remove("SOME_VALUE")
		})

		It("when env value is retrieved successfully", func() {
			val := utils.GetEnv("SOME_VALUE", "")
			Ω(val).To(Equal("test"))
		})

		It("when env value is retrieved by default", func() {
			val := utils.GetEnv("EXAMPLE", "default")
			Ω(val).To(Equal("default"))
		})
	})
})

var _ = Describe("converter unit tests", func() {
	type s struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	type t struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	type dynamoModel struct {
		Name  string `json:"name" dynamodbav:"Name"`
		Value string `json:"value" dynamodbav:"Value"`
	}

	Context("happy path", func() {
		It("convert interface to string", func() {
			source := s{Name: "example", Value: "example"}
			val := utils.ToString(source)
			Ω(val).To(Equal(`{"name":"example","value":"example"}`))
		})
		It("convert string to integer", func() {
			val := utils.StringToInt("10")
			Ω(val).To(Equal(int64(10)))
		})
		It("convert string to integer is zero", func() {
			val := utils.StringToInt("value")
			Ω(val).To(Equal(int64(0)))
		})
		It("convert encode string", func() {
			val := utils.EncodeStr("hello world")
			Ω(val).To(Equal("aGVsbG8gd29ybGQ="))
		})
		It("convert dynamodb to map", func() {
			dm := dynamoModel{Name: "example", Value: "example"}
			val := utils.ToDynamoDBMap(dm)
			Ω(val).To(Equal(map[string]types.AttributeValue{
				"Name":  &types.AttributeValueMemberS{Value: "example"},
				"Value": &types.AttributeValueMemberS{Value: "example"},
			}))
		})
		It("convert dynamodb list to interface", func() {
			target := make([]*dynamoModel, 0)
			source := []map[string]types.AttributeValue{
				{
					"Name":  &types.AttributeValueMemberS{Value: "example"},
					"Value": &types.AttributeValueMemberS{Value: "example"},
				},
			}
			utils.DynamoListToInterface(source, &target)
			log.Info(target)
			Ω(target).NotTo(BeNil())
			Ω(len(target)).To(Equal(1))
			Ω(target[0].Name).To(Equal("example"))
			Ω(target[0].Value).To(Equal("example"))
		})
		It("convert dynamodb map to interface", func() {
			var target *dynamoModel
			source := map[string]types.AttributeValue{
				"Name":  &types.AttributeValueMemberS{Value: "example"},
				"Value": &types.AttributeValueMemberS{Value: "example"},
			}
			utils.DynamoMapToInterface(source, &target)
			log.Info(target)
			Ω(target).NotTo(BeNil())
			Ω(target.Name).To(Equal("example"))
			Ω(target.Value).To(Equal("example"))
		})
		It("copy struct", func() {
			var target *t
			source := &s{Name: "example", Value: "example"}

			utils.CopyStruct(source, &target)
			Ω(target).NotTo(BeNil())
			Ω(target.Name).To(Equal("example"))
			Ω(target.Value).To(Equal("example"))
		})
		It("convert string to struct", func() {
			var target *t
			source := `{"name":"example","value":"example"}`
			utils.StringToStruct(source, &target)
			Ω(target).NotTo(BeNil())
			Ω(target.Name).To(Equal("example"))
			Ω(target.Value).To(Equal("example"))
		})
	})
})
