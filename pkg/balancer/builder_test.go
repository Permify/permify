package balancer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cespare/xxhash/v2"
	"google.golang.org/grpc/balancer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/consistent"
)

var _ = Describe("Builder", func() {
	var builder Builder

	BeforeEach(func() {
		builder = NewBuilder(xxhash.Sum64)
	})

	Describe("NewBuilder", func() {
		It("should create a new builder with hasher", func() {
			b := NewBuilder(xxhash.Sum64)
			Expect(b).ToNot(BeNil())
			Expect(b.Name()).To(Equal("consistenthashing"))
		})

		It("should create a new builder with custom hasher", func() {
			customHasher := func(data []byte) uint64 {
				return 12345
			}
			b := NewBuilder(customHasher)
			Expect(b).ToNot(BeNil())
			Expect(b.Name()).To(Equal("consistenthashing"))
		})
	})

	Describe("Name", func() {
		It("should return correct balancer name", func() {
			Expect(builder.Name()).To(Equal("consistenthashing"))
		})
	})

	Describe("Build", func() {
		It("should create a new balancer instance", func() {
			// We'll test the Build method with nil to avoid the interface issue
			bal := builder.Build(nil, balancer.BuildOptions{})
			Expect(bal).ToNot(BeNil())
			Expect(bal).To(BeAssignableToTypeOf(&Balancer{}))
		})

		It("should create multiple balancer instances", func() {
			bal1 := builder.Build(nil, balancer.BuildOptions{})
			bal2 := builder.Build(nil, balancer.BuildOptions{})
			Expect(bal1).ToNot(BeNil())
			Expect(bal2).ToNot(BeNil())
			Expect(bal1).ToNot(Equal(bal2))
		})
	})

	Describe("ParseConfig", func() {
		It("should parse valid JSON configuration", func() {
			jsonConfig := `{"partitionCount": 100, "replicationFactor": 3, "load": 1.25, "pickerWidth": 2}`
			config, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())
			cfg := config.(*Config)
			Expect(cfg.PartitionCount).To(Equal(100))
			Expect(cfg.ReplicationFactor).To(Equal(3))
			Expect(cfg.Load).To(Equal(1.25))
			Expect(cfg.PickerWidth).To(Equal(2))
		})

		It("should apply default values for missing fields", func() {
			jsonConfig := `{}`
			config, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).ToNot(HaveOccurred())
			cfg := config.(*Config)
			Expect(cfg.PartitionCount).To(Equal(consistent.DefaultPartitionCount))
			Expect(cfg.ReplicationFactor).To(Equal(consistent.DefaultReplicationFactor))
			Expect(cfg.Load).To(Equal(consistent.DefaultLoad))
			Expect(cfg.PickerWidth).To(Equal(consistent.DefaultPickerWidth))
		})

		It("should apply default values for zero values", func() {
			jsonConfig := `{"partitionCount": 0, "replicationFactor": 0, "load": 0.5, "pickerWidth": 0}`
			config, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).ToNot(HaveOccurred())
			cfg := config.(*Config)
			Expect(cfg.PartitionCount).To(Equal(consistent.DefaultPartitionCount))
			Expect(cfg.ReplicationFactor).To(Equal(consistent.DefaultReplicationFactor))
			Expect(cfg.Load).To(Equal(consistent.DefaultLoad))
			Expect(cfg.PickerWidth).To(Equal(consistent.DefaultPickerWidth))
		})

		It("should apply default values for negative values", func() {
			jsonConfig := `{"partitionCount": -10, "replicationFactor": -5, "load": -1.0, "pickerWidth": -2}`
			config, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).ToNot(HaveOccurred())
			cfg := config.(*Config)
			Expect(cfg.PartitionCount).To(Equal(consistent.DefaultPartitionCount))
			Expect(cfg.ReplicationFactor).To(Equal(consistent.DefaultReplicationFactor))
			Expect(cfg.Load).To(Equal(consistent.DefaultLoad))
			Expect(cfg.PickerWidth).To(Equal(consistent.DefaultPickerWidth))
		})

		It("should handle partial configuration", func() {
			jsonConfig := `{"partitionCount": 200, "load": 2.5}`
			config, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).ToNot(HaveOccurred())
			cfg := config.(*Config)
			Expect(cfg.PartitionCount).To(Equal(200))
			Expect(cfg.ReplicationFactor).To(Equal(consistent.DefaultReplicationFactor))
			Expect(cfg.Load).To(Equal(2.5))
			Expect(cfg.PickerWidth).To(Equal(consistent.DefaultPickerWidth))
		})

		It("should return error for invalid JSON", func() {
			jsonConfig := `{"partitionCount": "invalid", "replicationFactor": true}`
			_, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to unmarshal LB policy config"))
		})

		It("should return error for malformed JSON", func() {
			jsonConfig := `{"partitionCount": 100, "replicationFactor": 3,}`
			_, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to unmarshal LB policy config"))
		})

		It("should handle empty JSON", func() {
			jsonConfig := `{}`
			config, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())
		})

		It("should handle null JSON", func() {
			jsonConfig := `null`
			config, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())
		})

		It("should handle large values", func() {
			jsonConfig := `{"partitionCount": 999999, "replicationFactor": 100, "load": 999.99, "pickerWidth": 1000}`
			config, err := builder.ParseConfig(json.RawMessage(jsonConfig))
			Expect(err).ToNot(HaveOccurred())
			cfg := config.(*Config)
			Expect(cfg.PartitionCount).To(Equal(999999))
			Expect(cfg.ReplicationFactor).To(Equal(100))
			Expect(cfg.Load).To(Equal(999.99))
			Expect(cfg.PickerWidth).To(Equal(1000))
		})

		It("should be thread-safe", func() {
			jsonConfig := `{"partitionCount": 100, "replicationFactor": 3}`
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					_, err := builder.ParseConfig(json.RawMessage(jsonConfig))
					Expect(err).ToNot(HaveOccurred())
					done <- true
				}()
			}
			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})

	Describe("Config", func() {
		Describe("ServiceConfigJSON", func() {
			It("should generate valid JSON with all fields", func() {
				config := &Config{
					PartitionCount:    271,
					ReplicationFactor: 20,
					Load:              1.25,
					PickerWidth:       3,
				}
				jsonString, err := config.ServiceConfigJSON()
				Expect(err).ToNot(HaveOccurred())
				Expected := "{\"loadBalancingConfig\":[{\"consistenthashing\":{\"partitionCount\":271,\"replicationFactor\":20,\"load\":1.25,\"pickerWidth\":3}}]}"
				Expect(jsonString).To(Equal(Expected))
			})

			It("should apply default values for zero fields", func() {
				config := &Config{}
				jsonString, err := config.ServiceConfigJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonString).To(ContainSubstring("partitionCount"))
				Expect(jsonString).To(ContainSubstring("replicationFactor"))
				Expect(jsonString).To(ContainSubstring("load"))
				Expect(jsonString).To(ContainSubstring("pickerWidth"))
			})

			It("should apply default values for negative fields", func() {
				config := &Config{
					PartitionCount:    -10,
					ReplicationFactor: -5,
					Load:              -1.0,
					PickerWidth:       -2,
				}
				jsonString, err := config.ServiceConfigJSON()
				Expected := fmt.Sprintf("{\"loadBalancingConfig\":[{\"consistenthashing\":{\"partitionCount\":%d,\"replicationFactor\":%d,\"load\":%g,\"pickerWidth\":%d}}]}",
					consistent.DefaultPartitionCount,
					consistent.DefaultReplicationFactor,
					consistent.DefaultLoad,
					consistent.DefaultPickerWidth)
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonString).To(Equal(Expected))
			})

			It("should handle partial configuration", func() {
				config := &Config{
					PartitionCount: 500,
					Load:           2.0,
				}
				jsonString, err := config.ServiceConfigJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonString).To(ContainSubstring("\"partitionCount\":500"))
				Expect(jsonString).To(ContainSubstring("\"load\":2"))
			})

			It("should handle maximum values", func() {
				config := &Config{
					PartitionCount:    999999,
					ReplicationFactor: 1000,
					Load:              999.99,
					PickerWidth:       100,
				}
				jsonString, err := config.ServiceConfigJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonString).To(ContainSubstring("\"partitionCount\":999999"))
				Expect(jsonString).To(ContainSubstring("\"replicationFactor\":1000"))
				Expect(jsonString).To(ContainSubstring("\"load\":999.99"))
				Expect(jsonString).To(ContainSubstring("\"pickerWidth\":100"))
			})

			It("should be deterministic", func() {
				config := &Config{
					PartitionCount:    100,
					ReplicationFactor: 3,
					Load:              1.25,
					PickerWidth:       2,
				}
				json1, err1 := config.ServiceConfigJSON()
				json2, err2 := config.ServiceConfigJSON()
				Expect(err1).ToNot(HaveOccurred())
				Expect(err2).ToNot(HaveOccurred())
				Expect(json1).To(Equal(json2))
			})
		})
	})

	Describe("ConsistentMember", func() {
		var member ConsistentMember

		BeforeEach(func() {
			member = ConsistentMember{
				SubConn: &mockSubConnWrapper{},
				name:    "test-member",
			}
		})

		It("should return correct string representation", func() {
			Expect(member.String()).To(Equal("test-member"))
		})

		It("should handle empty name", func() {
			member.name = ""
			Expect(member.String()).To(Equal(""))
		})

		It("should handle special characters in name", func() {
			member.name = "member-with-special-chars!@#$%^&*()"
			expected := "member-with-special-chars!@#$%^&*()"
			Expect(member.String()).To(Equal(expected))
		})

		It("should handle unicode characters in name", func() {
			member.name = "成员-مع-🚀"
			expected := "成员-مع-🚀"
			Expect(member.String()).To(Equal(expected))
		})

		It("should handle long name", func() {
			longName := strings.Repeat("a", 1000)
			member.name = longName
			Expect(member.String()).To(Equal(longName))
		})
	})

	Describe("Error Handling", func() {
		It("should handle nil client connection gracefully", func() {
			Expect(func() {
				builder.Build(nil, balancer.BuildOptions{})
			}).ToNot(Panic())
		})

		It("should handle concurrent ParseConfig calls", func() {
			errors := make(chan error, 100)
			for i := 0; i < 100; i++ {
				go func(i int) {
					defer GinkgoRecover()
					jsonConfig := fmt.Sprintf(`{"partitionCount": %d}`, i+1)
					_, err := builder.ParseConfig(json.RawMessage(jsonConfig))
					errors <- err
				}(i)
			}
			for i := 0; i < 100; i++ {
				Expect(<-errors).ToNot(HaveOccurred())
			}
		})

		It("should handle concurrent Build calls", func() {
			balancers := make(chan balancer.Balancer, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					bal := builder.Build(nil, balancer.BuildOptions{})
					balancers <- bal
				}()
			}
			for i := 0; i < 10; i++ {
				bal := <-balancers
				Expect(bal).ToNot(BeNil())
			}
		})
	})
})
