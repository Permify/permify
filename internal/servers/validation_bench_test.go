package servers

import (
	"testing"

	"buf.build/go/protovalidate"

	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// These benchmarks measure request validation in isolation, which is the step
// that changed when protoc-gen-validate (generated Go branches) was replaced by
// protovalidate (CEL evaluated over protoreflect).
//
// Compare the numbers against the latency of a cache-hit Check: on the cold
// path the storage round-trip dwarfs validation, so a regression only matters
// relative to the fast path.

func benchValidator(b *testing.B) protovalidate.Validator {
	b.Helper()

	validator, err := protovalidate.New()
	if err != nil {
		b.Fatalf("failed to build validator: %v", err)
	}
	return validator
}

func benchCheckRequest() *v1.PermissionCheckRequest {
	return &v1.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &v1.PermissionCheckRequestMetadata{
			SchemaVersion: "v1",
			SnapToken:     "snap",
			Depth:         20,
		},
		Entity:     &v1.Entity{Type: "document", Id: "1"},
		Permission: "view",
		Subject:    &v1.Subject{Type: "user", Id: "1"},
	}
}

func benchBulkCheckRequest(items int) *v1.PermissionBulkCheckRequest {
	req := &v1.PermissionBulkCheckRequest{
		TenantId: "t1",
		Metadata: &v1.PermissionCheckRequestMetadata{
			SchemaVersion: "v1",
			SnapToken:     "snap",
			Depth:         20,
		},
		Items: make([]*v1.PermissionBulkCheckRequestItem, 0, items),
	}
	for range items {
		req.Items = append(req.Items, &v1.PermissionBulkCheckRequestItem{
			Entity:     &v1.Entity{Type: "document", Id: "1"},
			Permission: "view",
			Subject:    &v1.Subject{Type: "user", Id: "1"},
		})
	}
	return req
}

// BenchmarkValidateCheckRequest is the hot path: one validation per Check.
func BenchmarkValidateCheckRequest(b *testing.B) {
	validator := benchValidator(b)
	req := benchCheckRequest()

	// Warm the per-message CEL program cache so the first iteration does not
	// absorb compilation cost.
	if err := validator.Validate(req); err != nil {
		b.Fatalf("unexpected validation error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := validator.Validate(req); err != nil {
			b.Fatalf("unexpected validation error: %v", err)
		}
	}
}

// BenchmarkValidateBulkCheckRequest covers the top-level BulkCheck message,
// which the interceptor validates once including all nested items.
func BenchmarkValidateBulkCheckRequest(b *testing.B) {
	validator := benchValidator(b)
	req := benchBulkCheckRequest(100)

	if err := validator.Validate(req); err != nil {
		b.Fatalf("unexpected validation error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := validator.Validate(req); err != nil {
			b.Fatalf("unexpected validation error: %v", err)
		}
	}
}

// BenchmarkValidateBulkCheckItems mirrors the per-item validation that stays in
// the BulkCheck handler, where a failure becomes a DENIED result rather than an
// error and therefore cannot move to the interceptor.
func BenchmarkValidateBulkCheckItems(b *testing.B) {
	validator := benchValidator(b)
	items := benchBulkCheckRequest(100).GetItems()

	if err := validator.Validate(items[0]); err != nil {
		b.Fatalf("unexpected validation error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		for _, item := range items {
			if err := validator.Validate(item); err != nil {
				b.Fatalf("unexpected validation error: %v", err)
			}
		}
	}
}

// BenchmarkValidateCheckRequestParallel approximates concurrent traffic, where
// a shared validator must not become a contention point.
func BenchmarkValidateCheckRequestParallel(b *testing.B) {
	validator := benchValidator(b)
	req := benchCheckRequest()

	if err := validator.Validate(req); err != nil {
		b.Fatalf("unexpected validation error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := validator.Validate(req); err != nil {
				b.Fatalf("unexpected validation error: %v", err)
			}
		}
	})
}
