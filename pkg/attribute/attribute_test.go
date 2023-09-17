package attribute

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestAttribute -
func TestAttribute(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "attribute-suite")
}

var _ = Describe("attribute", func() {
	Context("Attribute", func() {
		isPublic, _ := anypb.New(&base.BooleanValue{Data: true})
		double, _ := anypb.New(&base.DoubleValue{Data: 100})
		integer, _ := anypb.New(&base.IntegerValue{Data: 45})

		It("ToString", func() {
			tests := []struct {
				target *base.Attribute
				str    string
			}{
				{
					target: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "is_public",
						Value:     isPublic,
					},
					str: "organization:1$is_public|boolean:true",
				},
				{
					target: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "balance",
						Value:     double,
					},
					str: "organization:1$balance|double:100",
				},
			}

			for _, tt := range tests {
				Expect(ToString(tt.target)).Should(Equal(tt.str))
			}
		})

		It("Attribute", func() {
			tests := []struct {
				target    string
				attribute *base.Attribute
			}{
				{
					target: "organization:1$is_public|boolean:true",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "is_public",
						Value:     isPublic,
					},
				},
				{
					target: "organization:1$balance|double:100",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "balance",
						Value:     double,
					},
				},
				{
					target: "user:1$age|integer:45",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "user",
							Id:   "1",
						},
						Attribute: "age",
						Value:     integer,
					},
				},
			}

			for _, tt := range tests {
				Expect(Attribute(tt.target)).Should(Equal(tt.attribute))
			}
		})

		It("EntityAndCallOrAttributeToString", func() {
			tests := []struct {
				entity          *base.Entity
				attributeOrCall string
				arguments       []*base.Argument
				result          string
			}{
				{
					entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					attributeOrCall: "check_credit",
					arguments: []*base.Argument{
						{
							Type: &base.Argument_ComputedAttribute{
								ComputedAttribute: &base.ComputedAttribute{
									Name: "credit",
								},
							},
						},
					},
					result: "organization:1$check_credit(credit)",
				},
				{
					entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					attributeOrCall: "is_public",
					arguments:       nil,
					result:          "organization:1$is_public",
				},
				{
					entity: &base.Entity{
						Type: "organization",
						Id:   "877",
					},
					attributeOrCall: "check_balance",
					arguments: []*base.Argument{
						{
							Type: &base.Argument_ContextAttribute{
								ContextAttribute: &base.ContextAttribute{
									Name: "amount",
								},
							},
						},
						{
							Type: &base.Argument_ComputedAttribute{
								ComputedAttribute: &base.ComputedAttribute{
									Name: "balance",
								},
							},
						},
					},
					result: "organization:877$check_balance(request.amount,balance)",
				},
			}

			for _, tt := range tests {
				Expect(EntityAndCallOrAttributeToString(tt.entity, tt.attributeOrCall, tt.arguments...)).Should(Equal(tt.result))
			}
		})

		It("EntityToString", func() {
			tests := []struct {
				entity *base.Entity
				result string
			}{
				{
					entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					result: "organization:1",
				},
				{
					entity: &base.Entity{
						Type: "repository",
						Id:   "abc",
					},
					result: "repository:abc",
				},
			}

			for _, tt := range tests {
				Expect(EntityToString(tt.entity)).Should(Equal(tt.result))
			}
		})
	})
})
