package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInitSubcommand(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build the gopher binary
	gopherBinary := buildGopher(t, tmpDir)

	// Run the gopher init command
	cmd := exec.Command(gopherBinary, "init", "maciakl/test")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gopher init command failed: %v\nOutput: %s", err, output)
	}

	// Check if the project directory was created
	projectDir := filepath.Join(tmpDir, "test")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Fatalf("project directory was not created: %s", projectDir)
	}

	// Check for the existence of expected files
	expectedFiles := []string{
		".gitignore",
		"go.mod",
		"README.md",
		"main.go",
		"main_test.go",
		".goreleaser.yaml",
	}
	for _, file := range expectedFiles {
		filePath := filepath.Join(projectDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("expected file was not created: %s", filePath)
		}
	}

	// Check the content of go.mod
	goModPath := filepath.Join(projectDir, "go.mod")
	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod file: %v", err)
	}
	if !strings.Contains(string(goModContent), "module github.com/maciakl/test") {
		t.Errorf("go.mod does not contain the correct module path. Content: %s", string(goModContent))
	}
}

func TestVersionSubcommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gopherBinary := buildGopher(t, tmpDir)

	aliases := []string{"version", "-version", "--version", "-v"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			cmd := exec.Command(gopherBinary, alias)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("gopher %s command failed: %v\nOutput: %s", alias, err, output)
			}

			if !strings.Contains(string(output), "Gopher v") {
				t.Errorf("expected output to contain 'Gopher v', but it didn't. Output: %s", output)
			}
		})
	}
}

func TestHelpSubcommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gopherBinary := buildGopher(t, tmpDir)

	aliases := []string{"help", "-help", "--help", "-h"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			cmd := exec.Command(gopherBinary, alias)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("gopher %s command failed: %v\nOutput: %s", alias, err, output)
			}

			if !strings.Contains(string(output), "Usage: gopher [subcommand] <arguments>") {
				t.Errorf("expected output to contain usage information, but it didn't. Output: %s", output)
			}
		})
	}
}

func TestNoArguments(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gopherBinary := buildGopher(t, tmpDir)

	cmd := exec.Command(gopherBinary)
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected command to fail, but it succeeded")
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, but got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatalf("expected ExitError, but got %T: %v", err, err)
	}

	if !strings.Contains(string(output), "Missing subcommand.") {
		t.Errorf("expected output to contain 'Missing subcommand.', but it didn't. Output: %s", output)
	}
	if !strings.Contains(string(output), "Usage: gopher [subcommand] <arguments>") {
		t.Errorf("expected output to contain usage information, but it didn't. Output: %s", output)
	}
}

func TestMakeSubcommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy go.mod file
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
		t.Fatalf("failed to create dummy go.mod file: %v", err)
	}

	gopherBinary := buildGopher(t, tmpDir)

	cmd := exec.Command(gopherBinary, "make")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gopher make command failed: %v\nOutput: %s", err, output)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "Makefile")); os.IsNotExist(err) {
		t.Errorf("Makefile was not created")
	}
}

func TestJustSubcommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy go.mod file
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
		t.Fatalf("failed to create dummy go.mod file: %v", err)
	}

	gopherBinary := buildGopher(t, tmpDir)

	cmd := exec.Command(gopherBinary, "just")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gopher just command failed: %v\nOutput: %s", err, output)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "Justfile")); os.IsNotExist(err) {
		t.Errorf("Justfile was not created")
	}
}

func TestScoopSubcommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create dummy files and directories
	distDir := filepath.Join(tmpDir, "dist")
	if err := os.Mkdir(distDir, 0755); err != nil {
		t.Fatalf("failed to create dist directory: %v", err)
	}
	checksumsFile := filepath.Join(distDir, "myproject_1.2.3_checksums.txt")
	checksumsContent := "0123456789abcdef  myproject_1.2.3_Windows_x86_64.zip"
	if err := os.WriteFile(checksumsFile, []byte(checksumsContent), 0644); err != nil {
		t.Fatalf("failed to create dummy checksums file: %v", err)
	}
	goModFile := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModFile, []byte("module github.com/maciakl/myproject"), 0644); err != nil {
		t.Fatalf("failed to create dummy go.mod file: %v", err)
	}
	mainGoFile := filepath.Join(tmpDir, "myproject.go")
	if err := os.WriteFile(mainGoFile, []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
		t.Fatalf("failed to create dummy main go file: %v", err)
	}

	gopherBinary := buildGopher(t, tmpDir)

	cmd := exec.Command(gopherBinary, "scoop")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gopher scoop command failed: %v\nOutput: %s", err, output)
	}

	scoopFile := filepath.Join(distDir, "myproject.json")
	if _, err := os.Stat(scoopFile); os.IsNotExist(err) {
		t.Errorf("scoop file was not created")
	}

	scoopContent, err := os.ReadFile(scoopFile)
	if err != nil {
		t.Fatalf("failed to read scoop file: %v", err)
	}
	if !strings.Contains(string(scoopContent), "\"version\": \"1.2.3\"") {
		t.Errorf("scoop file does not contain the correct version")
	}
	if !strings.Contains(string(scoopContent), "\"hash\": \"0123456789abcdef\"") {
		t.Errorf("scoop file does not contain the correct hash")
	}
}

func TestBumpSubcommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goModFile := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModFile, []byte("module myproject"), 0644); err != nil {
		t.Fatalf("failed to create dummy go.mod file: %v", err)
	}
	mainGoFile := filepath.Join(tmpDir, "myproject.go")
	if err := os.WriteFile(mainGoFile, []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
		t.Fatalf("failed to create dummy main go file: %v", err)
	}

	gopherBinary := buildGopher(t, tmpDir)

	testCases := []struct {
		bumpType    string
		expectedVer string
	}{
		{"patch", "1.2.4"},
		{"minor", "1.3.0"},
		{"major", "2.0.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.bumpType, func(t *testing.T) {
			// Reset the file content
			if err := os.WriteFile(mainGoFile, []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
				t.Fatalf("failed to reset dummy main go file: %v", err)
			}

			cmd := exec.Command(gopherBinary, "bump", tc.bumpType)
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("gopher bump command failed: %v\nOutput: %s", err, output)
			}

			content, err := os.ReadFile(mainGoFile)
			if err != nil {
				t.Fatalf("failed to read main go file: %v", err)
			}
			if !strings.Contains(string(content), fmt.Sprintf(`const version = "%s"`, tc.expectedVer)) {
				t.Errorf("expected version to be bumped to %s, but it wasn't. Content: %s", tc.expectedVer, content)
			}
		})
	}
}

// Unit tests for individual functions

func TestFindInFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	file := filepath.Join(tmpDir, "test.txt")
	content := "hello world\nfind me here\ngoodbye"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	line, err := findInFile(file, "find me")
	if err != nil {
		t.Fatalf("findInFile failed: %v", err)
	}
	if line != "find me here" {
		t.Errorf("expected to find 'find me here', but got '%s'", line)
	}

	line, err = findInFile(file, "not found")
	if err != nil {
		t.Fatalf("findInFile failed: %v", err)
	}
	if line != "" {
		t.Errorf("expected to find nothing, but got '%s'", line)
	}
}

func TestReplaceInFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	file := filepath.Join(tmpDir, "test.txt")
	content := "hello world\nreplace me here\ngoodbye"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := replaceInFile(file, "replace me here", "i am replaced"); err != nil {
		t.Fatalf("replaceInFile failed: %v", err)
	}

	newContent, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	if !strings.Contains(string(newContent), "i am replaced") {
		t.Errorf("expected file content to be replaced, but it wasn't. Content: %s", newContent)
	}
}

func TestGetName(t *testing.T) {
	testCases := []struct {
		uri      string
		expected string
	}{
		{"github.com/maciakl/test", "test"},
		{"maciakl/test", "test"},
		{"test", "test"},
	}

	for _, tc := range testCases {
		if name := getName(tc.uri); name != tc.expected {
			t.Errorf("expected name to be %s, but got %s", tc.expected, name)
		}
	}
}

func TestGetUsername(t *testing.T) {
	testCases := []struct {
		uri      string
		expected string
	}{
		{"github.com/maciakl/test", "maciakl"},
		{"maciakl/test", "maciakl"},
	}

	for _, tc := range testCases {
		if username := getUsername(tc.uri); username != tc.expected {
			t.Errorf("expected username to be %s, but got %s", tc.expected, username)
		}
	}
}

func TestGetVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	file := filepath.Join(tmpDir, "main.go")
	content := `package main; const version = "1.2.3"`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if version := getVersion(file); version != "1.2.3" {
		t.Errorf("expected version to be '1.2.3', but got '%s'", version)
	}
}

func TestGetModuleName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	file := filepath.Join(tmpDir, "go.mod")
	content := "module github.com/maciakl/myproject"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Temporarily change the current directory to the temporary directory
	// because getModuleName internally calls findInFile("go.mod", ...)
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if name := getModuleName(); name != "myproject" {
		t.Errorf("expected module name to be 'myproject', but got '%s'", name)
	}
}

func TestGetModule(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopher-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	file := filepath.Join(tmpDir, "go.mod")
	content := "module github.com/maciakl/myproject"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if module := getModule(); module != "github.com/maciakl/myproject" {
		t.Errorf("expected module to be 'github.com/maciakl/myproject', but got '%s'", module)
	}
}

func TestIncString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"0", "1"},
		{"9", "10"},
		{"123", "124"},
	}

	for _, tc := range testCases {
		if result := incString(tc.input); result != tc.expected {
			t.Errorf("expected %s, but got %s", tc.expected, result)
		}
	}
}

// buildGopher builds the gopher binary and returns the path to the binary.
func buildGopher(t *testing.T, tmpDir string) string {
	t.Helper()
	gopherBinary := filepath.Join(tmpDir, "gopher")
	if runtime.GOOS == "windows" {
		gopherBinary += ".exe"
	}

	buildCmd := exec.Command("go", "build", "-o", gopherBinary)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build gopher binary: %v", err)
	}
	return gopherBinary
}
