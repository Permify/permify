package bundle

import (
	"testing"

	"google.golang.org/protobuf/types/known/anypb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestTuple -
func TestBundle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bundle-suite")
}

var _ = Describe("bundle", func() {
	Context("RelationshipOperation", func() {
		It("should process relationship operations correctly", func() {
			tests := []struct {
				arguments map[string]string
				operation *base.Operation
			}{
				{
					arguments: map[string]string{
						"organizationID": "123",
						"platformID":     "32",
						"userID":         "165",
					},
					operation: &base.Operation{
						RelationshipsWrite: []string{
							"organization:{{.organizationID}}#admin@user:{{.userID}}",
							"platform:{{.platformID}}#member@user:{{.userID}}",
						},
						RelationshipsDelete: []string{
							"platform:{{.platformID}}#admin@user:{{.userID}}",
						},
					},
				},
			}

			for _, tt := range tests {
				tb, _, err := Operation(tt.arguments, tt.operation)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(tb.Write.GetTuples()).Should(Equal([]*base.Tuple{
					{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "123",
						},
						Relation: "admin",
						Subject: &base.Subject{
							Type: "user",
							Id:   "165",
						},
					},
					{
						Entity: &base.Entity{
							Type: "platform",
							Id:   "32",
						},
						Relation: "member",
						Subject: &base.Subject{
							Type: "user",
							Id:   "165",
						},
					},
				}))

				Expect(tb.Delete.GetTuples()).Should(Equal([]*base.Tuple{
					{
						Entity: &base.Entity{
							Type: "platform",
							Id:   "32",
						},
						Relation: "admin",
						Subject: &base.Subject{
							Type: "user",
							Id:   "165",
						},
					},
				}))
			}
		})

		It("should get invalid entity error", func() {
			tests := []struct {
				arguments map[string]string
				operation *base.Operation
			}{
				{
					arguments: map[string]string{
						"organization": "123",
						"platformID":   "32",
						"userID":       "165",
					},
					operation: &base.Operation{
						RelationshipsWrite: []string{
							"organization:{{.organizationID}}#admin@user:{{.userID}}",
							"platform:{{.platformID}}#member@user:{{.userID}}",
						},
						RelationshipsDelete: []string{
							"platform:{{.platformID}}#admin@user:{{.userID}}",
						},
					},
				},
			}

			for _, tt := range tests {
				_, _, err := Operation(tt.arguments, tt.operation)
				Expect(err.Error()).Should(Equal("invalid entity"))
			}
		})
	})

	Context("AttributeOperation", func() {
		It("should process attribute operations correctly", func() {
			tests := []struct {
				arguments map[string]string
				operation *base.Operation
			}{
				{
					arguments: map[string]string{
						"organizationID": "123",
						"platformID":     "32",
						"userID":         "165",
					},
					operation: &base.Operation{
						AttributesWrite: []string{
							"organization:{{.organizationID}}$public|boolean:true",
						},
						AttributesDelete: []string{
							"platform:{{.platformID}}$balance|integer:405",
						},
					},
				},
			}

			isPublic, _ := anypb.New(&base.BooleanValue{Data: true})
			integerValue, _ := anypb.New(&base.IntegerValue{Data: 405})

			for _, tt := range tests {
				_, ab, err := Operation(tt.arguments, tt.operation)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(ab.Write.GetAttributes()).Should(Equal([]*base.Attribute{
					{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "123",
						},
						Attribute: "public",
						Value:     isPublic,
					},
				}))

				Expect(ab.Delete.GetAttributes()).Should(Equal([]*base.Attribute{
					{
						Entity: &base.Entity{
							Type: "platform",
							Id:   "32",
						},
						Attribute: "balance",
						Value:     integerValue,
					},
				}))
			}
		})

		It("should get invalid attribute error", func() {
			tests := []struct {
				arguments map[string]string
				operation *base.Operation
			}{
				{
					arguments: map[string]string{
						"organizationID": "123",
						"ID":             "32",
						"userID":         "165",
					},
					operation: &base.Operation{
						AttributesWrite: []string{
							"organization:{{.organizationID}}$public|boolean:true",
						},
						AttributesDelete: []string{
							"platform:{{.platformID}}$balance|integer:405",
						},
					},
				},
			}

			for _, tt := range tests {
				_, _, err := Operation(tt.arguments, tt.operation)
				Expect(err.Error()).Should(Equal("invalid attribute"))
			}
		})
	})
})
