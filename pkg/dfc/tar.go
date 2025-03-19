package dfc

import (
	"strings"
)

// Command constants for tar commands
const (
	CommandGNUTar     = "tar"
	CommandBusyBoxTar = "tar"
)

// ConvertGNUTarToBusyboxTar converts a GNU tar command to the equivalent BusyBox tar command
// BusyBox tar has fewer options and some different syntax compared to GNU tar
func ConvertGNUTarToBusyboxTar(part *ShellPart) *ShellPart {
	if part.Command != CommandGNUTar {
		return part
	}

	// Create a new shell part with the same extra content and delimiter
	result := &ShellPart{
		ExtraPre:  part.ExtraPre,
		Command:   CommandBusyBoxTar,
		Delimiter: part.Delimiter,
	}

	// We'll use separate slices for options, files, and file option
	var options []string
	var files []string
	var hasFile bool
	var filename string

	i := 0

	// First pass to check for common options and gather information
	for i < len(part.Args) {
		arg := part.Args[i]

		// Handle the main operation flags (first argument usually)
		if i == 0 && !strings.HasPrefix(arg, "-") && len(arg) > 0 {
			// Convert combined options like "xvf" to individual options
			for _, c := range arg {
				switch c {
				case 'x':
					options = append(options, "-x")
				case 'c':
					options = append(options, "-c")
				case 'v':
					options = append(options, "-v")
				case 'f':
					hasFile = true
					// Skip adding -f here, we'll add it with the filename
					if i+1 < len(part.Args) {
						filename = part.Args[i+1]
						i++
					}
				case 'z':
					options = append(options, "-z")
				case 'j':
					options = append(options, "-j")
				default:
					// Pass through other single-letter options
					options = append(options, "-"+string(c))
				}
			}
			i++
			continue
		}

		// Handle --file=value format
		if strings.HasPrefix(arg, "--file=") {
			hasFile = true
			filename = arg[7:] // Extract part after --file=
			i++
			continue
		}

		// Handle long options and their short equivalents
		switch arg {
		// Extract operations
		case "--extract", "-x":
			options = append(options, "-x")

		// Create operations
		case "--create", "-c":
			options = append(options, "-c")

		// Verbose output
		case "--verbose", "-v":
			options = append(options, "-v")

		// File specification
		case "--file", "-f":
			hasFile = true
			if i+1 < len(part.Args) {
				filename = part.Args[i+1]
				i += 2
				continue
			}

		// Compress with gzip
		case "--gzip", "--gunzip", "-z":
			options = append(options, "-z")

		// Compress with bzip2
		case "--bzip2", "-j":
			options = append(options, "-j")

		// Change directory
		case "--directory", "-C":
			if i+1 < len(part.Args) {
				options = append(options, "-C", part.Args[i+1])
				i += 2
				continue
			}

		// Handle unsupported or ignored GNU tar options
		case "--same-owner", "--preserve-permissions", "--preserve-order",
			"--preserve", "--same-permissions", "--numeric-owner",
			"--overwrite", "--remove-files", "--ignore-failed-read":
			// These options are either default or not needed in BusyBox tar
			i++
			continue

		default:
			// Check if it's a long option we need to skip
			if strings.HasPrefix(arg, "--") {
				// Skip unknown long options
				i++
				continue
			}

			// If it doesn't start with -, it's probably a file or directory
			files = append(files, arg)
		}
		i++
	}

	// Build the final args in the correct order
	var resultArgs []string

	// First add all the options
	resultArgs = append(resultArgs, options...)

	// Then add the files list
	resultArgs = append(resultArgs, files...)

	// Finally add the file option at the end if present
	if hasFile && filename != "" {
		resultArgs = append(resultArgs, "-f", filename)
	}

	result.Args = resultArgs
	return result
}
