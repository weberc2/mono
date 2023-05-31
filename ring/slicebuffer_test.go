package ring

import (
	"encoding/json"
	"testing"
)

func TestSliceBuffer(t *testing.T) {
	for _, tc := range []struct {
		name        string
		state       *SliceBuffer[string]
		pushItem    string
		evictedItem string
		evicted     bool
		wantedItems []string
	}{
		{
			name:        "empty-cap-one",
			state:       must(NewSliceBuffer[string](1)),
			pushItem:    "hello",
			evicted:     false,
			wantedItems: []string{"hello"},
		},
		{
			name: "full",
			state: &SliceBuffer[string]{
				entries: []string{"hello", ""},
				start:   0,
				tail:    1,
			},
			pushItem:    "world",
			evicted:     true,
			evictedItem: "hello",
			wantedItems: []string{"world"},
		},
		{
			name: "inverted-spare-capacity",
			state: &SliceBuffer[string]{
				entries: []string{"", "0", "1", ""},
				tail:    3,
				start:   1,
			},
			pushItem:    "2",
			evicted:     false,
			wantedItems: []string{"0", "1", "2"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item, evicted := tc.state.Push(tc.pushItem)
			if tc.evicted {
				if !evicted {
					t.Fatal("expected eviction, but no value was evicted")
				}

				if tc.evictedItem != item {
					t.Fatalf(
						"wanted evicted item `%s`; found `%s`",
						tc.evictedItem,
						item,
					)
				}
			} else if evicted {
				t.Fatalf("expected no eviction, but `%s` was evicted", item)
			}

			items := tc.state.Items()
			if len(items) == len(tc.wantedItems) {
				for i := range items {
					if items[i] != tc.wantedItems[i] {
						goto MISMATCH
					}
				}
				return
			}
		MISMATCH:
			wanted, err := json.Marshal(tc.wantedItems)
			if err != nil {
				t.Fatalf("unexpected err: marshaling wanted items: %v", err)
			}
			found, err := json.Marshal(items)
			if err != nil {
				t.Fatalf("unexpected err: marshaling found items: %v", err)
			}
			t.Fatalf(
				"mismatched items:\n  wanted: %s\n  found: %s",
				wanted,
				found,
			)
		})
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
