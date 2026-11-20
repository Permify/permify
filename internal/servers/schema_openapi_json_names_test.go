package servers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestSchemaReadHTTPJSONNamesMatchOpenAPI(t *testing.T) {
	resp := &v1.SchemaReadResponse{
		Schema: &v1.SchemaDefinition{
			EntityDefinitions: map[string]*v1.EntityDefinition{
				"user": {Name: "user"},
			},
			RuleDefinitions: map[string]*v1.RuleDefinition{
				"is_weekday": {Name: "is_weekday"},
			},
			References: map[string]v1.SchemaDefinition_Reference{
				"user":       v1.SchemaDefinition_REFERENCE_ENTITY,
				"is_weekday": v1.SchemaDefinition_REFERENCE_RULE,
			},
		},
	}

	marshaler := &gwruntime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	raw, err := marshaler.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal schema read response: %v", err)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode marshaled response: %v", err)
	}

	schemaRaw, ok := body["schema"]
	if !ok {
		t.Fatalf("HTTP schema read JSON missing schema object: %s", raw)
	}

	var schema map[string]json.RawMessage
	if err := json.Unmarshal(schemaRaw, &schema); err != nil {
		t.Fatalf("decode schema object: %v", err)
	}

	for _, name := range []string{"entity_definitions", "rule_definitions"} {
		if _, ok := schema[name]; !ok {
			t.Errorf("HTTP schema JSON missing %q; keys=%v body=%s", name, jsonKeys(schema), raw)
		}
	}
	for _, name := range []string{"entityDefinitions", "ruleDefinitions"} {
		if _, ok := schema[name]; ok {
			t.Errorf("HTTP schema JSON unexpectedly used camelCase %q; body=%s", name, raw)
		}
	}

	root := findRepoRoot(t)
	specs := []struct {
		path  string
		props func(map[string]any) map[string]any
	}{
		{
			path: filepath.Join(root, "docs/api-reference/openapi.json"),
			props: func(doc map[string]any) map[string]any {
				return nestedMap(doc, "components", "schemas", "SchemaDefinition", "properties")
			},
		},
		{
			path: filepath.Join(root, "docs/api-reference/apidocs.swagger.json"),
			props: func(doc map[string]any) map[string]any {
				return nestedMap(doc, "definitions", "SchemaDefinition", "properties")
			},
		},
		{
			path: filepath.Join(root, "docs/api-reference/openapiv2/apidocs.swagger.json"),
			props: func(doc map[string]any) map[string]any {
				return nestedMap(doc, "definitions", "SchemaDefinition", "properties")
			},
		},
	}

	for _, spec := range specs {
		t.Run(spec.path, func(t *testing.T) {
			doc := readJSONObject(t, spec.path)
			props := spec.props(doc)
			if props == nil {
				t.Fatalf("SchemaDefinition properties missing in %s", spec.path)
			}
			for _, name := range []string{"entity_definitions", "rule_definitions"} {
				if _, ok := props[name]; !ok {
					t.Errorf("%s SchemaDefinition missing %q; properties=%v", spec.path, name, jsonKeys(props))
				}
			}
			for _, name := range []string{"entityDefinitions", "ruleDefinitions"} {
				if _, ok := props[name]; ok {
					t.Errorf("%s SchemaDefinition still documents camelCase %q", spec.path, name)
				}
			}
		})
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func readJSONObject(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return doc
}

func nestedMap(doc map[string]any, keys ...string) map[string]any {
	cur := any(doc)
	for _, key := range keys {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = obj[key]
		if !ok {
			return nil
		}
	}
	props, _ := cur.(map[string]any)
	return props
}

func jsonKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
