package engines

import (
	"errors"
	"reflect"

	"google.golang.org/protobuf/types/known/structpb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/Permify/permify/internal/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("utils", func() {
	Context("getEmptyProtoValueForType", func() {
		It("getEmptyProtoValueForType: Case 1", func() {
			tests := []struct {
				typ         base.AttributeType
				expectedAny *anypb.Any
				expectedErr error
			}{
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_STRING,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.StringValue",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.StringArrayValue",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_INTEGER,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.IntegerValue",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.IntegerArrayValue",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_DOUBLE,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.DoubleValue",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.DoubleArrayValue",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.BooleanValue",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.BooleanArrayValue",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: 15,
					expectedAny: &anypb.Any{
						TypeUrl: "",
						Value:   []byte(""),
					},
					expectedErr: errors.New("unknown type"),
				},
			}

			for _, test := range tests {
				result, err := getEmptyProtoValueForType(test.typ)

				if err == nil {
					Expect(err).ShouldNot(HaveOccurred())
				} else {
					Expect(err).Should(Equal(test.expectedErr))
				}

				if test.expectedErr == nil {
					Expect(test.expectedAny.GetTypeUrl()).Should(Equal(result.GetTypeUrl()))
					Expect(test.expectedAny.GetValue()).Should(Equal(result.GetValue()))
				}
			}
		})
	})

	Context("ConvertToAnyPB", func() {
		It("ConvertToAnyPB: Case 1", func() {
			// Test cases with supported types
			boolArray := []bool{true, false}
			intValue := 42
			intArray := []int32{1, 2, 3}
			floatValue := 3.14
			floatArray := []float64{1.1, 2.2}
			stringValue := "hello"
			stringArray := []string{"a", "b", "c"}

			supportedTestCases := []struct {
				value    interface{}
				typeName string
			}{
				{true, "base.v1.BooleanValue"},
				{boolArray, "base.v1.BooleanArrayValue"},
				{intValue, "base.v1.IntegerValue"},
				{intArray, "base.v1.IntegerArrayValue"},
				{floatValue, "base.v1.DoubleValue"},
				{floatArray, "base.v1.DoubleArrayValue"},
				{stringValue, "base.v1.StringValue"},
				{stringArray, "base.v1.StringArrayValue"},
			}

			for _, testCase := range supportedTestCases {
				anyPB, err := ConvertToAnyPB(testCase.value)
				if err != nil {
					Expect(err).ShouldNot(HaveOccurred())
				}
				if anyPB == nil {
					Expect(err).ShouldNot(HaveOccurred())
				}
				// Check if the type URL matches the expected type
				if anyPB.GetTypeUrl() != "type.googleapis.com/"+testCase.typeName {
					Expect(err).ShouldNot(HaveOccurred())
				}
			}

			// Test case with an unsupported type
			unsupportedValue := complex(1, 2)
			_, err := ConvertToAnyPB(unsupportedValue)
			if err == nil {
				Expect(err).ShouldNot(HaveOccurred())
			}
		})
	})

	Context("getEmptyValueForType", func() {
		It("getEmptyValueForType: Case 1", func() {
			// String type
			emptyString := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_STRING)
			Expect(emptyString).Should(Equal(""))

			// String array type
			emptyStringArray := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY)
			expectedEmptyStringArray := []string{}
			Expect(reflect.DeepEqual(emptyStringArray, expectedEmptyStringArray)).Should(Equal(true))

			// Integer type
			emptyInteger := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_INTEGER)
			Expect(emptyInteger).Should(Equal(0))

			// Integer array type
			emptyIntegerArray := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY)
			expectedEmptyIntegerArray := []int32{}
			Expect(reflect.DeepEqual(emptyIntegerArray, expectedEmptyIntegerArray)).Should(Equal(true))

			// Double type
			emptyDouble := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_DOUBLE)
			Expect(emptyDouble).Should(Equal(0.0))

			// Double array type
			emptyDoubleArray := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY)
			expectedEmptyDoubleArray := []float64{}
			Expect(reflect.DeepEqual(emptyDoubleArray, expectedEmptyDoubleArray)).Should(Equal(true))

			// Boolean type
			emptyBoolean := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN)
			Expect(emptyBoolean).Should(Equal(false))

			// Boolean array type
			emptyBooleanArray := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY)
			expectedEmptyBooleanArray := []bool{}
			Expect(reflect.DeepEqual(emptyBooleanArray, expectedEmptyBooleanArray)).Should(Equal(true))

			// Test case for an unknown type (returns nil)
			unknownType := getEmptyValueForType(base.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED)
			Expect(unknownType).Should(BeNil())
		})
	})

	Context("getDuplicates", func() {
		It("getDuplicates: Case 1", func() {
			// Test case with duplicates
			inputWithDuplicates := []string{"apple", "banana", "apple", "cherry", "banana", "date"}
			expectedDuplicates := []string{"apple", "banana"}

			duplicates := getDuplicates(inputWithDuplicates)

			Expect(isSameArray(expectedDuplicates, duplicates)).Should(Equal(true))

			// Test case with no duplicates
			inputWithoutDuplicates := []string{"apple", "banana", "cherry", "date"}
			expectedNoDuplicates := []string{}

			noDuplicates := getDuplicates(inputWithoutDuplicates)

			Expect(isSameArray(expectedNoDuplicates, noDuplicates)).Should(Equal(true))

			// Test case with an empty input slice
			emptyInput := []string{}
			expectedEmpty := []string{}

			emptyResult := getDuplicates(emptyInput)

			Expect(reflect.DeepEqual(emptyResult, expectedEmpty)).Should(Equal(true))
		})
	})

	Context("IsRelational", func() {
		It("IsRelational: Case 1", func() {
			sch, err := schema.NewSchemaFromStringDefinitions(false, `
			entity user {}
	
			entity repository {
	
				relation admin @user
	
				attribute ip_range string[]
				attribute public boolean
	
				permission edit = public
				permission view = check_ip_range(ip_range) or admin
				permission delete = view and admin
			}
	
			rule check_ip_range(ip_range string[]) {
				context.data.ip_address in ip_range
			}`)

			Expect(err).ShouldNot(HaveOccurred())

			Expect(IsRelational(sch.EntityDefinitions["repository"], "view")).Should(Equal(true))
			Expect(IsRelational(sch.EntityDefinitions["repository"], "delete")).Should(Equal(true))
			Expect(IsRelational(sch.EntityDefinitions["repository"], "admin")).Should(Equal(true))
			Expect(IsRelational(sch.EntityDefinitions["repository"], "edit")).Should(Equal(true))
			Expect(IsRelational(sch.EntityDefinitions["repository"], "public")).Should(Equal(false))
			Expect(IsRelational(sch.EntityDefinitions["repository"], "ip_range")).Should(Equal(false))
			Expect(IsRelational(sch.EntityDefinitions["repository"], "check_ip_range")).Should(Equal(false))
		})
	})

	Context("GenerateKey", func() {
		It("GenerateKey: Case 1", func() {
			k1 := GenerateKey(&base.PermissionCheckRequest{
				TenantId: "t1",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "1",
				},
				Permission: "view",
				Subject: &base.Subject{
					Type: "user",
					Id:   "12",
				},
			}, true)

			Expect(k1).Should(Equal("check|t1|organization:1#view@user:12"))

			ss, err := structpb.NewStruct(map[string]interface{}{
				"balance": 1000,
				"public":  true,
			})
			Expect(err).ShouldNot(HaveOccurred())

			k2 := GenerateKey(&base.PermissionCheckRequest{
				TenantId: "t1",
				Context: &base.Context{
					Tuples: []*base.Tuple{
						{
							Entity: &base.Entity{
								Type: "organization",
								Id:   "1",
							},
							Relation: "context_user",
							Subject: &base.Subject{
								Type: "user",
								Id:   "12",
							},
						},
					},
					Data: ss,
				},
				Entity: &base.Entity{
					Type: "organization",
					Id:   "1",
				},
				Permission: "view",
				Subject: &base.Subject{
					Type: "user",
					Id:   "12",
				},
			}, true)

			Expect(k2).Should(Equal("check|t1|organization:1#context_user@user:12,balance:1000,public:true|organization:1#view@user:12"))

			any1, err := anypb.New(&base.IntegerValue{
				Data: 1000,
			})
			Expect(err).ShouldNot(HaveOccurred())

			k3 := GenerateKey(&base.PermissionCheckRequest{
				TenantId: "t1",
				Context: &base.Context{
					Tuples: []*base.Tuple{
						{
							Entity: &base.Entity{
								Type: "organization",
								Id:   "1",
							},
							Relation: "context_user",
							Subject: &base.Subject{
								Type: "user",
								Id:   "12",
							},
						},
					},
					Attributes: []*base.Attribute{
						{
							Entity: &base.Entity{
								Type: "organization",
								Id:   "1",
							},
							Attribute: "public",
							Value:     any1,
						},
					},
					Data: ss,
				},
				Entity: &base.Entity{
					Type: "organization",
					Id:   "1",
				},
				Permission: "view",
				Subject: &base.Subject{
					Type: "user",
					Id:   "12",
				},
			}, true)

			Expect(k3).Should(Equal("check|t1|organization:1#context_user@user:12,organization:1$public|integer:1000,balance:1000,public:true|organization:1#view@user:12"))

			k4 := GenerateKey(&base.PermissionCheckRequest{
				TenantId: "t1",
				Context: &base.Context{
					Data: ss,
				},
				Entity: &base.Entity{
					Type: "organization",
					Id:   "1",
				},
				Permission: "public",
				Subject: &base.Subject{
					Type: "user",
					Id:   "12",
				},
			}, false)

			Expect(k4).Should(Equal("check|t1|balance:1000,public:true|organization:1$public"))
		})
	})
})
