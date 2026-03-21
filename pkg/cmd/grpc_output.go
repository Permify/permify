package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"google.golang.org/grpc/status"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// GRPCStatusError wraps gRPC status errors with context for callers and logging.
func GRPCStatusError(err error) error {
	if s, ok := status.FromError(err); ok {
		return fmt.Errorf("rpc: %s: %w", s.Message(), err)
	}
	return err
}

func formatCheckResult(w io.Writer, res *basev1.PermissionCheckResponse) {
	switch res.GetCan() {
	case basev1.CheckResult_CHECK_RESULT_ALLOWED:
		_, _ = fmt.Fprintln(w, "Result: allowed")
	case basev1.CheckResult_CHECK_RESULT_DENIED:
		_, _ = fmt.Fprintln(w, "Result: denied")
	default:
		_, _ = fmt.Fprintln(w, "Result: unspecified")
	}
	if m := res.GetMetadata(); m != nil && m.GetCheckCount() > 0 {
		_, _ = fmt.Fprintf(w, "Check count: %d\n", m.GetCheckCount())
	}
}

func formatSchemaSummary(w io.Writer, schema *basev1.SchemaDefinition) {
	if schema == nil {
		_, _ = fmt.Fprintln(w, "(empty schema)")
		return
	}
	var entities []string
	for name := range schema.GetEntityDefinitions() {
		entities = append(entities, name)
	}
	sort.Strings(entities)
	if len(entities) == 0 {
		_, _ = fmt.Fprintln(w, "No entity definitions.")
		return
	}
	_, _ = fmt.Fprintf(w, "Entities (%d):\n", len(entities))
	for _, name := range entities {
		def := schema.GetEntityDefinitions()[name]
		_, _ = fmt.Fprintf(w, "  • %s\n", name)
		if def == nil {
			continue
		}
		var rels []string
		for r := range def.GetRelations() {
			rels = append(rels, r)
		}
		sort.Strings(rels)
		if len(rels) > 0 {
			_, _ = fmt.Fprintf(w, "    relations: %s\n", strings.Join(rels, ", "))
		}
		var perms []string
		for p := range def.GetPermissions() {
			perms = append(perms, p)
		}
		sort.Strings(perms)
		if len(perms) > 0 {
			_, _ = fmt.Fprintf(w, "    permissions: %s\n", strings.Join(perms, ", "))
		}
	}
	var rules []string
	for name := range schema.GetRuleDefinitions() {
		rules = append(rules, name)
	}
	sort.Strings(rules)
	if len(rules) > 0 {
		_, _ = fmt.Fprintf(w, "Rules (%d): %s\n", len(rules), strings.Join(rules, ", "))
	}
}

func formatTupleLine(t *basev1.Tuple) string {
	if t == nil {
		return ""
	}
	subj := t.GetSubject()
	ent := t.GetEntity()
	subjStr := "?"
	if subj != nil {
		subjStr = subj.GetType() + ":" + subj.GetId()
		if r := subj.GetRelation(); r != "" {
			subjStr += "#" + r
		}
	}
	entStr := "?"
	if ent != nil {
		entStr = ent.GetType() + ":" + ent.GetId()
	}
	return fmt.Sprintf("%s#%s@%s", entStr, t.GetRelation(), subjStr)
}
