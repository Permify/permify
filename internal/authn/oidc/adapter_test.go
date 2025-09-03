package oidc

import (
	"bytes"
	"log/slog"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSlogAdapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SlogAdapter suite")
}

var _ = Describe("SlogAdapter", func() {
	var (
		buffer  *bytes.Buffer
		logger  *slog.Logger
		adapter SlogAdapter
	)

	BeforeEach(func() {
		buffer = &bytes.Buffer{}
		logger = slog.New(slog.NewTextHandler(buffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		adapter = SlogAdapter{Logger: logger}
	})

	Describe("Error", func() {
		It("should log error messages", func() {
			adapter.Error("test error message", "key1", "value1", "key2", "value2")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=ERROR"))
			Expect(output).To(ContainSubstring("msg=\"test error message\""))
			Expect(output).To(ContainSubstring("key1=value1"))
			Expect(output).To(ContainSubstring("key2=value2"))
		})

		It("should log error messages without key-value pairs", func() {
			adapter.Error("simple error message")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=ERROR"))
			Expect(output).To(ContainSubstring("msg=\"simple error message\""))
		})
	})

	Describe("Info", func() {
		It("should log info messages", func() {
			adapter.Info("test info message", "key1", "value1", "key2", "value2")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=INFO"))
			Expect(output).To(ContainSubstring("msg=\"test info message\""))
			Expect(output).To(ContainSubstring("key1=value1"))
			Expect(output).To(ContainSubstring("key2=value2"))
		})

		It("should log info messages without key-value pairs", func() {
			adapter.Info("simple info message")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=INFO"))
			Expect(output).To(ContainSubstring("msg=\"simple info message\""))
		})
	})

	Describe("Debug", func() {
		It("should log debug messages", func() {
			adapter.Debug("test debug message", "key1", "value1", "key2", "value2")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=INFO")) // Note: Debug uses Info level
			Expect(output).To(ContainSubstring("msg=\"test debug message\""))
			Expect(output).To(ContainSubstring("key1=value1"))
			Expect(output).To(ContainSubstring("key2=value2"))
		})

		It("should log debug messages without key-value pairs", func() {
			adapter.Debug("simple debug message")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=INFO")) // Note: Debug uses Info level
			Expect(output).To(ContainSubstring("msg=\"simple debug message\""))
		})
	})

	Describe("Warn", func() {
		It("should log warn messages", func() {
			adapter.Warn("test warn message", "key1", "value1", "key2", "value2")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=WARN"))
			Expect(output).To(ContainSubstring("msg=\"test warn message\""))
			Expect(output).To(ContainSubstring("key1=value1"))
			Expect(output).To(ContainSubstring("key2=value2"))
		})

		It("should log warn messages without key-value pairs", func() {
			adapter.Warn("simple warn message")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=WARN"))
			Expect(output).To(ContainSubstring("msg=\"simple warn message\""))
		})
	})

	Describe("Integration", func() {
		It("should handle multiple log levels in sequence", func() {
			adapter.Error("error message")
			adapter.Warn("warn message")
			adapter.Info("info message")
			adapter.Debug("debug message")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=ERROR"))
			Expect(output).To(ContainSubstring("level=WARN"))
			Expect(output).To(ContainSubstring("level=INFO"))
			Expect(output).To(ContainSubstring("msg=\"error message\""))
			Expect(output).To(ContainSubstring("msg=\"warn message\""))
			Expect(output).To(ContainSubstring("msg=\"info message\""))
			Expect(output).To(ContainSubstring("msg=\"debug message\""))
		})

		It("should handle empty messages", func() {
			adapter.Error("")
			adapter.Warn("")
			adapter.Info("")
			adapter.Debug("")

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=ERROR"))
			Expect(output).To(ContainSubstring("level=WARN"))
			Expect(output).To(ContainSubstring("level=INFO"))
			Expect(output).To(ContainSubstring("msg=\"\""))
		})

		It("should handle odd number of key-value pairs", func() {
			// This tests the behavior when there's an odd number of key-value pairs
			adapter.Info("test message", "key1", "value1", "key2") // key2 has no value

			output := buffer.String()
			Expect(output).To(ContainSubstring("level=INFO"))
			Expect(output).To(ContainSubstring("msg=\"test message\""))
			Expect(output).To(ContainSubstring("key1=value1"))
		})
	})
})
