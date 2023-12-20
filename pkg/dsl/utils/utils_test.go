package utils

import (
	"testing"

	"github.com/google/cel-go/cel"

	"github.com/google/cel-go/common/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "utils suite")
}

var _ = Describe("Utils", func() {
	Describe("Key function", func() {
		It("should concatenate two strings with a '#' in between", func() {
			result := Key("organization", "member")
			Expect(result).To(Equal("organization#member"))
		})

		It("should handle empty strings", func() {
			result := Key("", "")
			Expect(result).To(Equal("#"))
		})
	})

	Describe("ArgumentsAsCelEnv function", func() {
		It("should return an error for unrecognized attribute types", func() {
			arguments := map[string]base.AttributeType{
				"invalidAttribute": base.AttributeType(9999),
			}
			_, err := ArgumentsAsCelEnv(arguments)
			Expect(err.Error()).To(ContainSubstring("unrecognized AttributeType"))
		})
	})

	Describe("GetCelType function", func() {
		It("should return the CEL type for a attribute type", func() {
			_, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED)
			Expect(err.Error()).Should(Equal("unrecognized AttributeType: ATTRIBUTE_TYPE_UNSPECIFIED"))

			celStringType, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_STRING)
			Expect(err).NotTo(HaveOccurred())
			Expect(celStringType).To(Equal(types.StringType))

			celIntegerType, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_INTEGER)
			Expect(err).NotTo(HaveOccurred())
			Expect(celIntegerType).To(Equal(types.IntType))

			celBooleanType, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN)
			Expect(err).NotTo(HaveOccurred())
			Expect(celBooleanType).To(Equal(types.BoolType))

			celDoubleType, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_DOUBLE)
			Expect(err).NotTo(HaveOccurred())
			Expect(celDoubleType).To(Equal(types.DoubleType))

			celStringArrayType, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY)
			Expect(err).NotTo(HaveOccurred())
			Expect(celStringArrayType).To(Equal(cel.ListType(cel.StringType)))

			celIntegerArrayType, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY)
			Expect(err).NotTo(HaveOccurred())
			Expect(celIntegerArrayType).To(Equal(cel.ListType(cel.IntType)))

			celBooleanArrayType, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY)
			Expect(err).NotTo(HaveOccurred())
			Expect(celBooleanArrayType).To(Equal(cel.ListType(cel.BoolType)))

			celDoubleArrayType, err := GetCelType(base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY)
			Expect(err).NotTo(HaveOccurred())
			Expect(celDoubleArrayType).To(Equal(cel.ListType(cel.DoubleType)))
		})
	})

	Describe("ConvertProtoAnyToInterface function", func() {
		Context("when the proto message is of type StringValue", func() {
			It("should return the correct data", func() {
				str, err := anypb.New(&base.StringValue{Data: "string_data"})
				Expect(err).NotTo(HaveOccurred())

				strResult := ConvertProtoAnyToInterface(str)
				Expect(strResult).To(Equal("string_data"))

				str2, err := anypb.New(&base.StringValue{Data: "string_data"})
				str2.Value = []byte("asd")
				Expect(err).NotTo(HaveOccurred())

				strResult2 := ConvertProtoAnyToInterface(str2)
				Expect(strResult2).To(Equal(""))

				bo, err := anypb.New(&base.BooleanValue{Data: true})
				Expect(err).NotTo(HaveOccurred())

				boResult := ConvertProtoAnyToInterface(bo)
				Expect(boResult).To(Equal(true))

				bo2, err := anypb.New(&base.BooleanValue{Data: true})
				bo2.Value = []byte("asd")
				Expect(err).NotTo(HaveOccurred())

				boResult2 := ConvertProtoAnyToInterface(bo2)
				Expect(boResult2).To(Equal(false))

				integer, err := anypb.New(&base.IntegerValue{Data: 97})
				Expect(err).NotTo(HaveOccurred())

				integerResult := ConvertProtoAnyToInterface(integer)
				Expect(integerResult).To(Equal(int32(97)))

				integer2, err := anypb.New(&base.IntegerValue{Data: 97})
				integer2.Value = []byte("123")
				Expect(err).NotTo(HaveOccurred())

				integerResult2 := ConvertProtoAnyToInterface(integer2)
				Expect(integerResult2).To(Equal(0))

				double, err := anypb.New(&base.DoubleValue{Data: 90.234})
				Expect(err).NotTo(HaveOccurred())

				doubleResult := ConvertProtoAnyToInterface(double)
				Expect(doubleResult).To(Equal(90.234))

				double2, err := anypb.New(&base.DoubleValue{Data: 90.234})
				double2.Value = []byte("80")
				Expect(err).NotTo(HaveOccurred())

				doubleResult2 := ConvertProtoAnyToInterface(double2)
				Expect(doubleResult2).To(Equal(0.0))

				strArray, err := anypb.New(&base.StringArrayValue{Data: []string{"string_data_1", "string_data_2", "string_data_3"}})
				Expect(err).NotTo(HaveOccurred())

				strArrayResult := ConvertProtoAnyToInterface(strArray)
				Expect(strArrayResult).To(Equal([]string{"string_data_1", "string_data_2", "string_data_3"}))

				strArray2, err := anypb.New(&base.StringArrayValue{Data: []string{"string_data_1", "string_data_2", "string_data_3"}})
				strArray2.Value = []byte("asd")
				Expect(err).NotTo(HaveOccurred())

				strArrayResult2 := ConvertProtoAnyToInterface(strArray2)
				Expect(strArrayResult2).To(Equal([]string{}))

				integerArray, err := anypb.New(&base.IntegerArrayValue{Data: []int32{97, 23, 234, 564, 43, 3}})
				Expect(err).NotTo(HaveOccurred())

				integerArrayResult := ConvertProtoAnyToInterface(integerArray)
				Expect(integerArrayResult).To(Equal([]int32{97, 23, 234, 564, 43, 3}))

				integerArray2, err := anypb.New(&base.IntegerArrayValue{Data: []int32{97, 23, 234, 564, 43, 3}})
				integerArray2.Value = []byte("213")
				Expect(err).NotTo(HaveOccurred())

				integerArrayResul2 := ConvertProtoAnyToInterface(integerArray2)
				Expect(integerArrayResul2).To(Equal([]int32{}))

				doubleArray, err := anypb.New(&base.DoubleArrayValue{Data: []float64{90.234, 4234, 234, 4532, 23, 543.43}})
				Expect(err).NotTo(HaveOccurred())

				doubleArrayResult := ConvertProtoAnyToInterface(doubleArray)
				Expect(doubleArrayResult).To(Equal([]float64{90.234, 4234, 234, 4532, 23, 543.43}))

				doubleArray2, err := anypb.New(&base.DoubleArrayValue{Data: []float64{90.234, 4234, 234, 4532, 23, 543.43}})
				doubleArray2.Value = []byte("213")
				Expect(err).NotTo(HaveOccurred())

				doubleArrayResult2 := ConvertProtoAnyToInterface(doubleArray2)
				Expect(doubleArrayResult2).To(Equal([]float64{}))

				boArray, err := anypb.New(&base.BooleanArrayValue{Data: []bool{false, true}})
				Expect(err).NotTo(HaveOccurred())

				boArrayResult := ConvertProtoAnyToInterface(boArray)
				Expect(boArrayResult).To(Equal([]bool{false, true}))

				boArray2, err := anypb.New(&base.BooleanArrayValue{Data: []bool{false, true}})
				boArray2.Value = []byte("213")
				Expect(err).NotTo(HaveOccurred())

				boArrayResult2 := ConvertProtoAnyToInterface(boArray2)
				Expect(boArrayResult2).To(Equal([]bool{}))
			})
		})

		Context("when the proto message is of type BooleanValue", func() {
			It("should return the boolean data", func() {
				boolAny, err := anypb.New(&base.BooleanValue{Data: true})
				Expect(err).NotTo(HaveOccurred())

				boolResult := ConvertProtoAnyToInterface(boolAny)
				Expect(boolResult).To(BeTrue())

				stringAny, err := anypb.New(&base.StringValue{Data: "127.0.0.1"})
				Expect(err).NotTo(HaveOccurred())

				stringResult := ConvertProtoAnyToInterface(stringAny)
				Expect(stringResult).To(Equal("127.0.0.1"))

				integerAny, err := anypb.New(&base.IntegerValue{Data: 128})
				Expect(err).NotTo(HaveOccurred())

				integerResult := ConvertProtoAnyToInterface(integerAny)
				Expect(integerResult).To(Equal(int32(128)))

				doubleAny, err := anypb.New(&base.DoubleValue{Data: 97.052})
				Expect(err).NotTo(HaveOccurred())

				doubleResult := ConvertProtoAnyToInterface(doubleAny)
				Expect(doubleResult).To(Equal(97.052))

				boolArrayAny, err := anypb.New(&base.BooleanArrayValue{Data: []bool{true, false}})
				Expect(err).NotTo(HaveOccurred())

				boolArrayResult := ConvertProtoAnyToInterface(boolArrayAny)
				Expect(boolArrayResult).To(Equal([]bool{true, false}))

				stringArrayAny, err := anypb.New(&base.StringArrayValue{Data: []string{"127.0.0.1", "127.0.0.2"}})
				Expect(err).NotTo(HaveOccurred())

				stringArrayResult := ConvertProtoAnyToInterface(stringArrayAny)
				Expect(stringArrayResult).To(Equal([]string{"127.0.0.1", "127.0.0.2"}))

				integerArrayAny, err := anypb.New(&base.IntegerArrayValue{Data: []int32{128, 97, 389, 1000}})
				Expect(err).NotTo(HaveOccurred())

				integerArrayResult := ConvertProtoAnyToInterface(integerArrayAny)
				Expect(integerArrayResult).To(Equal([]int32{128, 97, 389, 1000}))

				doubleArrayAny, err := anypb.New(&base.DoubleArrayValue{Data: []float64{97.052, 234.234, 5342.232}})
				Expect(err).NotTo(HaveOccurred())

				doubleArrayResult := ConvertProtoAnyToInterface(doubleArrayAny)
				Expect(doubleArrayResult).To(Equal([]float64{97.052, 234.234, 5342.232}))
			})
		})

		Context("when the proto message type is unknown", func() {
			It("should return an empty string", func() {
				unknown, err := anypb.New(&anypb.Any{TypeUrl: "unknown"})
				Expect(err).NotTo(HaveOccurred())

				unknownResult := ConvertProtoAnyToInterface(unknown)
				Expect(unknownResult).To(Equal(""))
			})
		})
	})
})
