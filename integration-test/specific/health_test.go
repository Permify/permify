package specific

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var _ = Describe("health-test", func() {
	Context("Health", func() {
		It("Health: Success", func() {
			// Set up a connection to the server.
			conn, err := grpc.DialContext(context.Background(), "permify:3478", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				Expect(err).ShouldNot(HaveOccurred())
			}

			Expect(err).ShouldNot(HaveOccurred())

			healthClient := grpc_health_v1.NewHealthClient(conn)

			res, err := healthClient.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Status).Should(Equal(grpc_health_v1.HealthCheckResponse_SERVING))

			err = conn.Close()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
