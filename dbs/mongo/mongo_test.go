package mongo_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ricardojonathanromero/lambda-utilities/dbs/mongo"
)

var _ = Describe("unit tests", func() {
	Context("tests", func() {
		conn, err := mongo.NewConn()
		Ω(err).To(BeNil())
		Ω(conn).NotTo(BeNil())
	})
})
