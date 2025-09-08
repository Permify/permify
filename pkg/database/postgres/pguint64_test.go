package postgres

import (
	"github.com/jackc/pgtype"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("pguint64", func() {
	Context("Set", func() {
		It("Case 1: Success with int64", func() {
			var p pguint64
			err := p.Set(int64(12345))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Uint).Should(Equal(uint64(12345)))
			Expect(p.Status).Should(Equal(pgtype.Present))
		})

		It("Case 2: Success with uint64", func() {
			var p pguint64
			err := p.Set(uint64(98765))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Uint).Should(Equal(uint64(98765)))
			Expect(p.Status).Should(Equal(pgtype.Present))
		})

		It("Case 3: Error with negative int64", func() {
			var p pguint64
			err := p.Set(int64(-1))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("less than minimum value"))
		})

		It("Case 4: Error with unsupported type", func() {
			var p pguint64
			err := p.Set("invalid")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("cannot convert"))
		})
	})

	Context("Get", func() {
		It("Case 1: Present status", func() {
			p := pguint64{Uint: 12345, Status: pgtype.Present}
			result := p.Get()
			Expect(result).Should(Equal(uint64(12345)))
		})

		It("Case 2: Null status", func() {
			p := pguint64{Status: pgtype.Null}
			result := p.Get()
			Expect(result).Should(BeNil())
		})

		It("Case 3: Undefined status", func() {
			p := pguint64{Status: pgtype.Undefined}
			result := p.Get()
			Expect(result).Should(Equal(pgtype.Undefined))
		})
	})

	Context("DecodeText", func() {
		It("Case 1: Success", func() {
			var p pguint64
			err := p.DecodeText(nil, []byte("12345"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Uint).Should(Equal(uint64(12345)))
			Expect(p.Status).Should(Equal(pgtype.Present))
		})

		It("Case 2: Null input", func() {
			var p pguint64
			err := p.DecodeText(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Status).Should(Equal(pgtype.Null))
		})

		It("Case 3: Invalid input", func() {
			var p pguint64
			err := p.DecodeText(nil, []byte("invalid"))
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("EncodeText", func() {
		It("Case 1: Present status", func() {
			p := pguint64{Uint: 12345, Status: pgtype.Present}
			result, err := p.EncodeText(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(result)).Should(Equal("12345"))
		})

		It("Case 2: Null status", func() {
			p := pguint64{Status: pgtype.Null}
			result, err := p.EncodeText(nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(BeNil())
		})

		It("Case 3: Undefined status", func() {
			p := pguint64{Status: pgtype.Undefined}
			_, err := p.EncodeText(nil, nil)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("undefined status"))
		})
	})

	Context("Scan", func() {
		It("Case 1: Success with uint64", func() {
			var p pguint64
			err := p.Scan(uint64(12345))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Uint).Should(Equal(uint64(12345)))
			Expect(p.Status).Should(Equal(pgtype.Present))
		})

		It("Case 2: Success with int64", func() {
			var p pguint64
			err := p.Scan(int64(12345))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Uint).Should(Equal(uint64(12345)))
			Expect(p.Status).Should(Equal(pgtype.Present))
		})

		It("Case 3: Success with string", func() {
			var p pguint64
			err := p.Scan("12345")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Uint).Should(Equal(uint64(12345)))
			Expect(p.Status).Should(Equal(pgtype.Present))
		})

		It("Case 4: Null input", func() {
			var p pguint64
			err := p.Scan(nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Status).Should(Equal(pgtype.Null))
		})

		It("Case 5: Error with unsupported type", func() {
			var p pguint64
			err := p.Scan(complex(1, 2))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("cannot scan"))
		})
	})

	Context("Value", func() {
		It("Case 1: Present status", func() {
			p := pguint64{Uint: 12345, Status: pgtype.Present}
			result, err := p.Value()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(int64(12345)))
		})

		It("Case 2: Null status", func() {
			p := pguint64{Status: pgtype.Null}
			result, err := p.Value()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(BeNil())
		})

		It("Case 3: Undefined status", func() {
			p := pguint64{Status: pgtype.Undefined}
			_, err := p.Value()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("undefined status"))
		})
	})
})
