package memory

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/pkg/database/memory"
)

var _ = Describe("Watch", func() {
	var db *memory.Memory

	var watcher *Watch

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database

		watcher = NewWatcher(db)
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Watch", func() {
		It("watch", func() {
			_, errs := watcher.Watch(context.Background(), "t1", "")
			select {
			case err := <-errs:
				// Handle and assert the error
				Expect(err).ShouldNot(HaveOccurred())
			case <-time.After(time.Second * 10):
				fmt.Println("test timed out")
			}
		})
	})
})
