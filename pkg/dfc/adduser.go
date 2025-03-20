package dfc

import (
	"strings"
)

// ConvertUserAddToAddUser converts a useradd command to the equivalent adduser command
func ConvertUserAddToAddUser(part *ShellPart) *ShellPart {
	if part.Command != CommandUserAdd {
		return part
	}

	// Create a new shell part with the same extra content and delimiter
	result := &ShellPart{
		ExtraPre:  part.ExtraPre,
		Command:   CommandAddUser,
		Delimiter: part.Delimiter,
	}

	// Process arguments
	var resultArgs []string
	var username string
	var hasUsername bool
	i := 0

	// Process arguments
	for i < len(part.Args) {
		arg := part.Args[i]

		// Look for the username (first non-option argument)
		if !strings.HasPrefix(arg, "-") && !hasUsername {
			username = arg
			hasUsername = true
			i++
			continue
		}

		// Process options
		switch arg {
		// Options that are simply removed (create home is default in adduser)
		case "-m", "--create-home":
			i++
			continue

		// Options that are renamed
		case "-r", "--system":
			resultArgs = append(resultArgs, "--system")
			i++

		case "-M", "--no-create-home":
			resultArgs = append(resultArgs, "--no-create-home")
			i++

		// Options that need arguments and are renamed
		case "-s", "--shell":
			if i+1 < len(part.Args) {
				resultArgs = append(resultArgs, "--shell", part.Args[i+1])
				i += 2
			} else {
				i++
			}

		case "-d", "--home-dir":
			if i+1 < len(part.Args) {
				resultArgs = append(resultArgs, "--home", part.Args[i+1])
				i += 2
			} else {
				i++
			}

		case "-c", "--comment":
			if i+1 < len(part.Args) {
				resultArgs = append(resultArgs, "--gecos", part.Args[i+1])
				i += 2
			} else {
				i++
			}

		case "-g", "--gid":
			if i+1 < len(part.Args) {
				resultArgs = append(resultArgs, "--ingroup", part.Args[i+1])
				i += 2
			} else {
				i++
			}

		case "-u", "--uid":
			if i+1 < len(part.Args) {
				resultArgs = append(resultArgs, "--uid", part.Args[i+1])
				i += 2
			} else {
				i++
			}

		// Password options converted to --disabled-password
		case "-p", "--password":
			resultArgs = append(resultArgs, "--disabled-password")
			if i+1 < len(part.Args) && !strings.HasPrefix(part.Args[i+1], "-") {
				i += 2
			} else {
				i++
			}

		// Options that we skip along with their arguments
		case "-k", "--skel", "-N", "--no-user-group":
			if i+1 < len(part.Args) && !strings.HasPrefix(part.Args[i+1], "-") {
				i += 2
			} else {
				i++
			}

		// Include other parts that haven't been processed
		default:
			resultArgs = append(resultArgs, arg)
			i++
		}
	}

	// Add username at the end
	if hasUsername {
		resultArgs = append(resultArgs, username)
	}

	result.Args = resultArgs
	return result
}

// ConvertGroupAddToAddGroup converts a groupadd command to the equivalent addgroup command
func ConvertGroupAddToAddGroup(part *ShellPart) *ShellPart {
	if part.Command != CommandGroupAdd {
		return part
	}

	// Create a new shell part with the same extra content and delimiter
	result := &ShellPart{
		ExtraPre:  part.ExtraPre,
		Command:   CommandAddGroup,
		Delimiter: part.Delimiter,
	}

	// Process arguments
	var resultArgs []string
	var groupname string
	var hasGroupname bool
	i := 0

	// Process arguments
	for i < len(part.Args) {
		arg := part.Args[i]

		// Look for the groupname (first non-option argument)
		if !strings.HasPrefix(arg, "-") && !hasGroupname {
			groupname = arg
			hasGroupname = true
			i++
			continue
		}

		// Process options
		switch arg {
		// Options that are renamed
		case "-r", "--system":
			resultArgs = append(resultArgs, "--system")
			i++

		// Options that need arguments and are renamed
		case "-g", "--gid":
			if i+1 < len(part.Args) {
				resultArgs = append(resultArgs, "--gid", part.Args[i+1])
				i += 2
			} else {
				i++
			}

		// Options that we skip (not supported in addgroup)
		case "-f", "--force", "-o", "--non-unique":
			i++
			continue

		// Options that we skip along with their arguments
		case "-K", "--key", "-p", "--password":
			if i+1 < len(part.Args) && !strings.HasPrefix(part.Args[i+1], "-") {
				i += 2
			} else {
				i++
			}

		// Include other parts that haven't been processed
		default:
			resultArgs = append(resultArgs, arg)
			i++
		}
	}

	// Add groupname at the end
	if hasGroupname {
		resultArgs = append(resultArgs, groupname)
	}

	result.Args = resultArgs
	return result
}
