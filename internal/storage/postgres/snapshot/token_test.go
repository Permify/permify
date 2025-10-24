package snapshot

import (
	"testing"

	"github.com/jackc/pgtype"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/token"
)

// TestToken -
func TestToken(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "token-suite")
}

var _ = Describe("token", func() {
	Context("Encode", func() {
		It("Case 1: Success - Legacy format (empty snapshot)", func() {
			tests := []struct {
				target   token.SnapToken
				expected string
			}{
				{NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, ""), "BAAAAAAAAAA="},
				{NewToken(postgres.XID8{Uint: 12, Status: pgtype.Present}, ""), "DAAAAAAAAAA="},
				{NewToken(postgres.XID8{Uint: 43242, Status: pgtype.Present}, ""), "6qgAAAAAAAA="},
				{NewToken(postgres.XID8{Uint: 54342345, Status: pgtype.Present}, ""), "yTI9AwAAAAA="},
				{NewToken(postgres.XID8{Uint: 87648723472386, Status: pgtype.Present}, ""), "AhAHT7dPAAA="},
				{NewToken(postgres.XID8{Uint: 2349875239487420823, Status: pgtype.Present}, ""), "lzkihBRvnCA="},
			}

			for _, tt := range tests {
				Expect(tt.target.Encode().String()).Should(Equal(tt.expected))
			}
		})

		It("Case 1.1: Success - New format (with snapshot)", func() {
			tests := []struct {
				target   token.SnapToken
				expected string
			}{
				{NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, "100:100:"), "NDoxMDA6MTAwOg=="},
				{NewToken(postgres.XID8{Uint: 12, Status: pgtype.Present}, "200:200:150"), "MTI6MjAwOjIwMDoxNTA="},
				{NewToken(postgres.XID8{Uint: 43242, Status: pgtype.Present}, "300:300:250,280"), "NDMyNDI6MzAwOjMwMDoyNTAsMjgw"},
			}

			for _, tt := range tests {
				Expect(tt.target.Encode().String()).Should(Equal(tt.expected))
			}
		})

		It("Case 2: Fail", func() {
			tests := []struct {
				target   token.SnapToken
				expected string
			}{
				{NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, ""), " BAAAAAAAAAA="},
			}

			for _, tt := range tests {
				Expect(tt.target.Encode().String()).ShouldNot(Equal(tt.expected))
			}
		})

		It("Case 3: Eg Success", func() {
			tests := []struct {
				token  token.SnapToken
				target token.SnapToken
			}{
				{NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, ""), NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, "")},
			}

			for _, tt := range tests {
				Expect(tt.token.Eg(tt.target)).Should(BeTrue())
			}
		})

		It("Case 4: Gt Success", func() {
			tests := []struct {
				token  token.SnapToken
				target token.SnapToken
			}{
				{
					NewToken(postgres.XID8{Uint: 6, Status: pgtype.Present}, ""),
					NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, ""),
				},
			}

			for _, tt := range tests {
				Expect(tt.token.Gt(tt.target)).Should(BeTrue())
			}
		})

		It("Case 5: Lt Success", func() {
			tests := []struct {
				token  token.SnapToken
				target token.SnapToken
			}{
				{
					NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, ""),
					NewToken(postgres.XID8{Uint: 6, Status: pgtype.Present}, ""),
				},
			}

			for _, tt := range tests {
				Expect(tt.token.Lt(tt.target)).Should(BeTrue())
			}
		})
	})

	Context("Decode", func() {
		It("Case 1: Success - Legacy format (empty snapshot)", func() {
			tests := []struct {
				target   token.EncodedSnapToken
				expected token.SnapToken
			}{
				{EncodedToken{Value: "BAAAAAAAAAA="}, NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, "")},
				{EncodedToken{Value: "DAAAAAAAAAA="}, NewToken(postgres.XID8{Uint: 12, Status: pgtype.Present}, "")},
				{EncodedToken{Value: "6qgAAAAAAAA="}, NewToken(postgres.XID8{Uint: 43242, Status: pgtype.Present}, "")},
				{EncodedToken{Value: "yTI9AwAAAAA="}, NewToken(postgres.XID8{Uint: 54342345, Status: pgtype.Present}, "")},
				{EncodedToken{Value: "AhAHT7dPAAA="}, NewToken(postgres.XID8{Uint: 87648723472386, Status: pgtype.Present}, "")},
				{EncodedToken{Value: "lzkihBRvnCA="}, NewToken(postgres.XID8{Uint: 2349875239487420823, Status: pgtype.Present}, "")},
			}

			for _, tt := range tests {
				t, err := tt.target.Decode()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(t).Should(Equal(tt.expected))
			}
		})

		It("Case 1.1: Success - New format (with snapshot)", func() {
			tests := []struct {
				target   token.EncodedSnapToken
				expected token.SnapToken
			}{
				{EncodedToken{Value: "NDoxMDA6MTAwOg=="}, NewToken(postgres.XID8{Uint: 4, Status: pgtype.Present}, "100:100:")},
				{EncodedToken{Value: "MTI6MjAwOjIwMDoxNTA="}, NewToken(postgres.XID8{Uint: 12, Status: pgtype.Present}, "200:200:150")},
				{EncodedToken{Value: "NDMyNDI6MzAwOjMwMDoyNTAsMjgw"}, NewToken(postgres.XID8{Uint: 43242, Status: pgtype.Present}, "300:300:250,280")},
			}

			for _, tt := range tests {
				t, err := tt.target.Decode()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(t).Should(Equal(tt.expected))
			}
		})

		It("Case 2: Fail", func() {
			tests := []struct {
				target   token.EncodedSnapToken
				expected token.SnapToken
			}{
				{EncodedToken{Value: "invalid_base64"}, Token{Value: postgres.XID8{Uint: 4, Status: pgtype.Present}}},
			}

			for _, tt := range tests {
				t, err := tt.target.Decode()
				Expect(err).Should(HaveOccurred())
				Expect(t).Should(BeNil())
			}
		})
	})
})
