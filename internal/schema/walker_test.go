package schema

import (
	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
)

var _ = Describe("walker", func() {
	Context("walker", func() {
		It("Case 1", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user

				attribute is_public boolean

				permission edit = admin or is_public
			}
			entity container {
				relation parent @organization
				relation container_admin @user

				permission admin = parent.admin or container_admin

				permission edit = container_admin or parent.edit
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, _ := c.Compile()

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			err = w.Walk("container", "edit")

			spew.Dump(err)
		})
	})
})
