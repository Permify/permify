package servers

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	grpcValidate "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// oversizedTenant exceeds the 128 byte max_bytes rule shared by every
// tenant-scoped request message.
var oversizedTenant = strings.Repeat("a", 129)

// newValidationTestConn starts a gRPC server wired the same way Run() wires the
// real one: the protovalidate interceptor sits in front of every handler.
//
// The handlers are constructed with nil dependencies on purpose. Every case in
// this file must be rejected by the interceptor before any handler body runs,
// so a nil-pointer panic here is a meaningful failure signal rather than a
// fragile test setup.
func newValidationTestConn(t *testing.T) *grpc.ClientConn {
	t.Helper()

	validator, err := protovalidate.New()
	if err != nil {
		t.Fatalf("failed to build validator: %v", err)
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcValidate.UnaryServerInterceptor(validator)),
		grpc.StreamInterceptor(grpcValidate.StreamServerInterceptor(validator)),
	)

	v1.RegisterPermissionServer(srv, NewPermissionServer(&fakePermissionInvoker{}))
	v1.RegisterDataServer(srv, NewDataServer(nil, nil, nil, nil))
	v1.RegisterSchemaServer(srv, NewSchemaServer(nil, nil))
	v1.RegisterTenancyServer(srv, NewTenancyServer(nil, nil))
	v1.RegisterBundleServer(srv, NewBundleServer(nil, nil))
	v1.RegisterWatchServer(srv, NewWatchServer(nil, nil))

	lis := bufconn.Listen(1024 * 1024)
	go func() {
		_ = srv.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial test server: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		srv.Stop()
		_ = lis.Close()
	})

	return conn
}

// invalidCase describes one RPC invoked with a request that must fail
// validation. rpc is "<Service>.<Method>" and is matched against the service
// descriptors so the table cannot silently fall behind the proto.
type invalidCase struct {
	rpc  string
	call func(ctx context.Context, conn *grpc.ClientConn) error
}

func invalidCases() []invalidCase {
	return []invalidCase{
		{"Permission.Check", func(ctx context.Context, c *grpc.ClientConn) error {
			req := validPermissionCheckRequest()
			req.TenantId = oversizedTenant
			_, err := v1.NewPermissionClient(c).Check(ctx, req)
			return err
		}},
		{"Permission.BulkCheck", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewPermissionClient(c).BulkCheck(ctx, &v1.PermissionBulkCheckRequest{
				TenantId: oversizedTenant,
				Metadata: &v1.PermissionCheckRequestMetadata{Depth: 20},
				Items: []*v1.PermissionBulkCheckRequestItem{{
					Entity:     testEntity(),
					Permission: "view",
					Subject:    testSubject(),
				}},
			})
			return err
		}},
		{"Permission.Expand", func(ctx context.Context, c *grpc.ClientConn) error {
			req := validPermissionExpandRequest()
			req.TenantId = oversizedTenant
			_, err := v1.NewPermissionClient(c).Expand(ctx, req)
			return err
		}},
		{"Permission.LookupEntity", func(ctx context.Context, c *grpc.ClientConn) error {
			req := validPermissionLookupEntityRequest()
			req.TenantId = oversizedTenant
			_, err := v1.NewPermissionClient(c).LookupEntity(ctx, req)
			return err
		}},
		{"Permission.LookupEntityStream", func(ctx context.Context, c *grpc.ClientConn) error {
			req := validPermissionLookupEntityRequest()
			req.TenantId = oversizedTenant
			stream, err := v1.NewPermissionClient(c).LookupEntityStream(ctx, req)
			if err != nil {
				return err
			}
			_, err = stream.Recv()
			return err
		}},
		{"Permission.LookupSubject", func(ctx context.Context, c *grpc.ClientConn) error {
			req := validPermissionLookupSubjectRequest()
			req.TenantId = oversizedTenant
			_, err := v1.NewPermissionClient(c).LookupSubject(ctx, req)
			return err
		}},
		{"Permission.SubjectPermission", func(ctx context.Context, c *grpc.ClientConn) error {
			req := validPermissionSubjectPermissionRequest()
			req.TenantId = oversizedTenant
			_, err := v1.NewPermissionClient(c).SubjectPermission(ctx, req)
			return err
		}},

		{"Watch.Watch", func(ctx context.Context, c *grpc.ClientConn) error {
			stream, err := v1.NewWatchClient(c).Watch(ctx, &v1.WatchRequest{TenantId: oversizedTenant})
			if err != nil {
				return err
			}
			_, err = stream.Recv()
			return err
		}},

		{"Schema.Write", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewSchemaClient(c).Write(ctx, &v1.SchemaWriteRequest{
				TenantId: oversizedTenant,
				Schema:   "entity user {}",
			})
			return err
		}},
		{"Schema.PartialWrite", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewSchemaClient(c).PartialWrite(ctx, &v1.SchemaPartialWriteRequest{
				TenantId: oversizedTenant,
				Metadata: &v1.SchemaPartialWriteRequestMetadata{SchemaVersion: "v1"},
			})
			return err
		}},
		{"Schema.Read", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewSchemaClient(c).Read(ctx, &v1.SchemaReadRequest{
				TenantId: oversizedTenant,
				Metadata: &v1.SchemaReadRequestMetadata{SchemaVersion: "v1"},
			})
			return err
		}},
		{"Schema.List", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewSchemaClient(c).List(ctx, &v1.SchemaListRequest{TenantId: oversizedTenant})
			return err
		}},

		{"Data.Write", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewDataClient(c).Write(ctx, &v1.DataWriteRequest{
				TenantId: oversizedTenant,
				Metadata: &v1.DataWriteRequestMetadata{SchemaVersion: "v1"},
			})
			return err
		}},
		{"Data.WriteRelationships", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewDataClient(c).WriteRelationships(ctx, &v1.RelationshipWriteRequest{
				TenantId: oversizedTenant,
				Metadata: &v1.RelationshipWriteRequestMetadata{SchemaVersion: "v1"},
				Tuples:   []*v1.Tuple{{Entity: testEntity(), Relation: "viewer", Subject: testSubject()}},
			})
			return err
		}},
		{"Data.ReadRelationships", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewDataClient(c).ReadRelationships(ctx, &v1.RelationshipReadRequest{
				TenantId: oversizedTenant,
				Metadata: &v1.RelationshipReadRequestMetadata{SnapToken: "token"},
				Filter:   &v1.TupleFilter{},
			})
			return err
		}},
		{"Data.ReadAttributes", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewDataClient(c).ReadAttributes(ctx, &v1.AttributeReadRequest{
				TenantId: oversizedTenant,
				Metadata: &v1.AttributeReadRequestMetadata{SnapToken: "token"},
				Filter:   &v1.AttributeFilter{},
			})
			return err
		}},
		{"Data.Delete", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewDataClient(c).Delete(ctx, &v1.DataDeleteRequest{
				TenantId:        oversizedTenant,
				TupleFilter:     &v1.TupleFilter{},
				AttributeFilter: &v1.AttributeFilter{},
			})
			return err
		}},
		{"Data.DeleteRelationships", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewDataClient(c).DeleteRelationships(ctx, &v1.RelationshipDeleteRequest{
				TenantId: oversizedTenant,
				Filter:   &v1.TupleFilter{},
			})
			return err
		}},
		{"Data.RunBundle", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewDataClient(c).RunBundle(ctx, &v1.BundleRunRequest{
				TenantId: oversizedTenant,
				Name:     "bundle",
			})
			return err
		}},

		{"Bundle.Write", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewBundleClient(c).Write(ctx, &v1.BundleWriteRequest{
				TenantId: oversizedTenant,
				Bundles:  []*v1.DataBundle{{Name: "bundle"}},
			})
			return err
		}},
		{"Bundle.Read", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewBundleClient(c).Read(ctx, &v1.BundleReadRequest{
				TenantId: oversizedTenant,
				Name:     "bundle",
			})
			return err
		}},
		{"Bundle.Delete", func(ctx context.Context, c *grpc.ClientConn) error {
			_, err := v1.NewBundleClient(c).Delete(ctx, &v1.BundleDeleteRequest{
				TenantId: oversizedTenant,
				Name:     "bundle",
			})
			return err
		}},

		{"Tenancy.Create", func(ctx context.Context, c *grpc.ClientConn) error {
			// TenantCreateRequest validates id, not tenant_id. The pattern is
			// unanchored, so the value has to contain no allowed character at all.
			_, err := v1.NewTenancyClient(c).Create(ctx, &v1.TenantCreateRequest{
				Id:   "!!!",
				Name: "tenant",
			})
			return err
		}},
	}
}

// rpcsWithoutRequestRules lists RPCs whose request message carries no
// validation rules that any input can violate, so there is nothing for the
// interceptor to reject.
//
//	TenantDeleteRequest.id    - unconstrained (protoc-gen-validate had no rules
//	                            here either, only an inert ignore_empty: false)
//	TenantListRequest         - page_size is gte:1 but ignored when zero, and
//	                            uint32 cannot be negative; continuous_token is
//	                            unconstrained
var rpcsWithoutRequestRules = map[string]string{
	"Tenancy.Delete": "TenantDeleteRequest has no field rules",
	"Tenancy.List":   "TenantListRequest has no violable field rules",
}

// TestRequestValidationOverGRPC drives every RPC through a real client and
// asserts that malformed input is rejected with InvalidArgument before it
// reaches a handler.
func TestRequestValidationOverGRPC(t *testing.T) {
	conn := newValidationTestConn(t)

	for _, tc := range invalidCases() {
		t.Run(tc.rpc, func(t *testing.T) {
			err := tc.call(context.Background(), conn)
			if err == nil {
				t.Fatal("expected the request to be rejected")
			}
			if got := status.Code(err); got != codes.InvalidArgument {
				t.Fatalf("status = %v, want %v (err: %v)", got, codes.InvalidArgument, err)
			}
		})
	}
}

// TestValidationErrorCarriesViolationDetails locks in the structured violations
// the interceptor attaches alongside the status message, so clients can read
// which field failed instead of parsing free-form text.
func TestValidationErrorCarriesViolationDetails(t *testing.T) {
	conn := newValidationTestConn(t)

	req := validPermissionCheckRequest()
	req.TenantId = oversizedTenant
	_, err := v1.NewPermissionClient(conn).Check(context.Background(), req)
	if err == nil {
		t.Fatal("expected the request to be rejected")
	}

	details := status.Convert(err).Details()
	if len(details) == 0 {
		t.Fatal("expected violation details on the status")
	}

	var rendered []string
	sawViolations := false
	for _, detail := range details {
		msg, ok := detail.(proto.Message)
		if !ok {
			continue
		}
		if msg.ProtoReflect().Descriptor().FullName() == "buf.validate.Violations" {
			sawViolations = true
		}
		rendered = append(rendered, prototext.Format(msg))
	}

	if !sawViolations {
		t.Fatalf("expected a buf.validate.Violations detail, got %v", details)
	}
	if joined := strings.Join(rendered, "\n"); !strings.Contains(joined, "tenant_id") {
		t.Fatalf("expected the violation to name tenant_id, got:\n%s", joined)
	}
}

// TestBulkCheckRejectsInvalidItemOverTheWire pins down that one malformed item
// fails the whole BulkCheck call rather than coming back as a single DENIED
// result.
//
// This matches the behaviour before the protovalidate migration: the
// protoc-gen-validate code generated for PermissionBulkCheckRequest also looped
// over items and called Validate on each, so the old interceptor rejected the
// request too. The per-item check that remains in the handler is only reachable
// when the handler is called directly, not over gRPC.
func TestBulkCheckRejectsInvalidItemOverTheWire(t *testing.T) {
	conn := newValidationTestConn(t)

	_, err := v1.NewPermissionClient(conn).BulkCheck(context.Background(), &v1.PermissionBulkCheckRequest{
		TenantId: "t1",
		Metadata: &v1.PermissionCheckRequestMetadata{Depth: 20},
		Items: []*v1.PermissionBulkCheckRequestItem{
			{Entity: testEntity(), Permission: "view", Subject: testSubject()},
			{Entity: testEntity(), Permission: "", Subject: testSubject()},
		},
	})
	if err == nil {
		t.Fatal("expected the request to be rejected")
	}
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("status = %v, want %v (err: %v)", got, codes.InvalidArgument, err)
	}
}

// TestValidationErrorMapsToHTTPBadRequest covers the grpc-gateway surface: the
// InvalidArgument the interceptor returns has to reach HTTP clients as 400, not
// as a 500. GetStatus() used to swallow validation errors into Internal, so
// this guards the regression that would turn malformed input into a server
// error.
func TestValidationErrorMapsToHTTPBadRequest(t *testing.T) {
	conn := newValidationTestConn(t)

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{UseProtoNames: true},
		}),
	)
	if err := v1.RegisterPermissionHandler(context.Background(), mux, conn); err != nil {
		t.Fatalf("failed to register gateway handler: %v", err)
	}

	gateway := httptest.NewServer(mux)
	t.Cleanup(gateway.Close)

	body := `{"metadata":{"depth":20},"entity":{"type":"document","id":"1"},` +
		`"permission":"view","subject":{"type":"user","id":"1"}}`
	url := gateway.URL + "/v1/tenants/" + oversizedTenant + "/permissions/check"

	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("gateway request failed: %v", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read gateway response: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", resp.StatusCode, http.StatusBadRequest, payload)
	}
	if !strings.Contains(string(payload), "tenant_id") {
		t.Fatalf("expected the response to name tenant_id, got: %s", payload)
	}
}

// TestEveryRPCIsCoveredByValidationTable walks the service descriptors and
// fails when an RPC is neither exercised by invalidCases nor recorded in
// rpcsWithoutRequestRules. Adding an RPC therefore forces a decision about its
// request validation instead of letting it ship unchecked.
func TestEveryRPCIsCoveredByValidationTable(t *testing.T) {
	covered := make(map[string]bool, len(invalidCases()))
	for _, tc := range invalidCases() {
		if covered[tc.rpc] {
			t.Fatalf("duplicate entry for %s", tc.rpc)
		}
		covered[tc.rpc] = true
	}

	services := []protoreflect.ServiceDescriptor{
		v1.File_base_v1_service_proto.Services().ByName("Permission"),
		v1.File_base_v1_service_proto.Services().ByName("Watch"),
		v1.File_base_v1_service_proto.Services().ByName("Schema"),
		v1.File_base_v1_service_proto.Services().ByName("Data"),
		v1.File_base_v1_service_proto.Services().ByName("Bundle"),
		v1.File_base_v1_service_proto.Services().ByName("Tenancy"),
	}

	for _, svc := range services {
		if svc == nil {
			t.Fatal("service descriptor not found; the proto layout changed")
		}
		methods := svc.Methods()
		for i := range methods.Len() {
			rpc := string(svc.Name()) + "." + string(methods.Get(i).Name())
			if covered[rpc] {
				continue
			}
			if reason, ok := rpcsWithoutRequestRules[rpc]; ok {
				t.Logf("%s skipped: %s", rpc, reason)
				continue
			}
			t.Errorf("%s has no validation coverage: add it to invalidCases or to rpcsWithoutRequestRules", rpc)
		}
	}
}
