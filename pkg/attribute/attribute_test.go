package attribute

import (
	"testing"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestAttribute -
func TestAttribute(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "attribute-suite")
}

var _ = Describe("attribute", func() {
	Context("Attribute", func() {
		isPublic, _ := anypb.New(wrapperspb.Bool(true))
		double, _ := anypb.New(wrapperspb.Double(100))

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
						Type:      "boolean",
						Value:     isPublic,
					},
					str: "organization:1#is_public@boolean:true",
				},
				{
					target: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "balance",
						Type:      "double",
						Value:     double,
					},
					str: "organization:1#balance@double:100",
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
					target: "organization:1#is_public@boolean:true",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "is_public",
						Type:      "boolean",
						Value:     isPublic,
					},
				},
				{
					target: "organization:1#balance@double:100",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "balance",
						Type:      "double",
						Value:     double,
					},
				},
			}

			for _, tt := range tests {
				Expect(Attribute(tt.target)).Should(Equal(tt.attribute))
			}
		})
	})
})
