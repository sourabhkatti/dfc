package dfc

import (
	"testing"
)

func TestConvertGNUTarToBusyboxTar(t *testing.T) {
	testCases := []struct {
		name     string
		input    *ShellPart
		expected *ShellPart
	}{
		{
			name: "basic extract tar with short options",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"xf", "archive.tar"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "-f", "archive.tar"},
			},
		},
		{
			name: "create tar with verbose",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"cvf", "archive.tar", "file1", "file2"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-c", "-v", "file1", "file2", "-f", "archive.tar"},
			},
		},
		{
			name: "extract with gzip",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"xzf", "archive.tar.gz"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "-z", "-f", "archive.tar.gz"},
			},
		},
		{
			name: "extract with bzip2",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"xjf", "archive.tar.bz2"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "-j", "-f", "archive.tar.bz2"},
			},
		},
		{
			name: "extract to specific directory",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"xf", "archive.tar", "-C", "/tmp/extract"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "-C", "/tmp/extract", "-f", "archive.tar"},
			},
		},
		{
			name: "with long options",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"--create", "--verbose", "--file=archive.tar", "file1", "file2"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-c", "-v", "file1", "file2", "-f", "archive.tar"},
			},
		},
		{
			name: "skip unsupported GNU options",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"--extract", "--file", "archive.tar", "--same-owner", "--preserve-permissions", "file1"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "file1", "-f", "archive.tar"},
			},
		},
		{
			name: "preserves extra parts",
			input: &ShellPart{
				ExtraPre:  "# Extract archive",
				Command:   CommandGNUTar,
				Args:      []string{"xf", "archive.tar"},
				Delimiter: "&&",
			},
			expected: &ShellPart{
				ExtraPre:  "# Extract archive",
				Command:   CommandBusyBoxTar,
				Args:      []string{"-x", "-f", "archive.tar"},
				Delimiter: "&&",
			},
		},
		{
			name: "handles complex scenario",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"--extract", "--verbose", "--file", "archive.tar", "--directory", "/tmp", "--same-owner", "dir1", "file1"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "-v", "-C", "/tmp", "dir1", "file1", "-f", "archive.tar"},
			},
		},
		{
			name: "already busybox style - extract with individual options",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"-x", "-v", "-f", "archive.tar"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "-v", "-f", "archive.tar"},
			},
		},
		{
			name: "already busybox style - extract with directory option",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"-x", "-v", "-C", "/tmp", "-f", "archive.tar"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "-v", "-C", "/tmp", "-f", "archive.tar"},
			},
		},
		{
			name: "already busybox style - create with individual options",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"-c", "-v", "file1", "file2", "-f", "archive.tar"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-c", "-v", "file1", "file2", "-f", "archive.tar"},
			},
		},
		{
			name: "already busybox style - extract with gzip",
			input: &ShellPart{
				Command: CommandGNUTar,
				Args:    []string{"-x", "-z", "-f", "archive.tar.gz"},
			},
			expected: &ShellPart{
				Command: CommandBusyBoxTar,
				Args:    []string{"-x", "-z", "-f", "archive.tar.gz"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertGNUTarToBusyboxTar(tc.input)

			// Compare command
			if result.Command != tc.expected.Command {
				t.Errorf("Command: expected %q, got %q", tc.expected.Command, result.Command)
			}

			// Compare args
			if len(result.Args) != len(tc.expected.Args) {
				t.Errorf("Args length: expected %d, got %d\nExpected: %v\nGot: %v",
					len(tc.expected.Args), len(result.Args),
					tc.expected.Args, result.Args)
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
