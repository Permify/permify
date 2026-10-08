package engines

import (
	"context"
	"slices"
	"strings"
	"testing"
)

// normalize sorts a filter result, including the ids packed into a wildcard, so a
// comparison does not depend on the order the concurrent filter functions finish
// in.
func normalize(ids []string) []string {
	normalized := make([]string, 0, len(ids))
	for _, id := range ids {
		excluded, packed := strings.CutPrefix(id, ALL+"-")
		if !packed {
			normalized = append(normalized, id)
			continue
		}
		carved := strings.Split(excluded, ",")
		slices.Sort(carved)
		normalized = append(normalized, ALL+"-"+strings.Join(carved, ","))
	}
	slices.Sort(normalized)
	return normalized
}

// A filter result that carries "<>" alongside other ids means "everything except
// those ids". subjectFilterUnion and subjectFilterExclusion emit that same set
// packed into one string, "<>-1,2,3", so both spellings reach an enclosing filter
// and both have to be read the same way. Each case below runs the packed spelling
// and the unpacked one through the same filter and requires the same answer.
func TestSubjectFilterReadsBothWildcardSpellings(t *testing.T) {
	constant := func(ids ...string) SubjectFilterFunction {
		return func(context.Context) ([]string, error) { return ids, nil }
	}

	tests := []struct {
		name     string
		filter   func(context.Context, []SubjectFilterFunction, int) ([]string, error)
		packed   []SubjectFilterFunction
		unpacked []SubjectFilterFunction
		want     []string
	}{
		{
			// "not everything except 1" leaves only 1, and 1 is in the left set.
			name:     "exclusion, wildcard on the right",
			filter:   subjectFilterExclusion,
			packed:   []SubjectFilterFunction{constant("1", "2", "3"), constant(ALL + "-1")},
			unpacked: []SubjectFilterFunction{constant("1", "2", "3"), constant(ALL, "1")},
			want:     []string{},
		},
		{
			name:     "exclusion, wildcard on the left",
			filter:   subjectFilterExclusion,
			packed:   []SubjectFilterFunction{constant(ALL + "-1"), constant("2")},
			unpacked: []SubjectFilterFunction{constant(ALL, "1"), constant("2")},
			want:     []string{ALL + "-2,1"},
		},
		{
			name:     "intersection with a concrete set",
			filter:   subjectFilterIntersection,
			packed:   []SubjectFilterFunction{constant(ALL + "-1"), constant("1", "2")},
			unpacked: []SubjectFilterFunction{constant(ALL, "1"), constant("1", "2")},
			want:     []string{"2"},
		},
		{
			name:     "intersection of two wildcards",
			filter:   subjectFilterIntersection,
			packed:   []SubjectFilterFunction{constant(ALL + "-1"), constant(ALL + "-2")},
			unpacked: []SubjectFilterFunction{constant(ALL, "1"), constant(ALL, "2")},
			want:     []string{ALL + "-1,2"},
		},
		{
			name:     "union with a member of the wildcard set",
			filter:   subjectFilterUnion,
			packed:   []SubjectFilterFunction{constant(ALL + "-1"), constant("5")},
			unpacked: []SubjectFilterFunction{constant(ALL, "1"), constant("5")},
			want:     []string{ALL + "-1"},
		},
	}

	for _, test := range tests {
		for spelling, functions := range map[string][]SubjectFilterFunction{
			"packed":   test.packed,
			"unpacked": test.unpacked,
		} {
			t.Run(test.name+"/"+spelling, func(t *testing.T) {
				got, err := test.filter(context.Background(), functions, len(functions))
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(normalize(got), normalize(test.want)) {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			})
		}
	}
}
