package attribute

import (
	"errors"
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
	isPublic, _ := anypb.New(&base.BooleanValue{Data: true})
	boolArrayValue, _ := anypb.New(&base.BooleanArrayValue{Data: []bool{false, true}})
	stringValue, _ := anypb.New(&base.StringValue{Data: "string_value"})
	stringArrayValue, _ := anypb.New(&base.StringArrayValue{Data: []string{"127.0.0.1", "127.0.0.2"}})
	doubleValue, _ := anypb.New(&base.DoubleValue{Data: 100.01})
	doubleArrayValue, _ := anypb.New(&base.DoubleArrayValue{Data: []float64{100, 200}})
	integerValue, _ := anypb.New(&base.IntegerValue{Data: 45})
	integerArrayValue, _ := anypb.New(&base.IntegerArrayValue{Data: []int32{45, 55}})

	Context("Attribute", func() {
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
						Value:     doubleValue,
					},
					str: "organization:1$balance|double:100.01",
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
				error     error
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
					error: nil,
				},
				{
					target: "organization:1$is_public|boolean[]:false,true",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "is_public",
						Value:     boolArrayValue,
					},
					error: nil,
				},
				{
					target: "organization:1$val|string:string_value",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "val",
						Value:     stringValue,
					},
					error: nil,
				},
				{
					target: "organization:1$addresses|string[]:127.0.0.1,127.0.0.2",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "addresses",
						Value:     stringArrayValue,
					},
					error: nil,
				},
				{
					target: "organization:1$local|string:string_value",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "local",
						Value:     stringValue,
					},
					error: nil,
				},
				{
					target: "organization:1$balance|double:100.01",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "balance",
						Value:     doubleValue,
					},
					error: nil,
				},
				{
					target: "organization:1$is_public|boolean:asa",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "is_public",
						Value:     isPublic,
					},
					error: errors.New("failed to parse boolean: strconv.ParseBool: parsing \"asa\": invalid syntax"),
				},
				{
					target: "organization:1$is_public|boolean[]:asa",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "is_public",
						Value:     isPublic,
					},
					error: errors.New("failed to parse boolean: strconv.ParseBool: parsing \"asa\": invalid syntax"),
				},
				{
					target: "organization:1$balance|double:4eew",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "balance",
						Value:     doubleValue,
					},
					error: errors.New("failed to parse float: strconv.ParseFloat: parsing \"4eew\": invalid syntax"),
				},
				{
					target: "organization:1$balance|double[]:4eew",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "balance",
						Value:     doubleValue,
					},
					error: errors.New("failed to parse float: strconv.ParseFloat: parsing \"4eew\": invalid syntax"),
				},
				{
					target: "organization:1$balance|double[]:100,200",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "balance",
						Value:     doubleArrayValue,
					},
					error: nil,
				},
				{
					target: "organization:1$age|integer:4eew",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "age",
						Value:     doubleValue,
					},
					error: errors.New("failed to parse integer: strconv.ParseInt: parsing \"4eew\": invalid syntax"),
				},
				{
					target: "organization:1$age|integer[]:4eew",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Attribute: "age",
						Value:     doubleValue,
					},
					error: errors.New("failed to parse integer: strconv.ParseInt: parsing \"4eew\": invalid syntax"),
				},
				{
					target: "user:1$age|integer:45",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "user",
							Id:   "1",
						},
						Attribute: "age",
						Value:     integerValue,
					},
					error: nil,
				},
				{
					target: "user:1$ages|integer[]:45,55",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "user",
							Id:   "1",
						},
						Attribute: "ages",
						Value:     integerArrayValue,
					},
					error: nil,
				},
				{
					target: "user:1$age-integer:45",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "user",
							Id:   "1",
						},
						Attribute: "age",
						Value:     integerValue,
					},
					error: ErrInvalidAttribute,
				},
				{
					target: "user:1-age|integer:45",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "user",
							Id:   "1",
						},
						Attribute: "age",
						Value:     integerValue,
					},
					error: ErrInvalidEntity,
				},
				{
					target: "user:1$age|integer,45",
					attribute: &base.Attribute{
						Entity: &base.Entity{
							Type: "user",
							Id:   "1",
						},
						Attribute: "age",
						Value:     integerValue,
					},
					error: ErrInvalidAttribute,
				},
			}

			for _, tt := range tests {
				attr, err := Attribute(tt.target)
				if tt.error != nil {
					Expect(err.Error()).Should(Equal(tt.error.Error()))
				} else {
					Expect(attr).Should(Equal(tt.attribute))
				}
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
							Type: &base.Argument_ComputedAttribute{
								ComputedAttribute: &base.ComputedAttribute{
									Name: "balance",
								},
							},
						},
					},
					result: "organization:877$check_balance(balance)",
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

		It("TypeUrlToString", func() {
			tests := []struct {
				url    string
				result string
			}{
				{
					url:    "type.googleapis.com/base.v1.StringValue",
					result: "string",
				},
				{
					url:    "type.googleapis.com/base.v1.BooleanValue",
					result: "boolean",
				},
				{
					url:    "type.googleapis.com/base.v1.IntegerValue",
					result: "integer",
				},
				{
					url:    "type.googleapis.com/base.v1.DoubleValue",
					result: "double",
				},
				{
					url:    "type.googleapis.com/base.v1.StringArrayValue",
					result: "string[]",
				},
				{
					url:    "type.googleapis.com/base.v1.BooleanArrayValue",
					result: "boolean[]",
				},
				{
					url:    "type.googleapis.com/base.v1.IntegerArrayValue",
					result: "integer[]",
				},
				{
					url:    "type.googleapis.com/base.v1.DoubleArrayValue",
					result: "double[]",
				},
				{
					url:    "aa",
					result: "",
				},
			}

			for _, tt := range tests {
				Expect(TypeUrlToString(tt.url)).Should(Equal(tt.result))
			}
		})

		It("AnyToString", func() {
			tests := []struct {
				any    *anypb.Any
				result string
			}{
				{
					any:    isPublic,
					result: "true",
				},
				{
					any:    boolArrayValue,
					result: "false,true",
				},
				{
					any:    stringValue,
					result: "string_value",
				},
				{
					any:    stringArrayValue,
					result: "127.0.0.1,127.0.0.2",
				},
				{
					any:    doubleValue,
					result: "100.01",
				},
				{
					any:    doubleArrayValue,
					result: "100,200",
				},
				{
					any:    integerValue,
					result: "45",
				},
				{
					any:    integerArrayValue,
					result: "45,55",
				},
			}

			for _, tt := range tests {
				Expect(AnyToString(tt.any)).Should(Equal(tt.result))
			}
		})

		It("TypeToString", func() {
			tests := []struct {
				typ    base.AttributeType
				result string
			}{
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_INTEGER,
					result: "integer",
				},
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY,
					result: "integer[]",
				},
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_DOUBLE,
					result: "double",
				},
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY,
					result: "double[]",
				},
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_STRING,
					result: "string",
				},
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY,
					result: "string[]",
				},
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
					result: "boolean",
				},
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY,
					result: "boolean[]",
				},
				{
					typ:    base.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED,
					result: "undefined",
				},
			}

			for _, tt := range tests {
				Expect(TypeToString(tt.typ)).Should(Equal(tt.result))
			}
		})

		It("ValidateValue", func() {
			tests := []struct {
				any           *anypb.Any
				attributeType base.AttributeType
				err           error
			}{
				{
					any:           isPublic,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
					err:           nil,
				},
				{
					any:           boolArrayValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY,
					err:           nil,
				},
				{
					any:           stringValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_STRING,
					err:           nil,
				},
				{
					any:           stringArrayValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY,
					err:           nil,
				},
				{
					any:           doubleValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_DOUBLE,
					err:           nil,
				},
				{
					any:           doubleArrayValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY,
					err:           nil,
				},
				{
					any:           integerValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_INTEGER,
					err:           nil,
				},
				{
					any:           integerArrayValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY,
					err:           nil,
				},
				{
					any:           integerValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY,
					err:           errors.New("mismatched message type: got \"base.v1.IntegerArrayValue\", want \"base.v1.IntegerValue\""),
				},
				{
					any:           integerValue,
					attributeType: base.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED,
					err:           errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String()),
				},
			}

			for _, tt := range tests {
				err := ValidateValue(tt.any, tt.attributeType)
				if err != nil {
					Expect(err.Error()).Should(ContainSubstring(tt.err.Error()))
				}
			}
		})

		It("EntityAndAttributeToString", func() {
			tests := []struct {
				entity    *base.Entity
				attribute string
				result    string
			}{
				{
					entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					attribute: "is_public",
					result:    "repository:1$is_public",
				},
				{
					entity: &base.Entity{
						Type: "organization",
						Id:   "organization-879",
					},
					attribute: "credit",
					result:    "organization:organization-879$credit",
				},
			}

			for _, tt := range tests {
				result := EntityAndAttributeToString(tt.entity, tt.attribute)
				Expect(result).Should(Equal(tt.result))
			}
		})
	})
})
