//go:build integration
// +build integration

/*
Copyright 2025 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntegrationBuild tests building the Dockerfiles from the after files
// This test is only run when the -tags=integration flag is passed to go test
func TestIntegrationBuild(t *testing.T) {
	fmt.Println("Running TestIntegrationBuild")

	// Find all .after.Dockerfile files in the testdata directory
	files, err := filepath.Glob("../testdata/*.after.Dockerfile")
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	t.Logf("Found %d .after.Dockerfile files, will only test those with corresponding directories", len(files))
	availableTests := 0

	for _, file := range files {
		// Extract the test name from the file path
		// e.g., "../testdata/django.after.Dockerfile" -> "django"
		baseName := filepath.Base(file)
		testName := strings.TrimSuffix(baseName, ".after.Dockerfile")

		// Check if a directory with the same name exists
		dirPath := filepath.Join("../testdata", fmt.Sprintf("%s-integration", testName))
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			// Skip tests without a corresponding directory
			t.Logf("Skipping %s: no directory at %s", testName, dirPath)
			continue
		}

		availableTests++
		t.Logf("Will run test for %s using context directory %s", testName, dirPath)

		// Run the test for this file
		t.Run(testName, func(t *testing.T) {
			// Get the after file path
			afterFile := filepath.Join("../testdata", testName+".after.Dockerfile")

			// Check if the after file exists
			if _, err := os.Stat(afterFile); os.IsNotExist(err) {
				t.Fatalf("After file %s does not exist", afterFile)
			}

			// Build the Docker image from the temp Dockerfile
			tagName := fmt.Sprintf("dfc-%s-after:test", testName)
			buildImage(t, afterFile, tagName, dirPath)

			// Clean up the image after the test
			defer cleanupImage(t, tagName)
		})
	}

	if availableTests == 0 {
		t.Log("No tests were run because no directories matching test names were found")
	} else {
		t.Logf("Successfully ran %d integration tests", availableTests)
	}
}

// buildImage builds a Docker image from the given Dockerfile
func buildImage(t *testing.T, dockerfilePath, tagName, contextDir string) {
	t.Logf("Building Docker image %s using %s with context %s", tagName, dockerfilePath, contextDir)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Prepare the Docker build command
	cmd := exec.CommandContext(ctx, "docker", "build", "--progress=plain", "-t", tagName, "-f", dockerfilePath, contextDir)

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start Docker build: %v", err)
	}

	// Stream stdout in real-time
	go streamOutput(t, stdoutPipe, "stdout")

	// Stream stderr in real-time
	go streamOutput(t, stderrPipe, "stderr")

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("Docker build timed out after 5 minutes")
		}
		t.Fatalf("Docker build failed: %v", err)
	}

	t.Logf("Successfully built Docker image %s", tagName)
}

// cleanupImage removes a Docker image
func cleanupImage(t *testing.T, tagName string) {
	t.Logf("Cleaning up Docker image %s", tagName)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Prepare the Docker rmi command
	cmd := exec.CommandContext(ctx, "docker", "rmi", tagName)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to remove Docker image %s: %v\nOutput: %s", tagName, err, output)
		return
	}

	t.Logf("Successfully removed Docker image %s", tagName)
}

// streamOutput reads from a pipe and logs each line to the test logger
func streamOutput(t *testing.T, pipe io.ReadCloser, name string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		t.Logf("[%s] %s", name, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Error reading from %s: %v", name, err)
	}
}
