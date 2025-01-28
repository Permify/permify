package balancer

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("balancer", func() {
	Context("ServiceConfigJSON", func() {
		It("should generate valid JSON", func() {
			// Set up a sample configuration.
			config := &Config{
				PartitionCount:    271,
				ReplicationFactor: 20,
				Load:              1.25,
				PickerWidth:       3,
			}

			// Generate the JSON using ServiceConfigJSON.
			jsonString, err := config.ServiceConfigJSON()

			// Expect no error during JSON generation.
			Expect(err).ToNot(HaveOccurred(), "ServiceConfigJSON should not return an error")

			// Validate the parsed Config fields match the original configuration.
			Expect(jsonString).To(Equal("{\"loadBalancingConfig\":[{\"consistenthashing\":{\"partitionCount\":271,\"replicationFactor\":20,\"load\":1.25,\"pickerWidth\":3}}]}"))
		})
	})
})
