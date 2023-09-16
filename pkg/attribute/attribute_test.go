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
	})
})
