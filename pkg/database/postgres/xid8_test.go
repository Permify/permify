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

	Context("AssignTo", func() {
		It("Case 1: Success with *uint64", func() {
			x := XID8{Uint: 12345, Status: pgtype.Present}
			var target uint64
			err := x.AssignTo(&target)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(target).Should(Equal(uint64(12345)))
		})

		It("Case 2: Success with **uint64", func() {
			x := XID8{Uint: 12345, Status: pgtype.Present}
			var target *uint64
			err := x.AssignTo(&target)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(*target).Should(Equal(uint64(12345)))
		})

		It("Case 3: Success with **uint64 and null status", func() {
			x := XID8{Status: pgtype.Null}
			var target *uint64
			err := x.AssignTo(&target)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(target).Should(BeNil())
		})

		It("Case 4: Error with null status and *uint64", func() {
			x := XID8{Status: pgtype.Null}
			var target uint64
			err := x.AssignTo(&target)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("cannot assign"))
		})

		It("Case 5: Success with unsupported target type (no error)", func() {
			x := XID8{Uint: 12345, Status: pgtype.Present}
			var target string
			err := x.AssignTo(&target)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("DecodeBinary", func() {
		It("Case 1: Null input", func() {
			var x XID8
			err := x.DecodeBinary(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(x.Status).Should(Equal(pgtype.Null))
		})

		It("Case 2: Invalid input size", func() {
			var x XID8
			err := x.DecodeBinary(nil, []byte{0x01, 0x02})
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("invalid length"))
		})
	})

	Context("EncodeBinary", func() {
		It("Case 1: Present status", func() {
			x := XID8{Uint: 12345, Status: pgtype.Present}
			result, err := x.EncodeBinary(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30, 0x39}))
		})

		It("Case 2: Null status", func() {
			x := XID8{Status: pgtype.Null}
			result, err := x.EncodeBinary(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(BeNil())
		})

		It("Case 3: Undefined status", func() {
			x := XID8{Status: pgtype.Undefined}
			_, err := x.EncodeBinary(nil, nil)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("undefined status"))
		})
	})
})
