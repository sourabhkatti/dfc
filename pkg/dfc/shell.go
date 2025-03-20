/*
Copyright 2025 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package dfc

import (
	"strings"
)

// ShellCommand represents a parsed shell command or group of commands
type ShellCommand struct {
	Parts []*ShellPart // The parsed parts of this command
}

// ShellPart represents a single part of a shell command
type ShellPart struct {
	ExtraPre  string   // Environment variable dcecalrations and other command prefixes
	Command   string   // The command such as "apt-get"
	Args      []string // All the args such as "install" "-y" "nano" "vim" (includes pipe character)
	Delimiter string   // The delimiter for this part, such as "&&" or "||" or ";"
}

const partSeparator = " \\\n    "

// String converts a ShellCommand back to its string representation
func (sc *ShellCommand) String() string {
	// If no parts, return "true" as fallback
	if len(sc.Parts) == 0 {
		return "true"
	}

	s := ""
	for i, part := range sc.Parts {
		if i != 0 {
			s += partSeparator
		}
		if part.ExtraPre != "" {
			s += part.ExtraPre + " "
		}
		s += part.Command
		if len(part.Args) > 0 {
			s += " " + strings.Join(part.Args, " ")
		}
		if part.Delimiter != "" {
			s += " " + part.Delimiter
		}
	}
	return s
}

// ParseMultilineShell parses a shell command into a structured representation
func ParseMultilineShell(raw string) *ShellCommand {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	// Remove comments and normalize whitespace
	cleaned := removeComments(raw)
	if strings.TrimSpace(cleaned) == "" {
		return nil
	}

	// Known delimiters
	delimiters := []string{"&&", "||", ";", "|", "&"}

	var parts []*ShellPart
	remainingCmd := strings.TrimSpace(cleaned)

	for len(remainingCmd) > 0 {
		// Find next delimiter not inside quotes, parentheses, or subshells
		nextDelim, nextDelimPos := findNextDelimiter(remainingCmd, delimiters)
		if nextDelimPos == -1 {
			// No more delimiters, this is the last part
			part := parseShellPart(remainingCmd, "")
			parts = append(parts, part)
			break
		}

		// Split command into current part and remaining
		currentCmdRaw := strings.TrimSpace(remainingCmd[:nextDelimPos])
		part := parseShellPart(currentCmdRaw, nextDelim)
		parts = append(parts, part)

		// Move past the delimiter for next iteration
		remainingCmd = strings.TrimSpace(remainingCmd[nextDelimPos+len(nextDelim):])
	}

	return &ShellCommand{Parts: parts}
}

// removeComments removes all comments from the command string and normalizes newlines
func removeComments(input string) string {
	var result strings.Builder
	lines := strings.Split(input, "\n")

	for i, line := range lines {
		// Find comment position (if any)
		commentPos := -1
		inSingleQuote := false
		inDoubleQuote := false

		for j := 0; j < len(line); j++ {
			if line[j] == '\'' && !inDoubleQuote {
				inSingleQuote = !inSingleQuote
				continue
			}
			if line[j] == '"' && !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
				continue
			}

			// Only treat # as comment marker if outside quotes
			if line[j] == '#' && !inSingleQuote && !inDoubleQuote {
				commentPos = j
				break
			}
		}

		// Process line with possible comment removal
		var processedLine string
		if commentPos >= 0 {
			processedLine = strings.TrimSpace(line[:commentPos])
		} else {
			processedLine = strings.TrimSpace(line)
		}

		if processedLine != "" {
			// Check if the line ends with a backslash (line continuation)
			if strings.HasSuffix(processedLine, "\\") && i < len(lines)-1 {
				// Add the line without the trailing backslash
				result.WriteString(strings.TrimSpace(processedLine[:len(processedLine)-1]))
				result.WriteString(" ") // Just add a space instead of a newline
			} else {
				result.WriteString(processedLine)
				result.WriteString(" ") // Add space to separate from next line
			}
		}
	}

	return strings.TrimSpace(result.String())
}

// findNextDelimiter finds the position of the next delimiter not inside quotes/parentheses
func findNextDelimiter(cmd string, delimiters []string) (string, int) {
	inSingleQuote := false
	inDoubleQuote := false
	parenDepth := 0
	backtickDepth := 0
	subshellDepth := 0

	for i := 0; i < len(cmd); i++ {
		// Check for quote start/end
		if cmd[i] == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			continue
		}
		if cmd[i] == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			continue
		}

		// Check for parentheses
		if !inSingleQuote && !inDoubleQuote {
			if cmd[i] == '(' {
				parenDepth++
				continue
			}
			if cmd[i] == ')' && parenDepth > 0 {
				parenDepth--
				continue
			}

			// Check for backticks
			if cmd[i] == '`' {
				backtickDepth = 1 - backtickDepth // Toggle between 0 and 1
				continue
			}

			// Check for subshell $()
			if i < len(cmd)-1 && cmd[i] == '$' && cmd[i+1] == '(' {
				subshellDepth++
				i++ // Skip the next character
				continue
			}
			if cmd[i] == ')' && subshellDepth > 0 {
				subshellDepth--
				continue
			}

			// Skip comments
			if cmd[i] == '#' {
				break // Ignore everything until the end of this segment
			}

			// Only check for delimiters when not in any quotes or special sections
			if !inSingleQuote && !inDoubleQuote && parenDepth == 0 && backtickDepth == 0 && subshellDepth == 0 {
				for _, delim := range delimiters {
					if i+len(delim) <= len(cmd) && cmd[i:i+len(delim)] == delim {
						return delim, i
					}
				}
			}
		}
	}

	return "", -1
}

// parseShellPart parses a command part into command and args
func parseShellPart(cmdPart string, delimiter string) *ShellPart {
	cmdPart = strings.TrimSpace(cmdPart)

	// Special handling for parenthesized commands
	if strings.HasPrefix(cmdPart, "(") && strings.HasSuffix(cmdPart, ")") {
		return &ShellPart{
			Command:   cmdPart,
			Args:      nil,
			Delimiter: delimiter,
		}
	}

	// Tokenize the command part, respecting quotes
	tokens := tokenize(cmdPart)
	if len(tokens) == 0 {
		return &ShellPart{
			Command:   "",
			Delimiter: delimiter,
		}
	}

	// Find the actual command by skipping environment variable declarations
	commandIndex := findCommandIndex(tokens)

	// If we can't find a command after env vars, use all env vars as the command
	if commandIndex >= len(tokens) {
		return &ShellPart{
			Command:   strings.Join(tokens, " "),
			Args:      nil,
			Delimiter: delimiter,
		}
	}

	// Extract environment variables to ExtraPre and the actual command
	var extraPre string
	if commandIndex > 0 {
		extraPre = strings.Join(tokens[:commandIndex], " ")
	}

	return &ShellPart{
		ExtraPre:  extraPre,
		Command:   tokens[commandIndex],
		Args:      tokens[commandIndex+1:],
		Delimiter: delimiter,
	}
}

// findCommandIndex finds the index of the first token that's not an environment variable declaration
func findCommandIndex(tokens []string) int {
	for i, token := range tokens {
		// If it doesn't look like an env var assignment, consider it the command
		if !isEnvVarAssignment(token) {
			return i
		}
	}
	return len(tokens) // All tokens are env vars
}

// isEnvVarAssignment checks if a token is an environment variable assignment
func isEnvVarAssignment(token string) bool {
	// Check for assignment pattern (VAR=value)
	for i, c := range token {
		if c == '=' {
			// Make sure there's at least one character before '='
			return i > 0
		}
	}
	return false
}

// tokenize splits a command into tokens, respecting quotes
func tokenize(cmd string) []string {
	var tokens []string
	var currentToken strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	parenDepth := 0
	backtickDepth := 0
	subshellDepth := 0
	inToken := false

	for i := 0; i < len(cmd); i++ {
		char := cmd[i]

		// Handle quotes
		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			inToken = true
			currentToken.WriteByte(char)
			continue
		}
		if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			inToken = true
			currentToken.WriteByte(char)
			continue
		}

		// Handle parentheses, backticks, and subshells
		if !inSingleQuote && !inDoubleQuote {
			if char == '(' {
				parenDepth++
				inToken = true
				currentToken.WriteByte(char)
				continue
			}
			if char == ')' {
				parenDepth--
				inToken = true
				currentToken.WriteByte(char)
				continue
			}
			if char == '`' {
				backtickDepth = 1 - backtickDepth
				inToken = true
				currentToken.WriteByte(char)
				continue
			}
			if i < len(cmd)-1 && char == '$' && cmd[i+1] == '(' {
				subshellDepth++
				inToken = true
				currentToken.WriteByte(char)
				continue
			}
			if char == ')' && subshellDepth > 0 {
				subshellDepth--
				inToken = true
				currentToken.WriteByte(char)
				continue
			}
		}

		// Handle spaces to separate tokens
		if char == ' ' || char == '\t' {
			if inSingleQuote || inDoubleQuote || parenDepth > 0 || backtickDepth > 0 || subshellDepth > 0 {
				// Space inside quotes or special constructs - keep it
				inToken = true
				currentToken.WriteByte(char)
			} else if inToken {
				// End of a token
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
				inToken = false
			}
		} else {
			// Regular character
			inToken = true
			currentToken.WriteByte(char)
		}
	}

	// Add the last token if any
	if inToken {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}
