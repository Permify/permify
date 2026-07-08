package servers

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

type fakeTenantStore struct {
	createErr error
	deleteErr error
	listErr   error

	createdID   string
	createdName string
	deletedID   string
	tenants     []*v1.Tenant
}

func (f *fakeTenantStore) CreateTenant(_ context.Context, id, name string) (*v1.Tenant, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.createdID = id
	f.createdName = name
	return &v1.Tenant{Id: id, Name: name}, nil
}

func (f *fakeTenantStore) DeleteTenant(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedID = id
	return nil
}

func (f *fakeTenantStore) ListTenants(_ context.Context, _ database.Pagination) ([]*v1.Tenant, database.EncodedContinuousToken, error) {
	if f.listErr != nil {
		return nil, nil, f.listErr
	}
	return f.tenants, database.NewNoopContinuousToken().Encode(), nil
}

type fakeBundleStore struct {
	writeErr  error
	readErr   error
	deleteErr error

	written       []storage.Bundle
	deletedTenant string
	deletedName   string
	bundles       map[string]*v1.DataBundle
}

func newFakeBundleStore() *fakeBundleStore {
	return &fakeBundleStore{bundles: map[string]*v1.DataBundle{}}
}

func (f *fakeBundleStore) Write(_ context.Context, bundles []storage.Bundle) ([]string, error) {
	if f.writeErr != nil {
		return nil, f.writeErr
	}
	f.written = append(f.written, bundles...)

	names := make([]string, 0, len(bundles))
	for _, bundle := range bundles {
		names = append(names, bundle.Name)
		f.bundles[bundle.TenantID+"/"+bundle.Name] = bundle.DataBundle
	}
	return names, nil
}

func (f *fakeBundleStore) Read(_ context.Context, tenantID, name string) (*v1.DataBundle, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	return f.bundles[tenantID+"/"+name], nil
}

func (f *fakeBundleStore) Delete(_ context.Context, tenantID, name string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedTenant = tenantID
	f.deletedName = name
	delete(f.bundles, tenantID+"/"+name)
	return nil
}

type fakePermissionInvoker struct {
	err error

	checkReq             *v1.PermissionCheckRequest
	expandReq            *v1.PermissionExpandRequest
	lookupEntityReq      *v1.PermissionLookupEntityRequest
	lookupSubjectReq     *v1.PermissionLookupSubjectRequest
	subjectPermissionReq *v1.PermissionSubjectPermissionRequest
}

func (f *fakePermissionInvoker) Check(_ context.Context, request *v1.PermissionCheckRequest) (*v1.PermissionCheckResponse, error) {
	f.checkReq = request
	if f.err != nil {
		return nil, f.err
	}
	return &v1.PermissionCheckResponse{
		Can: v1.CheckResult_CHECK_RESULT_ALLOWED,
		Metadata: &v1.PermissionCheckResponseMetadata{
			CheckCount: 1,
		},
	}, nil
}

func (f *fakePermissionInvoker) Expand(_ context.Context, request *v1.PermissionExpandRequest) (*v1.PermissionExpandResponse, error) {
	f.expandReq = request
	if f.err != nil {
		return nil, f.err
	}
	return &v1.PermissionExpandResponse{}, nil
}

func (f *fakePermissionInvoker) LookupEntity(_ context.Context, request *v1.PermissionLookupEntityRequest) (*v1.PermissionLookupEntityResponse, error) {
	f.lookupEntityReq = request
	if f.err != nil {
		return nil, f.err
	}
	return &v1.PermissionLookupEntityResponse{
		EntityIds:       []string{"document-1"},
		ContinuousToken: "next-entities",
	}, nil
}

func (f *fakePermissionInvoker) LookupEntityStream(_ context.Context, request *v1.PermissionLookupEntityRequest, _ v1.Permission_LookupEntityStreamServer) error {
	f.lookupEntityReq = request
	return f.err
}

func (f *fakePermissionInvoker) LookupSubject(_ context.Context, request *v1.PermissionLookupSubjectRequest) (*v1.PermissionLookupSubjectResponse, error) {
	f.lookupSubjectReq = request
	if f.err != nil {
		return nil, f.err
	}
	return &v1.PermissionLookupSubjectResponse{
		SubjectIds:      []string{"user-1"},
		ContinuousToken: "next-subjects",
	}, nil
}

func (f *fakePermissionInvoker) SubjectPermission(_ context.Context, request *v1.PermissionSubjectPermissionRequest) (*v1.PermissionSubjectPermissionResponse, error) {
	f.subjectPermissionReq = request
	if f.err != nil {
		return nil, f.err
	}
	return &v1.PermissionSubjectPermissionResponse{
		Results: map[string]v1.CheckResult{
			"view": v1.CheckResult_CHECK_RESULT_ALLOWED,
		},
	}, nil
}

type testContextKey struct{}

func testEntity() *v1.Entity {
	return &v1.Entity{Type: "document", Id: "document-1"}
}

func testSubject() *v1.Subject {
	return &v1.Subject{Type: "user", Id: "user-1"}
}

func validPermissionCheckRequest() *v1.PermissionCheckRequest {
	return &v1.PermissionCheckRequest{
		TenantId:   "tenant-1",
		Metadata:   &v1.PermissionCheckRequestMetadata{SchemaVersion: "schema-1", SnapToken: "snap-1", Depth: 3},
		Entity:     testEntity(),
		Permission: "view",
		Subject:    testSubject(),
	}
}

func validPermissionExpandRequest() *v1.PermissionExpandRequest {
	return &v1.PermissionExpandRequest{
		TenantId:   "tenant-1",
		Metadata:   &v1.PermissionExpandRequestMetadata{SchemaVersion: "schema-1", SnapToken: "snap-1"},
		Entity:     testEntity(),
		Permission: "view",
	}
}

func validPermissionLookupEntityRequest() *v1.PermissionLookupEntityRequest {
	return &v1.PermissionLookupEntityRequest{
		TenantId:   "tenant-1",
		Metadata:   &v1.PermissionLookupEntityRequestMetadata{SchemaVersion: "schema-1", SnapToken: "snap-1", Depth: 3},
		EntityType: "document",
		Permission: "view",
		Subject:    testSubject(),
		PageSize:   10,
	}
}

func validPermissionLookupSubjectRequest() *v1.PermissionLookupSubjectRequest {
	return &v1.PermissionLookupSubjectRequest{
		TenantId:   "tenant-1",
		Metadata:   &v1.PermissionLookupSubjectRequestMetadata{SchemaVersion: "schema-1", SnapToken: "snap-1", Depth: 3},
		Entity:     testEntity(),
		Permission: "view",
		SubjectReference: &v1.RelationReference{
			Type:     "user",
			Relation: "member",
		},
		PageSize: 10,
	}
}

func validPermissionSubjectPermissionRequest() *v1.PermissionSubjectPermissionRequest {
	return &v1.PermissionSubjectPermissionRequest{
		TenantId: "tenant-1",
		Metadata: &v1.PermissionSubjectPermissionRequestMetadata{
			SchemaVersion: "schema-1",
			SnapToken:     "snap-1",
			Depth:         3,
		},
		Entity:  testEntity(),
		Subject: testSubject(),
	}
}

func TestHealthServer(t *testing.T) {
	server := NewHealthServer()
	if server == nil {
		t.Fatal("expected health server")
	}

	resp, err := server.Check(context.Background(), &health.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("unexpected health check error: %v", err)
	}
	if resp.GetStatus() != health.HealthCheckResponse_SERVING {
		t.Fatalf("expected serving status, got %v", resp.GetStatus())
	}

	err = server.Watch(&health.HealthCheckRequest{}, nil)
	if status.Code(err) != codes.Unimplemented {
		t.Fatalf("expected unimplemented watch status, got %v", status.Code(err))
	}

	ctx := context.WithValue(context.Background(), testContextKey{}, "value")
	gotCtx, err := server.AuthFuncOverride(ctx, "/grpc.health.v1.Health/Check")
	if err != nil {
		t.Fatalf("unexpected auth override error: %v", err)
	}
	if gotCtx != ctx {
		t.Fatal("expected auth override to return the original context")
	}
}

func TestNewContainerWiresDependencies(t *testing.T) {
	dr := storage.NewNoopRelationshipReader()
	dw := storage.NewNoopDataWriter()
	br := storage.NewNoopBundleReader()
	bw := storage.NewNoopBundleWriter()
	sr := storage.NewNoopSchemaReader()
	sw := storage.NewNoopSchemaWriter()
	tr := storage.NewNoopTenantReader()
	tw := storage.NewNoopTenantWriter()
	w := storage.NewNoopWatcher()

	container := NewContainer(nil, dr, dw, br, bw, sr, sw, tr, tw, w)
	if container == nil {
		t.Fatal("expected container")
	}

	if container.Invoker != nil || container.DR != dr || container.DW != dw || container.BR != br ||
		container.BW != bw || container.SR != sr || container.SW != sw || container.TR != tr ||
		container.TW != tw || container.W != w {
		t.Fatal("container did not preserve constructor dependencies")
	}
}

func TestHTTPNameFormatter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	if got := httpNameFormatter("", req); got != "GET <not found>" {
		t.Fatalf("expected fallback span name, got %q", got)
	}

	var got string
	mux := gwruntime.NewServeMux()
	err := mux.HandlePath(http.MethodPost, "/v1/tenants/{tenant_id}/schemas/write", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		got = httpNameFormatter("", r)
		w.WriteHeader(http.StatusNoContent)
	})
	if err != nil {
		t.Fatalf("failed to register gateway path: %v", err)
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/tenants/t1/schemas/write", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected mux handler to run, got status %d", recorder.Code)
	}

	if got != "POST /v1/tenants/{tenant_id=*}/schemas/write" {
		t.Fatalf("expected annotated span name, got %q", got)
	}
}

func TestInterceptorLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	adapter := InterceptorLogger(logger)
	adapter.Log(context.Background(), logging.LevelInfo, "request complete", "method", "Check")

	out := buf.String()
	if !strings.Contains(out, "request complete") || !strings.Contains(out, "method=Check") {
		t.Fatalf("expected adapted log output to contain message and field, got %q", out)
	}
}

func TestPermissionServerPassesThroughInvoker(t *testing.T) {
	invoker := &fakePermissionInvoker{}
	server := NewPermissionServer(invoker)
	if server == nil {
		t.Fatal("expected permission server")
	}

	checkReq := validPermissionCheckRequest()
	checkResp, err := server.Check(context.Background(), checkReq)
	if err != nil {
		t.Fatalf("unexpected check error: %v", err)
	}
	if invoker.checkReq != checkReq || checkResp.GetCan() != v1.CheckResult_CHECK_RESULT_ALLOWED {
		t.Fatalf("check did not return invoker response")
	}

	expandReq := validPermissionExpandRequest()
	expandResp, err := server.Expand(context.Background(), expandReq)
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}
	if invoker.expandReq != expandReq || expandResp == nil {
		t.Fatalf("expand did not return invoker response")
	}

	lookupEntityReq := validPermissionLookupEntityRequest()
	lookupEntityResp, err := server.LookupEntity(context.Background(), lookupEntityReq)
	if err != nil {
		t.Fatalf("unexpected lookup entity error: %v", err)
	}
	if invoker.lookupEntityReq != lookupEntityReq || !reflect.DeepEqual(lookupEntityResp.GetEntityIds(), []string{"document-1"}) {
		t.Fatalf("lookup entity did not return invoker response")
	}

	lookupSubjectReq := validPermissionLookupSubjectRequest()
	lookupSubjectResp, err := server.LookupSubject(context.Background(), lookupSubjectReq)
	if err != nil {
		t.Fatalf("unexpected lookup subject error: %v", err)
	}
	if invoker.lookupSubjectReq != lookupSubjectReq || !reflect.DeepEqual(lookupSubjectResp.GetSubjectIds(), []string{"user-1"}) {
		t.Fatalf("lookup subject did not return invoker response")
	}

	subjectPermissionReq := validPermissionSubjectPermissionRequest()
	subjectPermissionResp, err := server.SubjectPermission(context.Background(), subjectPermissionReq)
	if err != nil {
		t.Fatalf("unexpected subject permission error: %v", err)
	}
	if invoker.subjectPermissionReq != subjectPermissionReq ||
		subjectPermissionResp.GetResults()["view"] != v1.CheckResult_CHECK_RESULT_ALLOWED {
		t.Fatalf("subject permission did not return invoker response")
	}
}

func TestPermissionServerValidationAndInvokerErrors(t *testing.T) {
	invoker := &fakePermissionInvoker{}
	server := NewPermissionServer(invoker)

	_, err := server.Check(context.Background(), &v1.PermissionCheckRequest{})
	if err == nil {
		t.Fatal("expected invalid check request to fail")
	}
	if invoker.checkReq != nil {
		t.Fatal("invalid check request should not reach invoker")
	}

	invoker.err = errors.New(v1.ErrorCode_ERROR_CODE_NOT_FOUND.String())
	_, err = server.Check(context.Background(), validPermissionCheckRequest())
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected not found status, got %v", status.Code(err))
	}
}

func TestTenancyServer(t *testing.T) {
	store := &fakeTenantStore{
		tenants: []*v1.Tenant{{Id: "tenant-1", Name: "Tenant One"}},
	}
	server := NewTenancyServer(store, store)
	if server == nil {
		t.Fatal("expected tenancy server")
	}

	created, err := server.Create(context.Background(), &v1.TenantCreateRequest{
		Id:   "tenant-2",
		Name: "Tenant Two",
	})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if store.createdID != "tenant-2" || store.createdName != "Tenant Two" {
		t.Fatalf("tenant create was called with %q/%q", store.createdID, store.createdName)
	}
	if created.GetTenant().GetId() != "tenant-2" || created.GetTenant().GetName() != "Tenant Two" {
		t.Fatalf("unexpected create response: %v", created.GetTenant())
	}

	deleted, err := server.Delete(context.Background(), &v1.TenantDeleteRequest{Id: "tenant-2"})
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if store.deletedID != "tenant-2" || deleted.GetTenantId() != "tenant-2" {
		t.Fatalf("delete did not return deleted tenant id")
	}

	listed, err := server.List(context.Background(), &v1.TenantListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if !reflect.DeepEqual(listed.GetTenants(), store.tenants) {
		t.Fatalf("unexpected tenants: %v", listed.GetTenants())
	}
	if listed.GetContinuousToken() != "" {
		t.Fatalf("expected empty noop continuous token, got %q", listed.GetContinuousToken())
	}
}

func TestTenancyServerStorageError(t *testing.T) {
	store := &fakeTenantStore{
		createErr: errors.New(v1.ErrorCode_ERROR_CODE_NOT_FOUND.String()),
	}
	server := NewTenancyServer(store, store)

	_, err := server.Create(context.Background(), &v1.TenantCreateRequest{Id: "tenant-1", Name: "Tenant One"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected not found status, got %v", status.Code(err))
	}
}

func TestBundleServer(t *testing.T) {
	store := newFakeBundleStore()
	server := NewBundleServer(store, store)
	if server == nil {
		t.Fatal("expected bundle server")
	}

	bundle := &v1.DataBundle{
		Name: "starter",
		Operations: []*v1.Operation{{
			RelationshipsWrite: []string{"organization:1#admin@user:1"},
			AttributesWrite:    []string{"organization:1$active|boolean:true"},
		}},
	}

	written, err := server.Write(context.Background(), &v1.BundleWriteRequest{
		TenantId: "tenant-1",
		Bundles:  []*v1.DataBundle{bundle},
	})
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if !reflect.DeepEqual(written.GetNames(), []string{"starter"}) {
		t.Fatalf("unexpected written names: %v", written.GetNames())
	}
	if len(store.written) != 1 || store.written[0].TenantID != "tenant-1" || store.written[0].Name != "starter" {
		t.Fatalf("unexpected stored bundle write: %+v", store.written)
	}

	read, err := server.Read(context.Background(), &v1.BundleReadRequest{
		TenantId: "tenant-1",
		Name:     "starter",
	})
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if read.GetBundle() != bundle {
		t.Fatalf("expected stored bundle, got %v", read.GetBundle())
	}

	deleted, err := server.Delete(context.Background(), &v1.BundleDeleteRequest{
		TenantId: "tenant-1",
		Name:     "starter",
	})
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if deleted.GetName() != "starter" || store.deletedTenant != "tenant-1" || store.deletedName != "starter" {
		t.Fatalf("delete did not return/delete expected bundle")
	}
}

func TestBundleServerValidationAndStorageErrors(t *testing.T) {
	server := NewBundleServer(newFakeBundleStore(), newFakeBundleStore())
	duplicateBundle := &v1.DataBundle{
		Name: "duplicate",
		Operations: []*v1.Operation{{
			RelationshipsWrite: []string{"organization:1#admin@user:1", "organization:1#admin@user:1"},
		}},
	}
	_, err := server.Write(context.Background(), &v1.BundleWriteRequest{
		TenantId: "tenant-1",
		Bundles:  []*v1.DataBundle{duplicateBundle},
	})
	if err == nil {
		t.Fatal("expected duplicate bundle operation to fail validation")
	}

	store := newFakeBundleStore()
	store.readErr = errors.New(v1.ErrorCode_ERROR_CODE_NOT_FOUND.String())
	server = NewBundleServer(store, store)

	_, err = server.Read(context.Background(), &v1.BundleReadRequest{
		TenantId: "tenant-1",
		Name:     "missing",
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected not found status, got %v", status.Code(err))
	}
}
