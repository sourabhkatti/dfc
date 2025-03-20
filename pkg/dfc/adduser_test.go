/*
Copyright 2025 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package dfc

import (
	"testing"
)

func TestConvertUserAddToAddUser(t *testing.T) {
	testCases := []struct {
		name     string
		input    *ShellPart
		expected *ShellPart
	}{
		{
			name: "basic useradd",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"myuser"},
			},
		},
		{
			name: "create home directory",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-m", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"myuser"},
			},
		},
		{
			name: "create home directory long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--create-home", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"myuser"},
			},
		},
		{
			name: "system user",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-r", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--system", "myuser"},
			},
		},
		{
			name: "system user long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--system", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--system", "myuser"},
			},
		},
		{
			name: "custom shell",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-s", "/bin/bash", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--shell", "/bin/bash", "myuser"},
			},
		},
		{
			name: "custom shell long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--shell", "/bin/bash", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--shell", "/bin/bash", "myuser"},
			},
		},
		{
			name: "custom home directory",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-d", "/custom/home", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--home", "/custom/home", "myuser"},
			},
		},
		{
			name: "custom home directory long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--home-dir", "/custom/home", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--home", "/custom/home", "myuser"},
			},
		},
		{
			name: "with comment",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-c", "Test User", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--gecos", "Test User", "myuser"},
			},
		},
		{
			name: "with comment long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--comment", "Test User", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--gecos", "Test User", "myuser"},
			},
		},
		{
			name: "with password",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-p", "password123", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--disabled-password", "myuser"},
			},
		},
		{
			name: "with password long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--password", "password123", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--disabled-password", "myuser"},
			},
		},
		{
			name: "with primary group",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-g", "mygroup", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--ingroup", "mygroup", "myuser"},
			},
		},
		{
			name: "with primary group long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--gid", "mygroup", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--ingroup", "mygroup", "myuser"},
			},
		},
		{
			name: "with user ID",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-u", "1001", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--uid", "1001", "myuser"},
			},
		},
		{
			name: "with user ID long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--uid", "1001", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--uid", "1001", "myuser"},
			},
		},
		{
			name: "no home directory",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-M", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--no-create-home", "myuser"},
			},
		},
		{
			name: "no home directory long option",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"--no-create-home", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--no-create-home", "myuser"},
			},
		},
		{
			name: "multiple options",
			input: &ShellPart{
				Command: CommandUserAdd,
				Args:    []string{"-m", "-s", "/bin/bash", "-u", "1001", "-g", "mygroup", "myuser"},
			},
			expected: &ShellPart{
				Command: CommandAddUser,
				Args:    []string{"--shell", "/bin/bash", "--uid", "1001", "--ingroup", "mygroup", "myuser"},
			},
		},
		{
			name: "preserves extra parts",
			input: &ShellPart{
				ExtraPre:  "# This is a comment",
				Command:   CommandUserAdd,
				Args:      []string{"myuser"},
				Delimiter: "&&",
			},
			expected: &ShellPart{
				ExtraPre:  "# This is a comment",
				Command:   CommandAddUser,
				Args:      []string{"myuser"},
				Delimiter: "&&",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertUserAddToAddUser(tc.input)

			// Compare command
			if result.Command != tc.expected.Command {
				t.Errorf("Command: expected %q, got %q", tc.expected.Command, result.Command)
			}

			// Compare args
			if len(result.Args) != len(tc.expected.Args) {
				t.Errorf("Args length: expected %d, got %d", len(tc.expected.Args), len(result.Args))
			} else {
				for i, arg := range tc.expected.Args {
					if result.Args[i] != arg {
						t.Errorf("Arg[%d]: expected %q, got %q", i, arg, result.Args[i])
					}
				}
			}

			// Compare ExtraPre and Delimiter
			if result.ExtraPre != tc.expected.ExtraPre {
				t.Errorf("ExtraPre: expected %q, got %q", tc.expected.ExtraPre, result.ExtraPre)
			}
			if result.Delimiter != tc.expected.Delimiter {
				t.Errorf("Delimiter: expected %q, got %q", tc.expected.Delimiter, result.Delimiter)
			}
		})
	}
}

func TestConvertGroupAddToAddGroup(t *testing.T) {
	testCases := []struct {
		name     string
		input    *ShellPart
		expected *ShellPart
	}{
		{
			name: "basic groupadd",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"mygroup"},
			},
		},
		{
			name: "system group",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"-r", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"--system", "mygroup"},
			},
		},
		{
			name: "system group long option",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"--system", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"--system", "mygroup"},
			},
		},
		{
			name: "custom GID",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"-g", "1001", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"--gid", "1001", "mygroup"},
			},
		},
		{
			name: "custom GID long option",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"--gid", "1001", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"--gid", "1001", "mygroup"},
			},
		},
		{
			name: "with force option",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"-f", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"mygroup"},
			},
		},
		{
			name: "with force option long",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"--force", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"mygroup"},
			},
		},
		{
			name: "with non-unique option",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"-o", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"mygroup"},
			},
		},
		{
			name: "with non-unique option long",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"--non-unique", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"mygroup"},
			},
		},
		{
			name: "with password option",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"-p", "password123", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"mygroup"},
			},
		},
		{
			name: "with password option long",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"--password", "password123", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"mygroup"},
			},
		},
		{
			name: "multiple options",
			input: &ShellPart{
				Command: CommandGroupAdd,
				Args:    []string{"-r", "-g", "1001", "mygroup"},
			},
			expected: &ShellPart{
				Command: CommandAddGroup,
				Args:    []string{"--system", "--gid", "1001", "mygroup"},
			},
		},
		{
			name: "preserves extra parts",
			input: &ShellPart{
				ExtraPre:  "# This is a comment",
				Command:   CommandGroupAdd,
				Args:      []string{"mygroup"},
				Delimiter: "&&",
			},
			expected: &ShellPart{
				ExtraPre:  "# This is a comment",
				Command:   CommandAddGroup,
				Args:      []string{"mygroup"},
				Delimiter: "&&",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertGroupAddToAddGroup(tc.input)

			// Compare command
			if result.Command != tc.expected.Command {
				t.Errorf("Command: expected %q, got %q", tc.expected.Command, result.Command)
			}

			// Compare args
			if len(result.Args) != len(tc.expected.Args) {
				t.Errorf("Args length: expected %d, got %d", len(tc.expected.Args), len(result.Args))
			} else {
				for i, arg := range tc.expected.Args {
					if result.Args[i] != arg {
						t.Errorf("Arg[%d]: expected %q, got %q", i, arg, result.Args[i])
					}
				}
			}

			// Compare ExtraPre and Delimiter
			if result.ExtraPre != tc.expected.ExtraPre {
				t.Errorf("ExtraPre: expected %q, got %q", tc.expected.ExtraPre, result.ExtraPre)
			}
			if result.Delimiter != tc.expected.Delimiter {
				t.Errorf("Delimiter: expected %q, got %q", tc.expected.Delimiter, result.Delimiter)
			}
		})
	}
}
