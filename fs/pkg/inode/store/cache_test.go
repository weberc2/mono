package store

import (
	"encoding/json"
	"testing"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func TestCache_GetWhenEmpty(t *testing.T) {
	c := NewCache(2)

	var inode Inode
	if c.Get(0, &inode) {
		t.Fatal("empty cache: getting ino `0`: expected `false`; found `true`")
	}
}

func TestCache(t *testing.T) {
	type testCase struct {
		name          string
		capacity      int
		initialState  []Inode
		pushInput     Inode
		wantedEvicted *Inode
		getInput      Ino
		wanted        *Inode
	}

	testCases := []testCase{{
		name:          "empty",
		capacity:      2,
		pushInput:     Inode{Ino: 10},
		wantedEvicted: nil,
		getInput:      10,
		wanted:        &Inode{Ino: 10},
	}, {
		name:          "neither empty nor full",
		capacity:      2,
		initialState:  []Inode{{Ino: 9}},
		pushInput:     Inode{Ino: 10},
		wantedEvicted: nil,
		getInput:      10,
		wanted:        &Inode{Ino: 10},
	}, {
		name:          "eviction",
		capacity:      1,
		initialState:  []Inode{{Ino: 9}},
		pushInput:     Inode{Ino: 10},
		wantedEvicted: &Inode{Ino: 9},
		getInput:      9,
		wanted:        nil,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// initialize the test cache
			c := NewCache(tc.capacity)
			for i := range tc.initialState {
				var evicted Inode
				if c.Push(&tc.initialState[i], &evicted) {
					data, err := json.Marshal(&evicted)
					if err != nil {
						t.Fatalf("failed to marshal inode to json: %v", err)
					}
					t.Fatalf(
						"initializing test cache: unexpected eviction: %s",
						data,
					)
				}
			}

			var output, evicted Inode
			if c.Push(&tc.pushInput, &evicted) {
				// If there was an eviction but no eviction was expected, fail.
				if tc.wantedEvicted == nil {
					data, err := json.Marshal(&evicted)
					if err != nil {
						t.Fatalf("failed to marshal inode to json: %v", err)
					}
					t.Fatalf("unexpected eviction: %s", data)
				}

				// If there was an eviction and it was expected, make sure the
				// evicted inode was as expected.
				if *tc.wantedEvicted != evicted {
					wanted, err := json.Marshal(tc.wantedEvicted)
					if err != nil {
						t.Fatalf("failed to marshal inode to json: %v", err)
					}
					found, err := json.Marshal(&evicted)
					if err != nil {
						t.Fatalf("failed to marshal inode to json: %v", err)
					}

					t.Fatalf("evicted: wanted `%s`; found `%s`", wanted, found)
				}

				// If the eviction was expected and the evicted inode matched
				// expectation, then return successfully.
				return
			}

			// If there was no eviction but an eviction was expected, fail.
			if tc.wantedEvicted != nil {
				t.Fatal("expected eviction but no eviction reported")
			}

			// If there was no eviction and no eviction was expected, make sure
			// the evicted value wasn't mutated.
			if evicted != (Inode{}) {
				data, err := json.Marshal(&evicted)
				if err != nil {
					t.Fatalf("failed to marshal inode to json: %v", err)
				}
				t.Fatalf(
					"no eviction was reported but the evicted inode value "+
						"was modified: %s",
					data,
				)
			}

			// If there was no eviction nor expectation of an eviction and the
			// evicted inode value wasn't modified, then proceed to test
			// `Get()`.
			found := c.Get(tc.getInput, &output)
			if found {
				// If there was a match reported but no match was expected,
				// fail.
				if tc.wanted == nil {
					data, err := json.Marshal(&output)
					if err != nil {
						t.Fatalf("failed to marshal inode as json: %v", err)
					}

					t.Fatalf("Get(): unexpected match reported: %s", data)
				}

				// If there was a match reported *and* expected, verify that
				// the value meets expectations. If not, fail.
				if *tc.wanted != output {
					wanted, err := json.Marshal(tc.wanted)
					if err != nil {
						t.Fatalf("failed to marshal inode to json: %v", err)
					}
					found, err := json.Marshal(&output)
					if err != nil {
						t.Fatalf("failed to marshal inode to json: %v", err)
					}
					t.Fatalf("Get(): wanted `%s`; found `%s`", wanted, found)
				}

				// If there was a match reported and a match expected and the
				// matched value meets expectations, then return success.
				return
			}

			// If no match was found, but a match was expected, fail.
			if tc.wanted != nil {
				t.Fatal("Get(): expected match but no match was found")
			}

			// If no match was found nor expected, make sure that the output
			// inode value wasn't modified.
			if output != (Inode{}) {
				data, err := json.Marshal(&output)
				if err != nil {
					t.Fatalf("failed to marshal inode to json: %v", err)
				}
				t.Fatalf(
					"Get(): unexpected mutation to output inode value: %s",
					data,
				)
			}

			// If no match was found nor expected and the output value wasn't
			// tampered with, then return success
			return
		})
	}
}
