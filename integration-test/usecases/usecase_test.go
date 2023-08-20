package usecases

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/integration-test/usecases/shapes"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

func TestUsecase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "usecase-suite")
}

// Set up a connection to the server.
var conn *grpc.ClientConn

// Create a PermissionClient using the connection.
var (
	schemaClient     base.SchemaClient
	permissionClient base.PermissionClient
	dataClient       base.DataClient
)

var (
	// NOTION
	initialNotionSchemaVersion string
	initialNotionSnapToken     string

	// GOOGLE DOCS
	initialGoogleDocsSchemaVersion string
	initialGoogleDocsSnapToken     string

	// GOOGLE DOCS
	initialFacebookGroupsSchemaVersion string
	initialFacebookGroupsSnapToken     string
)

var _ = BeforeSuite(func() {
	ctx := context.Background()

	var err error
	// Set up a connection to the server.
	conn, err = grpc.DialContext(ctx, "permify:3478", grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).ShouldNot(HaveOccurred())

	// Create a PermissionClient using the connection.
	schemaClient = base.NewSchemaClient(conn)
	permissionClient = base.NewPermissionClient(conn)
	dataClient = base.NewDataClient(conn)

	// NOTION

	// WRITE SCHEMA

	nsw, err := schemaClient.Write(ctx, &base.SchemaWriteRequest{
		TenantId: "notion",
		Schema:   shapes.InitialNotionShape.Schema,
	})

	Expect(err).ShouldNot(HaveOccurred())

	initialNotionSchemaVersion = nsw.SchemaVersion

	// WRITE RELATIONSHIPS

	var notionTuples []*base.Tuple

	for _, t := range shapes.InitialNotionShape.Relationships {
		tup, err := tuple.Tuple(t)
		if err != nil {
			continue
		}

		notionTuples = append(notionTuples, tup)
	}

	nrw, err := dataClient.Write(ctx, &base.DataWriteRequest{
		Metadata: &base.DataWriteRequestMetadata{
			SchemaVersion: initialNotionSchemaVersion,
		},
		TenantId: "notion",
		Tuples:   notionTuples,
	})

	Expect(err).ShouldNot(HaveOccurred())

	initialNotionSnapToken = nrw.SnapToken

	// DOCS

	// WRITE SCHEMA

	dsw, err := schemaClient.Write(ctx, &base.SchemaWriteRequest{
		TenantId: "google-docs",
		Schema:   shapes.InitialGoogleDocsShape.Schema,
	})

	Expect(err).ShouldNot(HaveOccurred())

	initialGoogleDocsSchemaVersion = dsw.SchemaVersion

	// WRITE RELATIONSHIPS

	var googleDocsTuples []*base.Tuple

	for _, t := range shapes.InitialGoogleDocsShape.Relationships {
		tup, err := tuple.Tuple(t)
		if err != nil {
			continue
		}

		googleDocsTuples = append(googleDocsTuples, tup)
	}

	drw, err := dataClient.Write(ctx, &base.DataWriteRequest{
		Metadata: &base.DataWriteRequestMetadata{
			SchemaVersion: initialGoogleDocsSchemaVersion,
		},
		TenantId: "google-docs",
		Tuples:   googleDocsTuples,
	})

	Expect(err).ShouldNot(HaveOccurred())

	initialGoogleDocsSnapToken = drw.SnapToken

	// FACEBOOK GROUPS

	// WRITE SCHEMA

	fsw, err := schemaClient.Write(ctx, &base.SchemaWriteRequest{
		TenantId: "facebook-groups",
		Schema:   shapes.InitialFacebookGroupsShape.Schema,
	})

	Expect(err).ShouldNot(HaveOccurred())

	initialFacebookGroupsSchemaVersion = fsw.SchemaVersion

	// WRITE RELATIONSHIPS

	var facebookGroupsTuples []*base.Tuple

	for _, t := range shapes.InitialFacebookGroupsShape.Relationships {
		tup, err := tuple.Tuple(t)
		if err != nil {
			Expect(err).ShouldNot(HaveOccurred())
		}

		facebookGroupsTuples = append(facebookGroupsTuples, tup)
	}

	frw, err := dataClient.Write(ctx, &base.DataWriteRequest{
		Metadata: &base.DataWriteRequestMetadata{
			SchemaVersion: initialFacebookGroupsSchemaVersion,
		},
		TenantId: "facebook-groups",
		Tuples:   facebookGroupsTuples,
	})

	Expect(err).ShouldNot(HaveOccurred())

	initialFacebookGroupsSnapToken = frw.SnapToken
})

var _ = AfterSuite(func() {
	err := conn.Close()
	Expect(err).ShouldNot(HaveOccurred())
})
