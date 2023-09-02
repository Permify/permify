package engines

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

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
						TypeUrl: "type.googleapis.com/base.v1.String",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.StringArray",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_INTEGER,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.Integer",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.IntegerArray",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_DOUBLE,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.Double",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.DoubleArray",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.Boolean",
						Value:   []byte(""),
					},
					expectedErr: nil,
				},
				{
					typ: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY,
					expectedAny: &anypb.Any{
						TypeUrl: "type.googleapis.com/base.v1.BooleanArray",
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
				{true, "base.v1.Boolean"},
				{boolArray, "base.v1.BooleanArray"},
				{intValue, "base.v1.Integer"},
				{intArray, "base.v1.IntegerArray"},
				{floatValue, "base.v1.Double"},
				{floatArray, "base.v1.DoubleArray"},
				{stringValue, "base.v1.String"},
				{stringArray, "base.v1.StringArray"},
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

			Expect(reflect.DeepEqual(expectedDuplicates, duplicates)).Should(Equal(true))

			// Test case with no duplicates
			inputWithoutDuplicates := []string{"apple", "banana", "cherry", "date"}
			expectedNoDuplicates := []string{}

			noDuplicates := getDuplicates(inputWithoutDuplicates)

			Expect(reflect.DeepEqual(expectedNoDuplicates, noDuplicates)).Should(Equal(true))

			// Test case with an empty input slice
			emptyInput := []string{}
			expectedEmpty := []string{}

			emptyResult := getDuplicates(emptyInput)

			Expect(reflect.DeepEqual(emptyResult, expectedEmpty)).Should(Equal(true))
		})
	})
})
