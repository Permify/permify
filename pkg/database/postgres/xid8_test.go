package postgres

import (
	"github.com/jackc/pgtype"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("XID8", func() {
	Context("Set", func() {
		It("Case 1: Success with int64", func() {
			var x XID8
			err := x.Set(int64(12345))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Uint).Should(Equal(uint64(12345)))
			Expect(x.Status).Should(Equal(pgtype.Present))
		})

		It("Case 2: Success with uint64", func() {
			var x XID8
			err := x.Set(uint64(98765))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Uint).Should(Equal(uint64(98765)))
			Expect(x.Status).Should(Equal(pgtype.Present))
		})

		It("Case 3: Error with negative int64", func() {
			var x XID8
			err := x.Set(int64(-1))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("less than minimum value"))
		})

		It("Case 4: Error with unsupported type", func() {
			var x XID8
			err := x.Set("invalid")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("cannot convert"))
		})
	})

	Context("Get", func() {
		It("Case 1: Present status", func() {
			x := XID8{Uint: 12345, Status: pgtype.Present}
			result := x.Get()
			Expect(result).Should(Equal(uint64(12345)))
		})

		It("Case 2: Null status", func() {
			x := XID8{Status: pgtype.Null}
			result := x.Get()
			Expect(result).Should(BeNil())
		})
	})

	Context("DecodeText", func() {
		It("Case 1: Success", func() {
			var x XID8
			err := x.DecodeText(nil, []byte("12345"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Uint).Should(Equal(uint64(12345)))
			Expect(x.Status).Should(Equal(pgtype.Present))
		})

		It("Case 2: Null input", func() {
			var x XID8
			err := x.DecodeText(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Status).Should(Equal(pgtype.Null))
		})
	})

	Context("EncodeText", func() {
		It("Case 1: Present status", func() {
			x := XID8{Uint: 12345, Status: pgtype.Present}
			result, err := x.EncodeText(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(result)).Should(Equal("12345"))
		})

		It("Case 2: Null status", func() {
			x := XID8{Status: pgtype.Null}
			result, err := x.EncodeText(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(BeNil())
		})
	})

	Context("Scan", func() {
		It("Case 1: Success with uint64", func() {
			var x XID8
			err := x.Scan(uint64(12345))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Uint).Should(Equal(uint64(12345)))
			Expect(x.Status).Should(Equal(pgtype.Present))
		})

		It("Case 2: Success with int64", func() {
			var x XID8
			err := x.Scan(int64(12345))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Uint).Should(Equal(uint64(12345)))
			Expect(x.Status).Should(Equal(pgtype.Present))
		})

		It("Case 3: Success with string", func() {
			var x XID8
			err := x.Scan("12345")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Uint).Should(Equal(uint64(12345)))
			Expect(x.Status).Should(Equal(pgtype.Present))
		})

		It("Case 4: Null input", func() {
			var x XID8
			err := x.Scan(nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Status).Should(Equal(pgtype.Null))
		})
	})

	Context("Value", func() {
		It("Case 1: Present status", func() {
			x := XID8{Uint: 12345, Status: pgtype.Present}
			result, err := x.Value()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(int64(12345)))
		})

		It("Case 2: Null status", func() {
			x := XID8{Status: pgtype.Null}
			result, err := x.Value()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(BeNil())
		})
	})
})

var _ = Describe("Codecs", func() {
	Context("SnapshotCodec", func() {
		It("Case 1: FormatCode", func() {
			codec := SnapshotCodec{}
			Expect(codec.FormatCode()).Should(Equal(int16(pgtype.BinaryFormatCode)))
		})
	})

	Context("Uint64Codec", func() {
		It("Case 1: FormatCode", func() {
			codec := Uint64Codec{}
			Expect(codec.FormatCode()).Should(Equal(int16(pgtype.BinaryFormatCode)))
		})
	})
})
