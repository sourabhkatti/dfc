package dfc

import (
	"context"
	"slices"
	"strings"
)

// Distro represents a Linux distribution
type Distro string

// Manager represents a package manager
type Manager string

// Supported distributions
const (
	DistroDebian Distro = "debian"
	DistroFedora Distro = "fedora"
	DistroAlpine Distro = "alpine"
)

// Supported package managers
const (
	ManagerAptGet   Manager = "apt-get"
	ManagerApk      Manager = "apk"
	ManagerYum      Manager = "yum"
	ManagerDnf      Manager = "dnf"
	ManagerMicrodnf Manager = "microdnf"
	ManagerApt      Manager = "apt"
)

// Install subcommands
const (
	SubcommandInstall = "install"
	SubcommandAdd     = "add"
)

// Dockerfile directives
const (
	DirectiveFrom = "FROM"
	DirectiveRun  = "RUN"
	DirectiveUser = "USER"
	KeywordAs     = "AS"
)

// Default values
const (
	DefaultRegistryDomain = "cgr.dev"
	DefaultImageTag       = "latest-dev"
	DefaultUser           = "root"
	DefaultOrg            = "ORG"
)

// PackageManagerInfo holds metadata about a package manager
type PackageManagerInfo struct {
	Distro         Distro
	InstallKeyword string
}

// PackageManagerInfoMap maps package managers to their metadata
var PackageManagerInfoMap = map[Manager]PackageManagerInfo{
	ManagerAptGet: {Distro: DistroDebian, InstallKeyword: SubcommandInstall},
	ManagerApt:    {Distro: DistroDebian, InstallKeyword: SubcommandInstall},

	ManagerYum:      {Distro: DistroFedora, InstallKeyword: SubcommandInstall},
	ManagerDnf:      {Distro: DistroFedora, InstallKeyword: SubcommandInstall},
	ManagerMicrodnf: {Distro: DistroFedora, InstallKeyword: SubcommandInstall},

	ManagerApk: {Distro: DistroAlpine, InstallKeyword: SubcommandAdd},
}

// DockerfileLine represents a single line in a Dockerfile
type DockerfileLine struct {
	Raw       string       `json:"raw"`
	Converted string       `json:"converted,omitempty"`
	Extra     string       `json:"extra,omitempty"` // Comments and whitespace that appear before this line
	Stage     int          `json:"stage,omitempty"`
	From      *FromDetails `json:"from,omitempty"`
	Run       *RunDetails  `json:"run,omitempty"`
}

// FromDetails holds details about a FROM directive
type FromDetails struct {
	Base        string `json:"base,omitempty"`
	Tag         string `json:"tag,omitempty"`
	Digest      string `json:"digest,omitempty"`
	Alias       string `json:"alias,omitempty"`
	Parent      int    `json:"parent,omitempty"`
	BaseDynamic bool   `json:"baseDynamic,omitempty"`
	TagDynamic  bool   `json:"tagDynamic,omitempty"`
}

// RunDetails holds details about a RUN directive
type RunDetails struct {
	Distro   Distro           `json:"distro,omitempty"`
	Manager  Manager          `json:"manager,omitempty"`
	Packages []string         `json:"packages,omitempty"`
	Shell    *RunDetailsShell `json:"-"`
}

type RunDetailsShell struct {
	Before *ShellCommand
	After  *ShellCommand
}

// Dockerfile represents a parsed Dockerfile
type Dockerfile struct {
	Lines []*DockerfileLine `json:"lines"`
}

// String returns the Dockerfile content as a string
func (d *Dockerfile) String() string {
	var builder strings.Builder

	for i, line := range d.Lines {
		// Add the Extra content (comments, whitespace)
		if line.Extra != "" {
			builder.WriteString(line.Extra)
		}

		// If the line has been converted, use the converted content
		if line.Converted != "" {
			builder.WriteString(line.Converted)
			builder.WriteString("\n")
		} else if line.Raw != "" {
			// If this is a normal content line
			builder.WriteString(line.Raw)

			// If this is the last line, don't add a newline
			if i < len(d.Lines)-1 {
				builder.WriteString("\n")
			}
		}
	}

	return builder.String()
}

// ParseDockerfile parses a Dockerfile into a structured representation
func ParseDockerfile(ctx context.Context, content []byte) (*Dockerfile, error) {
	// Create a new Dockerfile
	dockerfile := &Dockerfile{
		Lines: []*DockerfileLine{},
	}

	// Split into lines while preserving original structure
	lines := strings.Split(string(content), "\n")

	var extraContent strings.Builder
	var currentInstruction strings.Builder
	var inMultilineInstruction bool
	currentStage := 0
	stageAliases := make(map[string]int) // Maps stage aliases to their index

	processCurrentInstruction := func() {
		if currentInstruction.Len() == 0 {
			return
		}

		instruction := currentInstruction.String()
		trimmedInstruction := strings.TrimSpace(instruction)
		upperInstruction := strings.ToUpper(trimmedInstruction)

		// Create a new Dockerfile line
		dockerfileLine := &DockerfileLine{
			Raw:   instruction,
			Extra: extraContent.String(),
			Stage: currentStage,
		}

		// Handle FROM instructions (case-insensitive)
		if strings.HasPrefix(upperInstruction, DirectiveFrom+" ") {
			currentStage++
			dockerfileLine.Stage = currentStage

			// Extract the FROM details
			fromPartIdx := len(DirectiveFrom + " ")
			fromPart := strings.TrimSpace(trimmedInstruction[fromPartIdx:])

			// Check for AS clause which defines an alias (case-insensitive)
			var alias string
			// Capture space + AS + space to get exact length
			asKeywordWithSpaces := " " + KeywordAs + " "

			// Split by case-insensitive " AS " pattern
			asParts := strings.Split(strings.ToUpper(fromPart), asKeywordWithSpaces)
			if len(asParts) > 1 {
				// Find the position of the case-insensitive " AS " to preserve case in the base part
				asIndex := strings.Index(strings.ToUpper(fromPart), asKeywordWithSpaces)
				if asIndex != -1 {
					// Use the original case for the base and alias
					basePart := strings.TrimSpace(fromPart[:asIndex])
					aliasPart := strings.TrimSpace(fromPart[asIndex+len(asKeywordWithSpaces):])
					fromPart = basePart
					alias = aliasPart

					// Store this alias for parent references
					stageAliases[strings.ToLower(alias)] = currentStage
				}
			}

			// Parse the image reference
			var base, tag, digest string

			// Check for digest
			if digestParts := strings.Split(fromPart, "@"); len(digestParts) > 1 {
				fromPart = digestParts[0]
				digest = digestParts[1]
			}

			// Check for tag
			if tagParts := strings.Split(fromPart, ":"); len(tagParts) > 1 {
				base = tagParts[0]
				tag = tagParts[1]
			} else {
				base = fromPart
			}

			// Check for parent reference (case-insensitive)
			var parent int
			if parentStage, exists := stageAliases[strings.ToLower(base)]; exists {
				parent = parentStage
			}

			// Create the FromDetails
			dockerfileLine.From = &FromDetails{
				Base:        base,
				Tag:         tag,
				Digest:      digest,
				Alias:       alias,
				Parent:      parent,
				BaseDynamic: strings.Contains(base, "$"),
				TagDynamic:  strings.Contains(tag, "$"),
			}
		}

		// Handle RUN instructions (case-insensitive)
		if strings.HasPrefix(upperInstruction, DirectiveRun+" ") {
			// Extract the command part (everything after "RUN ")
			cmdPartIdx := len(DirectiveRun + " ")
			cmdPart := strings.TrimSpace(trimmedInstruction[cmdPartIdx:])

			// Parse the shell command
			shellCmd := ParseMultilineShell(cmdPart)

			// Store the shell command in Run.Shell.Before
			if shellCmd != nil {
				dockerfileLine.Run = &RunDetails{
					Shell: &RunDetailsShell{
						Before: shellCmd,
					},
				}
			}
		}

		// Add the line to the Dockerfile
		dockerfile.Lines = append(dockerfile.Lines, dockerfileLine)

		// Reset
		currentInstruction.Reset()
		extraContent.Reset()
	}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Handle empty lines
		if trimmedLine == "" {
			if !inMultilineInstruction {
				extraContent.WriteString(line)
				extraContent.WriteString("\n")
			}
			continue
		}

		// Handle comments
		if strings.HasPrefix(trimmedLine, "#") {
			if !inMultilineInstruction {
				extraContent.WriteString(line)
				extraContent.WriteString("\n")
			}
			continue
		}

		// Check if this is the start of a new instruction or continuation
		if !inMultilineInstruction {
			// Check for continuation character
			if strings.HasSuffix(trimmedLine, "\\") {
				inMultilineInstruction = true
				currentInstruction.WriteString(line)
				currentInstruction.WriteString("\n")
			} else {
				// Single line instruction
				currentInstruction.WriteString(line)
				processCurrentInstruction()
			}
		} else {
			// Continuation of a multi-line instruction
			currentInstruction.WriteString(line)

			// Check if this is the end of the multi-line instruction
			if !strings.HasSuffix(trimmedLine, "\\") {
				inMultilineInstruction = false

				// We don't need to add a newline at the end of a completed multiline instruction
				// This prevents the extra newline that appears at the end of RUN commands
				// Only add newlines between individual lines, not at the end

				processCurrentInstruction()
			} else {
				// Not the end yet, add a newline
				currentInstruction.WriteString("\n")
			}
		}
	}

	// Process any remaining instruction
	if inMultilineInstruction {
		processCurrentInstruction()
	}

	// Capture any trailing whitespace or comments after the last directive
	if extraContent.Len() > 0 {
		// Remove trailing newline if present to avoid double newlines when generating output
		trailingContent := strings.TrimSuffix(extraContent.String(), "\n")
		dockerfile.Lines = append(dockerfile.Lines, &DockerfileLine{
			Raw: trailingContent,
		})
		extraContent.Reset()
	}

	return dockerfile, nil
}

// Options represents conversion options
type Options struct {
	Organization string
	Registry     string
	PackageMap   PackageMap
}

// PackagesConfig represents the structure of packages.yaml
type PackageMap map[Distro]map[string][]string

// Convert applies the conversion to the Dockerfile and returns a new converted Dockerfile
func (d *Dockerfile) Convert(ctx context.Context, opts Options) (*Dockerfile, error) {
	// Create a new Dockerfile for the converted content
	converted := &Dockerfile{
		Lines: make([]*DockerfileLine, len(d.Lines)),
	}

	// Convert each line
	for i, line := range d.Lines {
		// Create a deep copy of the line
		newLine := &DockerfileLine{
			Raw:   line.Raw,
			Extra: line.Extra,
			Stage: line.Stage,
		}

		if line.From != nil {
			newLine.From = &FromDetails{
				Base:        line.From.Base,
				Tag:         line.From.Tag,
				Digest:      line.From.Digest,
				Alias:       line.From.Alias,
				Parent:      line.From.Parent,
				BaseDynamic: line.From.BaseDynamic,
				TagDynamic:  line.From.TagDynamic,
			}

			// Apply FROM line conversion
			if shouldConvertFromLine(line.From) {
				newBase := convertBaseImage(line.From.Base, opts)

				// Handle ubuntu images specially - use "latest" tag
				var newTag string
				if isUbuntuImage(line.From.Base) {
					newTag = "latest"
				} else {
					newTag = convertImageTag(line.From.Tag, line.From.TagDynamic)
				}

				// Create the converted FROM line
				fromLine := DirectiveFrom + " " + newBase
				if newTag != "" {
					fromLine += ":" + newTag
				}
				if line.From.Digest != "" && !isUbuntuImage(line.From.Base) {
					fromLine += "@" + line.From.Digest
				}
				if line.From.Alias != "" {
					fromLine += " " + KeywordAs + " " + line.From.Alias
				}

				newLine.Converted = fromLine
			}
		}

		// Process RUN commands
		if line.Run != nil && line.Run.Shell != nil && line.Run.Shell.Before != nil {
			beforeShell := line.Run.Shell.Before
			modifiedShell := beforeShell

			// Initialize RunDetails with Before shell
			newLine.Run = &RunDetails{
				Shell: &RunDetailsShell{
					Before: beforeShell,
				},
			}

			// Second pass: modify commands
			modified, distro, manager, packages, convertedShell, err := convertPackageManagerCommands(modifiedShell, opts.PackageMap)
			if err != nil {
				return nil, err
			}
			newLine.Run.Distro = distro
			newLine.Run.Manager = manager
			newLine.Run.Packages = packages

			// If we modified the shell command, set After and Converted
			if modified {
				newLine.Run.Shell.After = convertedShell

				// Extract the original RUN directive from the raw line to preserve case
				rawLine := line.Raw
				upperRawLine := strings.ToUpper(rawLine)

				// Find the position of the case-insensitive "RUN " directive
				runPrefix := DirectiveRun + " "
				runIndex := strings.Index(upperRawLine, runPrefix)

				if runIndex != -1 {
					// Get the original case of the RUN directive
					originalRunDirective := rawLine[runIndex : runIndex+len(runPrefix)]
					newLine.Converted = originalRunDirective + convertedShell.String()
				} else {
					// Fallback if we can't find the directive (shouldn't happen)
					newLine.Converted = DirectiveRun + " " + convertedShell.String()
				}
			}
		}

		// Add the converted line to the result
		converted.Lines[i] = newLine
	}

	// Second pass: add USER root directives where needed
	// First determine which stages have converted RUN lines
	stagesWithConvertedRuns := make(map[int]bool)
	// Also keep track of stages that already have USER root directives
	stagesWithUserRoot := make(map[int]bool)

	// First pass - identify stages with converted RUN lines and existing USER root directives
	for _, line := range converted.Lines {
		// Check if this is a converted RUN line
		if line.Run != nil && line.Converted != "" {
			stagesWithConvertedRuns[line.Stage] = true
		}

		// Check if this line is a USER directive with root
		raw := line.Raw
		converted := line.Converted

		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(raw)), DirectiveUser+" ") &&
			strings.Contains(strings.ToLower(raw), DefaultUser) {
			stagesWithUserRoot[line.Stage] = true
		}

		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(converted)), DirectiveUser+" ") &&
			strings.Contains(strings.ToLower(converted), DefaultUser) {
			stagesWithUserRoot[line.Stage] = true
		}
	}

	// If we found any stages with converted RUN lines, add USER root after the FROM
	if len(stagesWithConvertedRuns) > 0 {
		for _, line := range converted.Lines {
			// Check if this is a FROM line in a stage that has converted RUN lines
			if line.From != nil && stagesWithConvertedRuns[line.Stage] {
				// If the FROM line was converted and there's no USER root directive in this stage already
				if line.Converted != "" && !stagesWithUserRoot[line.Stage] {
					// Add a USER root directive after this FROM line
					line.Converted += "\n" + DirectiveUser + " " + DefaultUser
					// Mark this stage as having a USER root directive
					stagesWithUserRoot[line.Stage] = true
				}
			}
		}
	}

	return converted, nil
}

// shouldConvertFromLine determines if a FROM line should be converted
func shouldConvertFromLine(from *FromDetails) bool {
	// Skip conversion for scratch, parent stages, or dynamic bases
	if from.Base == "scratch" || from.Parent > 0 || from.BaseDynamic {
		return false
	}
	return true
}

// convertBaseImage returns the converted base image
func convertBaseImage(base string, opts Options) string {
	// Extract the basename (part after the last slash)
	var basename string
	if lastSlash := strings.LastIndex(base, "/"); lastSlash != -1 {
		basename = base[lastSlash+1:]
	} else {
		basename = base
	}

	// Special case for ubuntu - use chainguard-base instead
	if basename == "ubuntu" {
		basename = "chainguard-base"
	}

	// If registry is specified, use registry/basename
	if opts.Registry != "" {
		return opts.Registry + "/" + basename
	}

	// Otherwise use DefaultRegistryDomain/org/basename
	org := opts.Organization
	if org == "" {
		org = DefaultOrg
	}

	return DefaultRegistryDomain + "/" + org + "/" + basename
}

// convertImageTag returns the converted image tag
func convertImageTag(tag string, isDynamic bool) string {
	if tag == "" {
		return DefaultImageTag
	}

	// Remove anything after and including the first hyphen
	if hyphenIndex := strings.Index(tag, "-"); hyphenIndex != -1 {
		tag = tag[:hyphenIndex]
	}

	// If tag is dynamic, just add -dev
	return tag + "-dev"
}

// isUbuntuImage checks if the base is ubuntu
func isUbuntuImage(base string) bool {
	// Extract the basename (part after the last slash)
	var basename string
	if lastSlash := strings.LastIndex(base, "/"); lastSlash != -1 {
		basename = base[lastSlash+1:]
	} else {
		basename = base
	}

	return basename == "ubuntu"
}

// convertPackageManagerCommands converts package manager commands in a shell command
// to the Alpine equivalent (apk add)
func convertPackageManagerCommands(shell *ShellCommand, packageMap PackageMap) (bool, Distro, Manager, []string, *ShellCommand, error) {
	if shell == nil {
		return false, "", "", nil, nil, nil
	}

	// Determine which distro/package manager we're going to focus on
	var distro Distro
	var firstPM Manager
	var firstPMInstallIndex = -1
	packagesDetected := []string{}
	packagesToInstall := []string{}

	// Process all shell parts in one pass
	for i, part := range shell.Parts {
		if Manager(part.Command) == firstPM || (firstPM == "" && PackageManagerInfoMap[Manager(part.Command)].Distro != "") {
			// Set the package manager if it's the first one we've found
			if firstPM == "" {
				firstPM = Manager(part.Command)
				distro = PackageManagerInfoMap[firstPM].Distro
			}

			// Check if this is an install command
			pmInfo := PackageManagerInfoMap[firstPM]
			if len(part.Args) > 0 && part.Args[0] == pmInfo.InstallKeyword {
				if firstPMInstallIndex == -1 {
					firstPMInstallIndex = i
				}

				// Collect packages, applying mapping if available
				for _, arg := range part.Args[1:] {
					if !strings.HasPrefix(arg, "-") {
						packagesDetected = append(packagesDetected, arg)

						if distroMap, exists := packageMap[distro]; exists && distroMap[arg] != nil {
							packagesToInstall = append(packagesToInstall, distroMap[arg]...)
						} else {
							packagesToInstall = append(packagesToInstall, arg)
						}
					}
				}
			}
		}
	}

	// If we have no packages to install, nothing to do
	if len(packagesToInstall) == 0 || firstPMInstallIndex == -1 {
		return false, distro, firstPM, nil, shell, nil
	}

	// Sort and deduplicate packages
	slices.Sort(packagesDetected)
	packagesDetected = slices.Compact(packagesDetected)
	slices.Sort(packagesToInstall)
	packagesToInstall = slices.Compact(packagesToInstall)

	// Create new shell parts, preserving the original order
	newParts := make([]*ShellPart, 0, len(shell.Parts))

	// Add parts before the first package manager install command (non-PM only)
	for i := 0; i < firstPMInstallIndex; i++ {
		if Manager(shell.Parts[i].Command) != firstPM {
			newParts = append(newParts, cloneShellPart(shell.Parts[i]))
		}
	}

	// Add the apk add command
	delimiter := ""
	if firstPMInstallIndex < len(shell.Parts)-1 {
		delimiter = shell.Parts[firstPMInstallIndex].Delimiter
	}
	newParts = append(newParts, &ShellPart{
		ExtraPre:  shell.Parts[firstPMInstallIndex].ExtraPre,
		Command:   "apk",
		Args:      append([]string{"add", "-U"}, packagesToInstall...),
		Delimiter: delimiter,
	})

	// Add remaining parts (non-PM only)
	for i := firstPMInstallIndex + 1; i < len(shell.Parts); i++ {
		if Manager(shell.Parts[i].Command) != firstPM {
			part := cloneShellPart(shell.Parts[i])
			// Last part should have no delimiter
			if i == len(shell.Parts)-1 {
				part.Delimiter = ""
			}
			newParts = append(newParts, part)
		}
	}

	// Fix the last delimiter
	if len(newParts) > 0 {
		newParts[len(newParts)-1].Delimiter = ""
	}

	return true, distro, firstPM, packagesDetected, &ShellCommand{Parts: newParts}, nil
}

// Helper function to clone a shell part
func cloneShellPart(part *ShellPart) *ShellPart {
	newPart := &ShellPart{
		ExtraPre:  part.ExtraPre,
		Command:   part.Command,
		Delimiter: part.Delimiter,
	}
	if part.Args != nil {
		newPart.Args = make([]string, len(part.Args))
		copy(newPart.Args, part.Args)
	}
	return newPart
}
