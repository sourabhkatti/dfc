/*
Copyright 2025 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package dfc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestShellParsing(t *testing.T) {
	type testCase struct {
		name        string
		raw         string
		expected    string
		wantCommand *ShellCommand
	}
	cases := []testCase{}

	for _, delimiter := range []string{"&&", "||", ";", "|"} { // TODO: constant for these?
		cases = append(cases, testCase{
			name:     "basic " + delimiter,
			raw:      `echo hello ` + delimiter + ` echo world`,
			expected: `echo hello ` + delimiter + partSeparator + `echo world`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command:   "echo",
						Args:      []string{"hello"},
						Delimiter: delimiter,
					},
					{
						Command: "echo",
						Args:    []string{"world"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "spacing " + delimiter,
			raw:      `    echo     hello    ` + delimiter + `    echo     world   `,
			expected: `echo hello ` + delimiter + partSeparator + `echo world`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command:   "echo",
						Args:      []string{"hello"},
						Delimiter: delimiter,
					},
					{
						Command: "echo",
						Args:    []string{"world"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "quoted args " + delimiter,
			raw:      `echo "hello notanarg"    other ` + delimiter + ` echo world    'not an arg'    other`,
			expected: `echo "hello notanarg" other ` + delimiter + partSeparator + `echo world 'not an arg' other`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command:   "echo",
						Args:      []string{`"hello notanarg"`, "other"},
						Delimiter: delimiter,
					},
					{
						Command: "echo",
						Args:    []string{"world", `'not an arg'`, "other"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "parentheses section treated as single command  " + delimiter,
			raw:      `(echo "hello" && echo "bye") ` + delimiter + ` echo world`,
			expected: `(echo "hello" && echo "bye") ` + delimiter + partSeparator + `echo world`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command:   `(echo "hello" && echo "bye")`,
						Delimiter: delimiter,
					},
					{
						Command: "echo",
						Args:    []string{"world"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "subshell section treated as single arg  " + delimiter,
			raw:      `echo $(echo "hello") ` + delimiter + ` echo world`,
			expected: `echo $(echo "hello") ` + delimiter + partSeparator + `echo world`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command:   `echo`,
						Args:      []string{`$(echo "hello")`},
						Delimiter: delimiter,
					},
					{
						Command: "echo",
						Args:    []string{"world"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "backtick section treated as single arg  " + delimiter,
			raw:      `echo ` + "`" + `echo "hello"` + "`" + ` ` + delimiter + ` echo world`,
			expected: `echo ` + "`" + `echo "hello"` + "`" + ` ` + delimiter + partSeparator + `echo world`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command:   `echo`,
						Args:      []string{"`" + `echo "hello"` + "`"},
						Delimiter: delimiter,
					},
					{
						Command: "echo",
						Args:    []string{"world"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "delimiter inside quotes ignored " + delimiter,
			raw:      `echo "hello notanarg ` + delimiter + `" other ` + delimiter + ` echo world 'not an arg ` + delimiter + `' other`,
			expected: `echo "hello notanarg ` + delimiter + `" other ` + delimiter + partSeparator + `echo world 'not an arg ` + delimiter + `' other`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command:   "echo",
						Args:      []string{`"hello notanarg ` + delimiter + `"`, "other"},
						Delimiter: delimiter,
					},
					{
						Command: "echo",
						Args:    []string{"world", `'not an arg ` + delimiter + `'`, "other"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "env vars get preserved " + delimiter,
			raw:      `A=1 B="2 ||" C= echo hello ` + delimiter + ` X=3 Y=4 Z="5 &&" echo world`,
			expected: `A=1 B="2 ||" C= echo hello ` + delimiter + partSeparator + `X=3 Y=4 Z="5 &&" echo world`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						ExtraPre:  `A=1 B="2 ||" C=`,
						Command:   `echo`,
						Args:      []string{"hello"},
						Delimiter: delimiter,
					},
					{
						ExtraPre: `X=3 Y=4 Z="5 &&"`,
						Command:  `echo`,
						Args:     []string{"world"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "comments and whitespace get stripped " + delimiter,
			raw:      "# comment before\n" + `echo hello ` + delimiter + ` echo world` + "\n# comment after\n",
			expected: `echo hello ` + delimiter + partSeparator + `echo world`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command:   "echo",
						Args:      []string{"hello"},
						Delimiter: delimiter,
					},
					{
						Command: "echo",
						Args:    []string{"world"},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "incomplete commands parsed correctly, doublequote " + delimiter,
			raw:      `echo "hello world ` + delimiter + ` blah blah blah`,
			expected: `echo "hello world ` + delimiter + ` blah blah blah`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command: "echo",
						Args:    []string{`"hello world ` + delimiter + ` blah blah blah`},
					},
				},
			},
		})

		cases = append(cases, testCase{
			name:     "incomplete commands parsed correctly, singlequote " + delimiter,
			raw:      `echo 'hello world ` + delimiter + ` blah blah blah`,
			expected: `echo 'hello world ` + delimiter + ` blah blah blah`,
			wantCommand: &ShellCommand{
				Parts: []*ShellPart{
					{
						Command: "echo",
						Args:    []string{`'hello world ` + delimiter + ` blah blah blah`},
					},
				},
			},
		})
	}

	cases = append(cases, testCase{
		name: "real world - django  ",
		raw: `apt-get update \
    && apt-get install --assume-yes --no-install-recommends \
        g++ \
        gcc \
        libc6-dev \
        libpq-dev \
        zlib1g-dev \
    && python3 -m pip install --no-cache-dir -r ${REQ_FILE} \
    && apt-get purge --assume-yes --auto-remove \
        g++ \
        gcc \
        libc6-dev \
        libpq-dev \
        zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*`,
		expected: `apt-get update && \
    apt-get install --assume-yes --no-install-recommends g++ gcc libc6-dev libpq-dev zlib1g-dev && \
    python3 -m pip install --no-cache-dir -r ${REQ_FILE} && \
    apt-get purge --assume-yes --auto-remove g++ gcc libc6-dev libpq-dev zlib1g-dev && \
    rm -rf /var/lib/apt/lists/*`,
		wantCommand: &ShellCommand{
			Parts: []*ShellPart{
				{
					Command:   "apt-get",
					Args:      []string{"update"},
					Delimiter: "&&",
				},
				{
					Command:   "apt-get",
					Args:      []string{"install", "--assume-yes", "--no-install-recommends", "g++", "gcc", "libc6-dev", "libpq-dev", "zlib1g-dev"},
					Delimiter: "&&",
				},
				{
					Command:   "python3",
					Args:      []string{"-m", "pip", "install", "--no-cache-dir", "-r", "${REQ_FILE}"},
					Delimiter: "&&",
				},
				{
					Command:   "apt-get",
					Args:      []string{"purge", "--assume-yes", "--auto-remove", "g++", "gcc", "libc6-dev", "libpq-dev", "zlib1g-dev"},
					Delimiter: "&&",
				},
				{
					Command: "rm",
					Args:    []string{"-rf", "/var/lib/apt/lists/*"},
				},
			},
		},
	})

	cases = append(cases, testCase{
		name: "real world - inner comment stripped  ",
		raw: `apt-get update \
# some comment here
# more comments here
&& X=1 Y='2' apt-get install -qy something \
&& apt-get remove -y somethingelse`,
		expected: `apt-get update && \
    X=1 Y='2' apt-get install -qy something && \
    apt-get remove -y somethingelse`,
		wantCommand: &ShellCommand{
			Parts: []*ShellPart{
				{
					Command:   "apt-get",
					Args:      []string{"update"},
					Delimiter: "&&",
				},
				{
					ExtraPre:  `X=1 Y='2'`,
					Command:   "apt-get",
					Args:      []string{"install", "-qy", "something"},
					Delimiter: "&&",
				},
				{
					Command: "apt-get",
					Args:    []string{"remove", "-y", "somethingelse"},
				},
			},
		},
	})

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMultilineShell(tt.raw)
			if got == nil {
				t.Fatalf("%s: got nil shell command", tt.name)
			}

			if diff := cmp.Diff(tt.wantCommand, got); diff != "" {
				t.Errorf("%s: shell parse mismatch (-want, +got):\n%s\n", tt.name, diff)
			}

			// Make sure the command can reconstruct properly
			reconstructed := got.String()
			if diff := cmp.Diff(tt.expected, reconstructed); diff != "" {
				t.Errorf("%s: reconstructing shell (-want, +got):\n%s\n", tt.name, diff)
			}
		})
	}
}
