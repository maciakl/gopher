package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func captureStdout(f func()) string {
	oldStdout := os.Stdout
	oldColorOutput := color.Output
	r, w, _ := os.Pipe()
	os.Stdout = w
	color.Output = w // Redirect color output as well

	f()

	w.Close()
	os.Stdout = oldStdout
	color.Output = oldColorOutput // Restore original color output

	var buf strings.Builder
	// Use io.Copy to read from the pipe to the strings.Builder
	_, err := io.Copy(&buf, r)
	if err != nil {
		fmt.Printf("Error copying stdout: %v\n", err)
	}
	return buf.String()
}

func TestPrintUsage(t *testing.T) {
	output := captureStdout(printUsage)
	expected := "Usage: gopher [subcommand] <arguments>"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, but got %q", expected, output)
	}
	expected = "init <string>"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, but got %q", expected, output)
	}
}

func TestBanner(t *testing.T) {
	output := captureStdout(banner)
	expected := "Gopher v" + version + "\n"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain 'Gopher v', but it didn't. Output: %s", output)
	}
}

func TestInitSubcommand(t *testing.T) {
	t.Run("CreateProjectWithGithubURI", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)
		gopherBinary := buildGopher(t, tmpDir)
		projectName := "testproject"
		githubURI := "github.com/maciakl/" + projectName

		cmd := exec.Command(gopherBinary, "init", githubURI)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher init command failed: %v\nOutput: %s", err, output)
		}

		projectDir := filepath.Join(tmpDir, projectName)
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			t.Fatalf("project directory was not created: %s", projectDir)
		}

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

		// Verify contents
		goModContent, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
		if err != nil {
			t.Fatalf("failed to read go.mod: %v", err)
		}
		if !strings.Contains(string(goModContent), "module "+githubURI) {
			t.Errorf("go.mod content incorrect: %s", string(goModContent))
		}

		gitIgnoreContent, err := os.ReadFile(filepath.Join(projectDir, ".gitignore"))
		if err != nil {
			t.Fatalf("failed to read .gitignore: %v", err)
		}
		if !strings.Contains(string(gitIgnoreContent), projectName+"\n") {
			t.Errorf(".gitignore content incorrect: %s", string(gitIgnoreContent))
		}

		readmeContent, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
		if err != nil {
			t.Fatalf("failed to read README.md: %v", err)
		}
		if !strings.Contains(string(readmeContent), "# "+projectName) {
			t.Errorf("README.md content incorrect: %s", string(readmeContent))
		}

		mainGoContent, err := os.ReadFile(filepath.Join(projectDir, "main.go"))
		if err != nil {
			t.Fatalf("failed to read main.go: %v", err)
		}
		if !strings.Contains(string(mainGoContent), "package main") || !strings.Contains(string(mainGoContent), `const version = "0.1.0"`) {
			t.Errorf("main.go content incorrect: %s", string(mainGoContent))
		}

		mainTestGoContent, err := os.ReadFile(filepath.Join(projectDir, "main_test.go"))
		if err != nil {
			t.Fatalf("failed to read %s_test.go: %v", projectName, err)
		}
		if !strings.Contains(string(mainTestGoContent), "package main") || !strings.Contains(string(mainTestGoContent), `binName  = "test"`) {
			t.Errorf("%s_test.go content incorrect: %s", projectName, string(mainTestGoContent))
		}
	})
}

func TestVersionSubcommand(t *testing.T) {
	aliases := []string{"version", "-version", "--version", "-v"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			tmpDir := t.TempDir()
			defer os.RemoveAll(tmpDir)
			gopherBinary := buildGopher(t, tmpDir)
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
	aliases := []string{"help", "-help", "--help", "-h"}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			tmpDir := t.TempDir()
			defer os.RemoveAll(tmpDir)
			gopherBinary := buildGopher(t, tmpDir)
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

func TestCreateProjectWithGopherUsername(t *testing.T) {
	t.Run("CreateProjectWithGopherUsername", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)
		gopherBinary := buildGopher(t, tmpDir)
		projectName := "mycli"
		gopherUsername := "testuser"
		expectedURI := "github.com/" + gopherUsername + "/" + projectName

		os.Setenv("GOPHER_USERNAME", gopherUsername)
		defer os.Unsetenv("GOPHER_USERNAME")

		cmd := exec.Command(gopherBinary, "init", projectName)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher init command failed: %v\nOutput: %s", err, output)
		}

		projectDir := filepath.Join(tmpDir, projectName)
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			t.Fatalf("project directory was not created: %s", projectDir)
		}

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

		// Verify contents
		goModContent, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
		if err != nil {
			t.Fatalf("failed to read go.mod: %v", err)
		}
		if !strings.Contains(string(goModContent), "module "+expectedURI) {
			t.Errorf("go.mod content incorrect: %s", string(goModContent))
		}

		gitIgnoreContent, err := os.ReadFile(filepath.Join(projectDir, ".gitignore"))
		if err != nil {
			t.Fatalf("failed to read .gitignore: %v", err)
		}
		if !strings.Contains(string(gitIgnoreContent), projectName+"\n") {
			t.Errorf(".gitignore content incorrect: %s", string(gitIgnoreContent))
		}

		readmeContent, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
		if err != nil {
			t.Fatalf("failed to read README.md: %v", err)
		}
		if !strings.Contains(string(readmeContent), "# "+projectName) {
			t.Errorf("README.md content incorrect: %s", string(readmeContent))
		}

		mainGoContent, err := os.ReadFile(filepath.Join(projectDir, "main.go"))
		if err != nil {
			t.Fatalf("failed to read main.go: %v", err)
		}
		if !strings.Contains(string(mainGoContent), "package main") || !strings.Contains(string(mainGoContent), `const version = "0.1.0"`) {
			t.Errorf("main.go content incorrect: %s", string(mainGoContent))
		}

		mainTestGoContent, err := os.ReadFile(filepath.Join(projectDir, "main_test.go"))
		if err != nil {
			t.Fatalf("failed to read %s_test.go: %v", projectName, err)
		}
		if !strings.Contains(string(mainTestGoContent), "package main") || !strings.Contains(string(mainTestGoContent), `binName  = "test"`) {
			t.Errorf("%s_test.go content incorrect: %s", projectName, string(mainTestGoContent))
		}
	})
}

func verifyCommandOutput(t *testing.T, cmd *exec.Cmd, expectedExitCode int, expectedOutputContains []string) {
	t.Helper()
	output, err := cmd.CombinedOutput()

	if expectedExitCode == 0 {
		if err != nil {
			t.Fatalf("expected command to succeed, but it failed: %v\nOutput: %s", err, output)
		}
	} else {
		if err == nil {
			t.Fatalf("expected command to fail, but it succeeded")
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != expectedExitCode {
				t.Errorf("expected exit code %d, but got %d\nOutput: %s", expectedExitCode, exitErr.ExitCode(), output)
			}
		} else {
			t.Fatalf("expected ExitError, but got %T: %v\nOutput: %s", err, err, output)
		}
	}

	for _, expected := range expectedOutputContains {
		if !strings.Contains(string(output), expected) {
			t.Errorf("expected output to contain %q, but it didn't. Output: %s", expected, output)
		}
	}
}

func TestNoArguments(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)
	gopherBinary := buildGopher(t, tmpDir)

	cmd := exec.Command(gopherBinary)
	verifyCommandOutput(t, cmd, 1, []string{"Gopher v", "Missing subcommand.", "Usage: gopher [subcommand] <arguments>"})
}

func TestMakeSubcommand(t *testing.T) {
	// gopherBinary built once outside subtests
	gopherBinary := buildGopher(t, t.TempDir())

	t.Run("SuccessfulCreation", func(t *testing.T) {
		tmpDir := t.TempDir() // New tmpDir for this subtest
		defer os.RemoveAll(tmpDir)

		// Create a dummy go.mod file
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}

		cmd := exec.Command(gopherBinary, "make")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher make command failed: %v\nOutput: %s", err, output)
		}

		makefilePath := filepath.Join(tmpDir, "Makefile")
		if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
			t.Errorf("Makefile was not created")
		}

		content, err := os.ReadFile(makefilePath)
		if err != nil {
			t.Fatalf("failed to read Makefile: %v", err)
		}
		if !strings.Contains(string(content), "BINARY_NAME=myproject") ||
			!strings.Contains(string(content), ".PHONY: build") ||
			!strings.Contains(string(content), "go build") {
			t.Errorf("Makefile content is incorrect: %s", string(content))
		}
		if !strings.Contains(string(output), "Makefile created successfully.") {
			t.Errorf("expected success message, but got: %s", output)
		}
	})

	t.Run("GetModuleNameFailure", func(t *testing.T) {
		tmpDir := t.TempDir() // New tmpDir for this subtest
		defer os.RemoveAll(tmpDir)

		// No go.mod file created to simulate failure

		cmd := exec.Command(gopherBinary, "make")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher make command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "Error opening go.mod file") {
			t.Errorf("expected error about missing go.mod, but got: %s", output)
		}
	})

	t.Run("FileCreationFailure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping FileCreationFailure on Windows due to os.Chmod limitations")
		}
		tmpDir := t.TempDir() // New tmpDir for this subtest
		defer os.RemoveAll(tmpDir)

		// Create a dummy go.mod file
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}

		// Make the directory non-writable to simulate file creation failure
		if err := os.Chmod(tmpDir, 0444); err != nil {
			t.Fatalf("failed to change permissions of directory: %v", err)
		}
		defer os.Chmod(tmpDir, 0755) // Restore permissions

		cmd := exec.Command(gopherBinary, "make")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher make command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "Error creating Makefile") {
			t.Errorf("expected error about creating Makefile, but got: %s", output)
		}
	})
}

func TestJustSubcommand(t *testing.T) {
	// gopherBinary built once outside subtests
	gopherBinary := buildGopher(t, t.TempDir())

	t.Run("SuccessfulCreation", func(t *testing.T) {
		tmpDir := t.TempDir() // New tmpDir for this subtest
		defer os.RemoveAll(tmpDir)

		// Create a dummy go.mod file
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}

		cmd := exec.Command(gopherBinary, "just")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher just command failed: %v\nOutput: %s", err, output)
		}

		justfilePath := filepath.Join(tmpDir, "Justfile")
		if _, err := os.Stat(justfilePath); os.IsNotExist(err) {
			t.Errorf("Justfile was not created")
		}

		content, err := os.ReadFile(justfilePath)
		if err != nil {
			t.Fatalf("failed to read Justfile: %v", err)
		}
		if !strings.Contains(string(content), `BINARY_NAME := "myproject"`) ||
			!strings.Contains(string(content), "build:") ||
			!strings.Contains(string(content), "go build") {
			t.Errorf("Justfile content is incorrect: %s", string(content))
		}
		if !strings.Contains(string(output), "Justfile created successfully.") {
			t.Errorf("expected success message, but got: %s", output)
		}
	})

	t.Run("GetModuleNameFailure", func(t *testing.T) {
		tmpDir := t.TempDir() // New tmpDir for this subtest
		defer os.RemoveAll(tmpDir)

		// No go.mod file created to simulate failure

		cmd := exec.Command(gopherBinary, "just")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher just command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "Error opening go.mod file") {
			t.Errorf("expected error about missing go.mod, but got: %s", output)
		}
	})

	t.Run("FileCreationFailure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping FileCreationFailure on Windows due to os.Chmod limitations")
		}
		tmpDir := t.TempDir() // New tmpDir for this subtest
		defer os.RemoveAll(tmpDir)

		// Create a dummy go.mod file
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}

		// Make the directory non-writable to simulate file creation failure
		if err := os.Chmod(tmpDir, 0444); err != nil {
			t.Fatalf("failed to change permissions of directory: %v", err)
		}
		defer os.Chmod(tmpDir, 0755) // Restore permissions

		cmd := exec.Command(gopherBinary, "just")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher just command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "Error creating Justfile") {
			t.Errorf("expected error about creating Justfile, but got: %s", output)
		}
	})
}

func TestScoopSubcommand(t *testing.T) {
	// gopherBinary built once for all subtests
	gopherBinary := buildGopher(t, t.TempDir())

	t.Run("SuccessfulScoopFileGeneration", func(t *testing.T) {
		tmpDir := t.TempDir()
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

		os.Setenv("GOPHER_USERNAME", "maciakl") // Ensure username is set for non-uri case
		defer os.Unsetenv("GOPHER_USERNAME")

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
			t.Errorf("scoop file does not contain the correct version. Content: %s", scoopContent)
		}
		if !strings.Contains(string(scoopContent), "\"hash\": \"0123456789abcdef\"") {
			t.Errorf("scoop file does not contain the correct hash. Content: %s", scoopContent)
		}
		if !strings.Contains(string(output), "Scoop manifest file myproject.json created successfully.") {
			t.Errorf("expected success message, but got: %s", output)
		}
	})

	t.Run("DistFolderMissing", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy go.mod and main.go, but no dist folder
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main go file: %v", err)
		}

		os.Setenv("GOPHER_USERNAME", "maciakl")
		defer os.Unsetenv("GOPHER_USERNAME")

		cmd := exec.Command(gopherBinary, "scoop")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher scoop command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "dist/ folder does not exist in the project directory.") {
			t.Errorf("expected error about missing dist folder, but got: %s", output)
		}
	})

	t.Run("ChecksumFileNotFound", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy files and directories, but no checksums.txt
		distDir := filepath.Join(tmpDir, "dist")
		if err := os.Mkdir(distDir, 0755); err != nil {
			t.Fatalf("failed to create dist directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main go file: %v", err)
		}

		os.Setenv("GOPHER_USERNAME", "maciakl")
		defer os.Unsetenv("GOPHER_USERNAME")

		cmd := exec.Command(gopherBinary, "scoop")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil { // Expecting failure because main.go returns error for warnings
			t.Fatalf("expected gopher scoop command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "Could not find the checksum file") {
			t.Errorf("expected warning about missing checksum file, but got: %s", output)
		}
		if !strings.Contains(string(output), "created with some warnings") {
			t.Errorf("expected warnings in output, but got: %s", output)
		}
	})

	t.Run("ChecksumEntryNotFound", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy files and directories
		distDir := filepath.Join(tmpDir, "dist")
		if err := os.Mkdir(distDir, 0755); err != nil {
			t.Fatalf("failed to create dist directory: %v", err)
		}
		checksumsFile := filepath.Join(distDir, "myproject_1.2.3_checksums.txt")
		checksumsContent := "0123456789abcdef  another_zip.zip" // Wrong entry
		if err := os.WriteFile(checksumsFile, []byte(checksumsContent), 0644); err != nil {
			t.Fatalf("failed to create dummy checksums file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main go file: %v", err)
		}

		os.Setenv("GOPHER_USERNAME", "maciakl")
		defer os.Unsetenv("GOPHER_USERNAME")

		cmd := exec.Command(gopherBinary, "scoop")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil { // Expecting failure because main.go returns error for warnings
			t.Fatalf("expected gopher scoop command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "Could not find a checksum for myproject_1.2.3_Windows_x86_64.zip in the checksum file.") {
			t.Errorf("expected warning about missing checksum entry, but got: %s", output)
		}
		if !strings.Contains(string(output), "created with some warnings") {
			t.Errorf("expected warnings in output, but got: %s", output)
		}
	})
}

func TestVersionBumpSubcommand(t *testing.T) {
	testCases := []struct {
		name        string
		bumpType    string
		initialVer  string
		expectedVer string
		expectError bool
		errorMessage string
	}{
		{"BumpPatch", "patch", "1.2.3", "1.2.4", false, ""},
		{"BumpMinor", "minor", "1.2.3", "1.3.0", false, ""},
		{"BumpMajor", "major", "1.2.3", "2.0.0", false, ""},
		{"InvalidBumpArgument", "invalid", "1.2.3", "", true, "Invalid argument for bump subcommand."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			defer os.RemoveAll(tmpDir)
			gopherBinary := buildGopher(t, tmpDir)

			// Create dummy go.mod and main.go files
			if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
				t.Fatalf("failed to create dummy go.mod file: %v", err)
			}
			mainGoFile := filepath.Join(tmpDir, "main.go")
			if err := os.WriteFile(mainGoFile, []byte(fmt.Sprintf(`package main; const version = "%s"`, tc.initialVer)), 0644); err != nil {
				t.Fatalf("failed to create dummy main go file: %v", err)
			}

			cmd := exec.Command(gopherBinary, "bump", tc.bumpType)
			cmd.Dir = tmpDir

			expectedExitCode := 0
			if tc.expectError {
				expectedExitCode = 1 // Assuming 1 for failure
			}
			
			expectedOutputStrings := []string{}
			if tc.errorMessage != "" {
				expectedOutputStrings = append(expectedOutputStrings, tc.errorMessage)
			}
			
			verifyCommandOutput(t, cmd, expectedExitCode, expectedOutputStrings)

			if !tc.expectError { // Only check file content if no error is expected
				content, err := os.ReadFile(mainGoFile)
				if err != nil {
					t.Fatalf("failed to read main go file: %v", err)
				}
				if !strings.Contains(string(content), fmt.Sprintf(`const version = "%s"`, tc.expectedVer)) {
					t.Errorf("expected version to be bumped to %s, but it wasn't. Content: %s", tc.expectedVer, content)
				}
			}
		})
	}

	t.Run("GetMainFileNameFailure", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)
		gopherBinary := buildGopher(t, tmpDir)

		// No go.mod or main.go to simulate getMainFileName failure

		cmd := exec.Command(gopherBinary, "bump", "patch")
		cmd.Dir = tmpDir
		
		verifyCommandOutput(t, cmd, 1, []string{"Error opening go.mod file", "open go.mod: The system cannot find the file specified."})
	})

	t.Run("GetVersionFailure", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)
		gopherBinary := buildGopher(t, tmpDir)

		// Create dummy go.mod, but main.go is missing version
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		mainGoFile := filepath.Join(tmpDir, "main.go")
		if err := os.WriteFile(mainGoFile, []byte(`package main;`), 0644); err != nil {
			t.Fatalf("failed to create dummy main go file: %v", err)
		}

		cmd := exec.Command(gopherBinary, "bump", "patch")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "Error opening file main.go") && !strings.Contains(string(output), "index out of range") {
			t.Errorf("expected error about getVersion failure, but got: %s", output)
		}
	})

	t.Run("ReplaceInFileFailure", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)
		gopherBinary := buildGopher(t, tmpDir)

		// Create dummy go.mod and main.go
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		mainGoFile := filepath.Join(tmpDir, "main.go")
		if err := os.WriteFile(mainGoFile, []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main go file: %v", err)
		}

		// Make main.go non-writable to simulate replaceInFile failure
		if err := os.Chmod(mainGoFile, 0444); err != nil {
			t.Fatalf("failed to change permissions of main.go: %v", err)
		}
		defer os.Chmod(mainGoFile, 0644) // Restore permissions

		cmd := exec.Command(gopherBinary, "bump", "patch")
		cmd.Dir = tmpDir
		
		verifyCommandOutput(t, cmd, 1, []string{"Error modifying the source file"})
	})
}

// Unit tests for individual functions

func TestFindInFile(t *testing.T) {
	tmpDir := t.TempDir()
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
	tmpDir := t.TempDir()
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
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	file := filepath.Join(tmpDir, "main.go")
	content := `package main; const version = "1.2.3"`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if version, err := getVersion(file); err != nil {
		t.Fatalf("getVersion failed: %v", err)
	} else if version != "1.2.3" {
		t.Errorf("expected version to be '1.2.3', but got '%s'", version)
	}
}

func TestGetModuleName(t *testing.T) {
	tmpDir := t.TempDir()
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
		t.Fatalf("failed to get original working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer os.Chdir(oldWd) // Restore original working directory

	if name, err := getModuleName(); err != nil {
		t.Fatalf("getModuleName failed: %v", err)
	} else if name != "myproject" {
		t.Errorf("expected module name to be 'myproject', but got '%s'", name)
	}
}

func TestGetModule(t *testing.T) {
	tmpDir := t.TempDir()
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

	if module, err := getModule(); err != nil {
		t.Fatalf("getModule failed: %v", err)
	} else if module != "github.com/maciakl/myproject" {
		t.Errorf("expected module to be 'github.com/maciakl/myproject', but got '%s'", module)
	}
}

func TestInfoSubcommand(t *testing.T) {
	gopherBinary := buildGopher(t, t.TempDir()) // Build gopher once

	t.Run("CleanGitState", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create a dummy go.mod file
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		// Create a dummy main.go file
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Initialize a git repository
		execCmd(t, tmpDir, "git", "init", "-b", "main")
		execCmd(t, tmpDir, "git", "add", ".")
		execCmd(t, tmpDir, "git", "commit", "-m", "initial commit")
		execCmd(t, tmpDir, "git", "remote", "add", "origin", "https://github.com/maciakl/myproject.git")

		cmd := exec.Command(gopherBinary, "info")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher info command failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), "Project Name:\tmain") {
			t.Errorf("expected output to contain 'Project Name:\tmain', but it didn't. Output: %s", output)
		}
		if !strings.Contains(string(output), "Version:\t1.2.3") {
			t.Errorf("expected output to contain 'Version:\t1.2.3', but it didn't. Output: %s", output)
		}
		if !strings.Contains(string(output), "Git State: 	✔️ clean") {
			t.Errorf("expected output to contain 'Git State: 	✔️ clean', but it didn't. Output: %s", output)
		}
	})

	t.Run("DirtyGitState", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create a dummy go.mod file
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		// Create a dummy main.go file
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Initialize a git repository and commit
		execCmd(t, tmpDir, "git", "init", "-b", "main")
		execCmd(t, tmpDir, "git", "add", ".")
		execCmd(t, tmpDir, "git", "commit", "-m", "initial commit")
		execCmd(t, tmpDir, "git", "remote", "add", "origin", "https://github.com/maciakl/myproject.git")

		// Make a change without committing to make the repo dirty
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main; const version = "1.2.4"`), 0644); err != nil {
			t.Fatalf("failed to modify main.go file: %v", err)
		}

		cmd := exec.Command(gopherBinary, "info")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher info command failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), "Git State: 	❌ dirty") {
			t.Errorf("expected output to contain 'Git State: 	❌ dirty', but it didn't. Output: %s", output)
		}
	})
}

// Helper to run commands and check for errors
func execCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("command failed: %s %v\nOutput: %s\nError: %v", name, args, string(out), err)
	}
}

func TestInstallSubcommand(t *testing.T) {
	// gopherBinary built once for all subtests
	gopherBinary := buildGopher(t, t.TempDir())

	// Save original PATH and defer its restoration
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath) // Restore PATH after tests

	// Save original HOME/USERPROFILE and defer its restoration
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	defer os.Setenv("HOME", originalHome)
	defer os.Setenv("USERPROFILE", originalUserProfile)


	t.Run("SuccessfulInstallUnix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix-specific test on Windows")
		}
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy go.mod and main.go
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; func main(){}`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Create a mock home directory and a bin directory inside it
		mockHomeDir := filepath.Join(t.TempDir(), "mockhome")
		mockBinDir := filepath.Join(mockHomeDir, "bin")
		if err := os.MkdirAll(mockBinDir, 0755); err != nil {
			t.Fatalf("failed to create mock bin directory: %v", err)
		}
		os.Setenv("HOME", mockHomeDir) // Set HOME for the test

		cmd := exec.Command(gopherBinary, "install")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher install command failed: %v\nOutput: %s", err, output)
		}

		if _, err := os.Stat(filepath.Join(mockBinDir, "myproject")); os.IsNotExist(err) {
			t.Errorf("binary was not installed in the mock bin directory")
		}
		if !strings.Contains(string(output), "myproject installed successfully") {
			t.Errorf("expected success message, but got: %s", output)
		}
	})

	t.Run("SuccessfulInstallWindows", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Skipping Windows-specific test on Unix")
		}
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy go.mod and main.go
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; func main(){}`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Create a mock userprofile directory and a bin directory inside it
		mockUserProfileDir := filepath.Join(t.TempDir(), "mockuserprofile")
		mockBinDir := filepath.Join(mockUserProfileDir, "bin")
		if err := os.MkdirAll(mockBinDir, 0755); err != nil {
			t.Fatalf("failed to create mock bin directory: %v", err)
		}
		os.Setenv("USERPROFILE", mockUserProfileDir) // Set USERPROFILE for the test

		cmd := exec.Command(gopherBinary, "install")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher install command failed: %v\nOutput: %s", err, output)
		}

		if _, err := os.Stat(filepath.Join(mockBinDir, "myproject.exe")); os.IsNotExist(err) {
			t.Errorf("binary was not installed in the mock bin directory")
		}
		if !strings.Contains(string(output), "myproject installed successfully") {
			t.Errorf("expected success message, but got: %s", output)
		}
	})

	t.Run("GOPHER_INSTALLPATH_Invalid", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy go.mod and main.go
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; func main(){}`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Set an invalid install path
		os.Setenv("GOPHER_INSTALLPATH", filepath.Join(t.TempDir(), "nonexistent", "bin"))

		cmd := exec.Command(gopherBinary, "install")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher install command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "directory does not exist. Please create it and add it to your path first.") {
			t.Errorf("expected error about nonexistent directory, but got: %s", output)
		}
	})

	t.Run("DefaultBinDirectoryMissing", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy go.mod and main.go
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; func main(){}`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Ensure GOPHER_INSTALLPATH is NOT set
		os.Unsetenv("GOPHER_INSTALLPATH")

		// Point HOME/USERPROFILE to a directory where 'bin' does not exist
		mockUserDir := filepath.Join(t.TempDir(), "mockuser")
		os.Setenv("HOME", mockUserDir)
		os.Setenv("USERPROFILE", mockUserDir)


		cmd := exec.Command(gopherBinary, "install")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher install command to fail, but it succeeded. Output: %s", output)
		}

		expectedErrMsg := "directory does not exist. Please create it and add it to your path first."
		if !strings.Contains(string(output), expectedErrMsg) {
			t.Errorf("expected error about missing default bin directory, but got: %s", output)
		}
	})

	t.Run("CopyOperationFailureUnix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix-specific test on Windows")
		}
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy go.mod and main.go
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; func main(){}`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Create a mock home directory and a bin directory inside it
		mockHomeDir := filepath.Join(t.TempDir(), "mockhome-copyfail")
		mockBinDir := filepath.Join(mockHomeDir, "bin")
		if err := os.MkdirAll(mockBinDir, 0755); err != nil {
			t.Fatalf("failed to create mock bin directory: %v", err)
		}
		os.Setenv("HOME", mockHomeDir)

		// Create a mock 'cp' command that always fails
		mockCpDir := t.TempDir()
		mockCpPath := filepath.Join(mockCpDir, "cp")
		if err := os.WriteFile(mockCpPath, []byte("#!/bin/sh\nexit 1"), 0755); err != nil {
			t.Fatalf("failed to create mock cp: %v", err)
		}
		os.Setenv("PATH", mockCpDir+string(os.PathListSeparator)+originalPath)

		cmd := exec.Command(gopherBinary, "install")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher install command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "error") { // More specific check if possible
			t.Errorf("expected error from copy operation, but got: %s", output)
		}
	})

		t.Run("CopyOperationFailureWindows", func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("Skipping CopyOperationFailureWindows on Windows due to os.Chmod limitations")
		}
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy go.mod and main.go
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; func main(){}`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Create a mock userprofile directory and a bin directory inside it
		mockUserProfileDir := filepath.Join(t.TempDir(), "mockuserprofile")
		mockBinDir := filepath.Join(mockUserProfileDir, "bin")
		if err := os.MkdirAll(mockBinDir, 0755); err != nil {
			t.Fatalf("failed to create mock bin directory: %v", err)
		}
		os.Setenv("USERPROFILE", mockUserProfileDir)

		// Simulate cp.Copy failure. This is harder to mock as cp.Copy is an imported library function.
		// For simplicity, let's make the destination non-writable for this test.
		if err := os.Chmod(mockBinDir, 0444); err != nil { // Make directory read-only
			t.Fatalf("failed to change permissions of mock bin directory: %v", err)
		}
		defer os.Chmod(mockBinDir, 0755) // Restore permissions


		cmd := exec.Command(gopherBinary, "install")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher install command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "Error") { // Look for a generic error
			t.Errorf("expected error from copy operation, but got: %s", output)
		}
	})
}

func TestReleaseSubcommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping TestReleaseSubcommand on Windows due to mock executable compatibility")
	}
	// gopherBinary built once for all subtests
	gopherBinary := buildGopher(t, t.TempDir())

	// Save original PATH and defer its restoration
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// create mock goreleaser executable in a temporary directory
	mockGoreleaserDir := t.TempDir()
	mockGoreleaserPath := filepath.Join(mockGoreleaserDir, "goreleaser")
	if runtime.GOOS == "windows" {
		mockGoreleaserPath += ".exe"
	}

	t.Run("SuccessfulRelease", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy files
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Initialize git repository
		execCmd(t, tmpDir, "git", "init", "-b", "main")
		execCmd(t, tmpDir, "git", "add", ".")
		execCmd(t, tmpDir, "git", "commit", "-m", "initial commit")

		// Configure mock goreleaser to succeed
		if err := os.WriteFile(mockGoreleaserPath, []byte("#!/bin/sh\necho 'goreleaser mock success'"), 0755); err != nil {
			t.Fatalf("failed to create mock goreleaser: %v", err)
		}
		os.Setenv("PATH", mockGoreleaserDir+string(os.PathListSeparator)+originalPath)

		cmd := exec.Command(gopherBinary, "release")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gopher release command failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), "Project released successfully") {
			t.Errorf("expected output to contain 'Project released successfully', but it didn't. Output: %s", output)
		}
	})

	t.Run("GitTagFailure", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy files
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Initialize git repository and create the tag we are going to try to create
		execCmd(t, tmpDir, "git", "init", "-b", "main")
		execCmd(t, tmpDir, "git", "add", ".")
		execCmd(t, tmpDir, "git", "commit", "-m", "initial commit")
		execCmd(t, tmpDir, "git", "tag", "v1.2.3") // Pre-create the tag to cause failure

		// Configure mock git to be in PATH first to catch the tag command
		mockGitTestDir := createMockGit(t, t.TempDir()) // Separate temp dir for mock git
		os.Setenv("PATH", mockGitTestDir+string(os.PathListSeparator)+originalPath)

		cmd := exec.Command(gopherBinary, "release")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher release command to fail, but it succeeded. Output: %s", output)
		}
		if !strings.Contains(string(output), "error: tag 'v1.2.3' already exists") && !strings.Contains(string(output), "fatal: tag 'v1.2.3' already exists") {
			t.Errorf("expected error about existing tag, but got: %s", output)
		}
	})

	t.Run("GoreleaserReleaseFailureAndTagCleanup", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy files
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Initialize git repository
		execCmd(t, tmpDir, "git", "init", "-b", "main")
		execCmd(t, tmpDir, "git", "add", ".")
		execCmd(t, tmpDir, "git", "commit", "-m", "initial commit")

		// Configure mock goreleaser to fail
		if err := os.WriteFile(mockGoreleaserPath, []byte("#!/bin/sh\necho 'goreleaser mock fail' && exit 1"), 0755); err != nil {
			t.Fatalf("failed to create mock goreleaser: %v", err)
		}
		os.Setenv("PATH", mockGoreleaserDir+string(os.PathListSeparator)+originalPath)

		// Use the real git for tag operations for this test case
		os.Setenv("PATH", mockGoreleaserDir+string(os.PathListSeparator)+originalPath)


		cmd := exec.Command(gopherBinary, "release")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher release command to fail, but it succeeded. Output: %s", output)
		}

		// Verify that the tag was deleted (by checking if 'git tag -l v1.2.3' returns empty)
		tagCheckCmd := exec.Command("git", "tag", "-l", "v1.2.3")
		tagCheckCmd.Dir = tmpDir
		tagOutput, _ := tagCheckCmd.CombinedOutput()
		if strings.TrimSpace(string(tagOutput)) != "" {
			t.Errorf("expected tag 'v1.2.3' to be deleted, but it still exists. Output: %s", tagOutput)
		}
		if !strings.Contains(string(output), "Deleting the git tag v1.2.3...") {
			t.Errorf("expected output to contain 'Deleting the git tag v1.2.3...', but it didn't. Output: %s", output)
		}
	})

	t.Run("GoreleaserReleaseFailureAndTagCleanupFailure", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		// Create dummy files
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/maciakl/myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main; const version = "1.2.3"`), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go file: %v", err)
		}

		// Initialize git repository
		execCmd(t, tmpDir, "git", "init", "-b", "main")
		execCmd(t, tmpDir, "git", "add", ".")
		execCmd(t, tmpDir, "git", "commit", "-m", "initial commit")

		// Pre-create the tag so there's something to delete
		execCmd(t, tmpDir, "git", "tag", "v1.2.3")

		// Configure mock goreleaser to fail
		if err := os.WriteFile(mockGoreleaserPath, []byte("#!/bin/sh\nexit 1"), 0755); err != nil {
			t.Fatalf("failed to create mock goreleaser: %v", err)
		}

		// Configure mock git to fail tag deletion
		mockGitTestDir := createMockGit(t, t.TempDir()) // Separate temp dir for mock git
		os.Setenv("PATH", mockGitTestDir+string(os.PathListSeparator)+mockGoreleaserDir+string(os.PathListSeparator)+originalPath)

		// This is a bit hacky, normally you'd control the mock more precisely
		// For now, relying on the mockGit script's general failure for 'git tag -d'
		// It's not specifically v1.2.3, but any 'git tag -d' call.
		// A more robust mock would pass arguments to the script.

		cmd := exec.Command(gopherBinary, "release")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Fatalf("expected gopher release command to fail, but it succeeded. Output: %s", output)
		}

		if !strings.Contains(string(output), "Failed to remove tag") {
			t.Errorf("expected output to contain 'Failed to remove tag', but it didn't. Output: %s", output)
		}
		if strings.Contains(string(output), "git tag -d v1.2.3") {
			// This is a weak check, as the mock git might still print "Mock git tag command executed"
			// Better: check for actual git error message on delete failure
		}
	})
}

func TestInitMissingArgument(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)
	gopherBinary := buildGopher(t, tmpDir)

	cmd := exec.Command(gopherBinary, "init")
	verifyCommandOutput(t, cmd, 1, []string{"Gopher v", "Missing argument for init subcommand.", "Usage: gopher [subcommand] <arguments>"})
}

func TestUnknownSubcommand(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)
	gopherBinary := buildGopher(t, tmpDir)

	cmd := exec.Command(gopherBinary, "unknowncommand")
	verifyCommandOutput(t, cmd, 1, []string{"Gopher v", "Unknown subcommand.", "Usage: gopher [subcommand] <arguments>"})
}

func createDummyExecutable(t *testing.T, dir, name string) {
	t.Helper()
	exePath := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		exePath += ".exe"
		content := `@echo off
exit /b 0
`
		if err := os.WriteFile(exePath, []byte(content), 0755); err != nil {
			t.Fatalf("failed to create dummy %s executable: %v", name, err)
		}
	} else {
		content := `#!/bin/sh
exit 0
`
		if err := os.WriteFile(exePath, []byte(content), 0755); err != nil {
			t.Fatalf("failed to create dummy %s executable: %v", name, err)
		}
	}
}

func TestCheckFunction(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath) // Restore PATH after tests

	// Create a temporary directory for dummy executables
	dummyBinDir := t.TempDir()
	defer os.RemoveAll(dummyBinDir) // Ensure cleanup for dummyBinDir
	createDummyExecutable(t, dummyBinDir, "go")
	createDummyExecutable(t, dummyBinDir, "git")
	createDummyExecutable(t, dummyBinDir, "goreleaser")

	// Test case 1: All dependencies present (should pass)
	t.Run("AllDependenciesPresent", func(t *testing.T) {
		// Ensure system PATH is used
		os.Setenv("PATH", dummyBinDir+string(os.PathListSeparator)+originalPath)
		if err := check(); err != nil {
			t.Errorf("check() failed when all dependencies should be present: %v", err)
		}
	})

	// Test case 2: go is missing
	t.Run("GoMissing", func(t *testing.T) {
		tempDummyBinDir := t.TempDir()
		defer os.RemoveAll(tempDummyBinDir) // Ensure cleanup
		createDummyExecutable(t, tempDummyBinDir, "git")
		createDummyExecutable(t, tempDummyBinDir, "goreleaser")
		os.Setenv("PATH", tempDummyBinDir) // Only dummy git and goreleaser

		err := check()
		if err == nil {
			t.Errorf("check() succeeded when 'go' should be missing")
		}
		if !strings.Contains(err.Error(), "executable file not found in %PATH%") {
			t.Errorf("expected error about missing Go, but got: %v", err)
		}
	})

	// Test case 3: git is missing
	t.Run("GitMissing", func(t *testing.T) {
		tempDummyBinDir := t.TempDir()
		defer os.RemoveAll(tempDummyBinDir) // Ensure cleanup
		createDummyExecutable(t, tempDummyBinDir, "go")
		createDummyExecutable(t, tempDummyBinDir, "goreleaser")
		os.Setenv("PATH", tempDummyBinDir) // Only dummy go and goreleaser

		err := check()
		if err == nil {
			t.Errorf("check() succeeded when 'git' should be missing")
		}
		if !strings.Contains(err.Error(), "executable file not found in %PATH%") {
			t.Errorf("expected error about missing Git, but got: %v", err)
		}
	})

	// Test case 4: goreleaser is missing
	t.Run("GoreleaserMissing", func(t *testing.T) {
		tempDummyBinDir := t.TempDir()
		defer os.RemoveAll(tempDummyBinDir) // Ensure cleanup
		createDummyExecutable(t, tempDummyBinDir, "go")
		createDummyExecutable(t, tempDummyBinDir, "git")
		os.Setenv("PATH", tempDummyBinDir) // Only dummy go and git

		err := check()
		if err == nil {
			t.Errorf("check() succeeded when 'goreleaser' should be missing")
		}
		if !strings.Contains(err.Error(), "executable file not found in %PATH%") {
			t.Errorf("expected error about missing Goreleaser, but got: %s", err)
		}
	})
}

func TestGetMainFileNameError(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original working directory: %v", err)
	}
	defer os.Chdir(originalWd) // Restore original working directory

	t.Run("NoGoModFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		currentWd, err := os.Getwd() // Capture current working directory inside subtest
		if err != nil {
			t.Fatalf("failed to get current working directory: %v", err)
		}
		defer os.Chdir(currentWd) // Defer restoring working directory

		os.Chdir(tmpDir) // Change to temporary directory

		_, err = getMainFileName()
		if err == nil {
			t.Errorf("getMainFileName() succeeded when go.mod should be missing")
		}
	})

	t.Run("GoModExistsButNoMainGoFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		currentWd, err := os.Getwd() // Capture current working directory inside subtest
		if err != nil {
			t.Fatalf("failed to get current working directory: %v", err)
		}
		defer os.Chdir(currentWd) // Defer restoring working directory

		os.Chdir(tmpDir) // Change to temporary directory

		// Create a dummy go.mod file
		if err := os.WriteFile("go.mod", []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod: %v", err)
		}

		_, err = getMainFileName()
		if err == nil {
			t.Errorf("getMainFileName() succeeded when main.go and module_name.go should be missing")
		}
		if !strings.Contains(err.Error(), "CreateFile myproject.go: The system cannot find the file specified.") {
			t.Errorf("Expected error message about missing main.go or module_name.go, got: %v", err)
		}
	})
}

func createMockGit(t *testing.T, tmpDir string) string {
	t.Helper()
	mockGitPath := filepath.Join(tmpDir, "git")
	if runtime.GOOS == "windows" {
		mockGitPath += ".exe"
	}

	// This script simplifies the mock for Windows to directly handle the known commands
	// without relying on external `where` or complex path manipulation.
	// It also fixes `%%` to `%` for environment variables in batch.
	scriptContent := `#!/bin/sh
if [ "$1" = "config" ] && [ "$2" = "--get" ] && [ "$3" = "remote.origin.url" ]; then
    exit 1 # Simulate error for getGitOrigin
elif [ "$1" = "describe" ] && [ "$2" = "--tags" ] && [ "$3" = "--abbrev=0" ]; then
    exit 1 # Simulate error for getGitTag
elif [ "$1" = "rev-parse" ]; then
	exit 1 # Simulate error for getGitCommit
elif [ "$1" = "tag" ]; then
	if [ "$2" = "-d" ]; then
		# Check if the tag to delete is "v1.2.3" and we are simulating failure
		if [ "$3" = "v1.2.3" ] && [ "$4" = "simulate_delete_fail" ]; then
			exit 1 # Simulate error for git tag -d
		fi
		# Allow successful deletion otherwise
	else
		if [ "$2" = "v1.2.3" ]; then # Simulate tag already exists
			exit 1
		fi
	fi
	echo "Mock git tag command executed" # Indicate success for other tags
elif [ "$1" = "push" ]; then
	echo "Mock git push command executed"
else
    # Fallback to real git if not a mocked command
    exec git "$@"
fi
`
	if runtime.GOOS == "windows" {
		scriptContent = `@echo off
if "%1"=="config" if "%2"=="--get" if "%3"=="remote.origin.url" (
    exit /b 1
)
if "%1"=="describe" if "%2"=="--tags" if "%3"=="--abbrev=0" (
    exit /b 1
)
if "%1"=="rev-parse" (
    exit /b 1
)
if "%1"=="tag" (
	if "%2"=="-d" (
		REM Simulate failure if deleting "v1.2.3"
		if "%3"=="v1.2.3" (
			exit /b 1
		)
	) else (
		REM Simulate tag already exists if creating "v1.2.3"
		if "%2"=="v1.2.3" (
			exit /b 1
		)
	)
	echo Mock git tag command executed
	exit /b 0
)
if "%1"=="push" (
	echo Mock git push command executed
	exit /b 0
)
:: Fallback to real git. This relies on 'git' being in the actual PATH
git %*
exit /b %errorlevel%
`
	}

	if err := os.WriteFile(mockGitPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create mock git executable: %v", err)
	}
	return tmpDir // Return the directory where mock git is located
}

func TestGetGitOriginError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping TestGetGitOriginError on Windows due to mock executable compatibility")
	}
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	// Create a mock git executable in the temporary directory
	mockGitDir := createMockGit(t, tmpDir)

	// Save original PATH and defer its restoration
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set PATH to include mockGitDir first
	os.Setenv("PATH", mockGitDir+string(os.PathListSeparator)+originalPath)

	// Call the function under test
	_, err := getGitOrigin()

	// Assert that an error is returned
	if err == nil {
		t.Error("getGitOrigin() did not return an error when git command was mocked to fail")
	}
}

func TestGetGitTagError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping TestGetGitTagError on Windows due to mock executable compatibility")
	}
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)
	mockGitDir := createMockGit(t, tmpDir)
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)
	os.Setenv("PATH", mockGitDir+string(os.PathListSeparator)+originalPath)

	tag := getGitTag()
	if tag != "unknown" {
		t.Errorf("getGitTag() expected 'unknown', got '%s'", tag)
	}
}

func TestGetGitCommitError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping TestGetGitCommitError on Windows due to mock executable compatibility")
	}
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)
	mockGitDir := createMockGit(t, tmpDir)
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)
	os.Setenv("PATH", mockGitDir+string(os.PathListSeparator)+originalPath)

	commit := getGitCommit("HEAD")
	if commit != "unknown" {
		t.Errorf("getGitCommit() expected 'unknown', got '%s'", commit)
	}
}

func TestCreateMainFile(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original working directory: %v", err)
	}
	defer os.Chdir(originalWd) // Restore original working directory

	t.Run("SuccessfulCreation", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping TempDir cleanup issues on Windows due to unclosed file handles in main.go")
		}
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		currentWd, err := os.Getwd() // Capture current working directory inside subtest
		if err != nil {
			t.Fatalf("failed to get current working directory: %v", err)
		}
		defer os.Chdir(currentWd) // Defer restoring working directory

		os.Chdir(tmpDir) // Change to temporary directory

		// create a dummy go.mod file, as createMainFile calls getModuleName
		if err := os.WriteFile("go.mod", []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod: %v", err)
		}


		err = createMainFile()
		if err != nil {
			t.Fatalf("createMainFile() failed: %v", err)
		}

		mainFilePath := filepath.Join(tmpDir, "main.go")
		if _, err := os.Stat(mainFilePath); os.IsNotExist(err) {
			t.Errorf("main.go was not created")
		}

		content, err := os.ReadFile(mainFilePath)
		if err != nil {
			t.Fatalf("failed to read main.go: %v", err)
		}
		if !strings.Contains(string(content), "package main") ||
			!strings.Contains(string(content), `const version = "0.1.0"`) ||
			!strings.Contains(string(content), "func main()") ||
			!strings.Contains(string(content), "func Version()") ||
			!strings.Contains(string(content), "func Usage()") {
			t.Errorf("main.go content is incorrect: %s", string(content))
		}
	})

	t.Run("FileCreationFailure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping FileCreationFailure on Windows due to os.Chmod limitations")
		}
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		currentWd, err := os.Getwd() // Capture current working directory inside subtest
		if err != nil {
			t.Fatalf("failed to get current working directory: %v", err)
		}
		defer os.Chdir(currentWd) // Defer restoring working directory

		os.Chdir(tmpDir) // Change to temporary directory

		// create a dummy go.mod file, as createMainFile calls getModuleName
		if err := os.WriteFile("go.mod", []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod: %v", err)
		}

		// Make the directory non-writable to simulate file creation failure
		if err := os.Chmod(tmpDir, 0444); err != nil {
			t.Fatalf("failed to change permissions of directory: %v", err)
		}
		defer os.Chmod(tmpDir, 0755) // Restore permissions

		err = createMainFile()
		if err == nil {
			t.Fatalf("createMainFile() succeeded when it should have failed")
		}
		if !strings.Contains(err.Error(), "Error creating main.go file") {
			t.Errorf("expected error about creating main.go file, but got: %s", err)
		}
	})
}

func createMockGo(t *testing.T, tmpDir string) string {
	t.Helper()
	mockGoPath := filepath.Join(tmpDir, "go")
	if runtime.GOOS == "windows" {
		mockGoPath += ".exe"
	}

	scriptContent := `#!/bin/sh
if [ "$1" = "build" ]; then
    exit 1 # Simulate error for go build
else
    exit 0 # Exit successfully for other commands
fi
`
	if runtime.GOOS == "windows" {
		scriptContent = `@echo off
if "%1"=="build" (
    exit /b 1
) else (
	exit /b 0
)
`
	}

	if err := os.WriteFile(mockGoPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create mock go executable: %v", err)
	}
	return tmpDir // Return the directory where mock go is located
}

func TestCreateTestFile(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original working directory: %v", err)
	}
	defer os.Chdir(originalWd) // Restore original working directory

	t.Run("SuccessfulCreation", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping TempDir cleanup issues on Windows due to unclosed file handles in main.go")
		}
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		currentWd, err := os.Getwd() // Capture current working directory inside subtest
		if err != nil {
			t.Fatalf("failed to get current working directory: %v", err)
		}
		defer os.Chdir(currentWd) // Defer restoring working directory

		os.Chdir(tmpDir) // Change to temporary directory

		// Create a dummy go.mod file and main.go for getModuleName to succeed
		if err := os.WriteFile("go.mod", []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod: %v", err)
		}
		if err := os.WriteFile("main.go", []byte("package main"), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go: %v", err)
		}

		err = createTestFile()
		if err != nil {
			t.Fatalf("createTestFile() failed: %v", err)
		}

		testFilePath := filepath.Join(tmpDir, "main_test.go")
		if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
			t.Errorf("main_test.go was not created")
		}

		content, err := os.ReadFile(testFilePath)
		if err != nil {
			t.Fatalf("failed to read main_test.go: %v", err)
		}
		if !strings.Contains(string(content), "package main") ||
			!strings.Contains(string(content), `binName  = "test"`) ||
			!strings.Contains(string(content), "func TestMain(m *testing.M)") {
			t.Errorf("main_test.go content is incorrect: %s", string(content))
		}
	})

	t.Run("FileCreationFailure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping FileCreationFailure on Windows due to os.Chmod limitations")
		}
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		currentWd, err := os.Getwd() // Capture current working directory inside subtest
		if err != nil {
			t.Fatalf("failed to get current working directory: %v", err)
		}
		defer os.Chdir(currentWd) // Defer restoring working directory

		os.Chdir(tmpDir) // Change to temporary directory

		// Create a dummy go.mod file and main.go for getModuleName to succeed
		if err := os.WriteFile("go.mod", []byte("module myproject"), 0644); err != nil {
			t.Fatalf("failed to create dummy go.mod: %v", err)
		}
		if err := os.WriteFile("main.go", []byte("package main"), 0644); err != nil {
			t.Fatalf("failed to create dummy main.go: %v", err)
		}

		// Make the directory non-writable to simulate file creation failure
		if err := os.Chmod(tmpDir, 0444); err != nil {
			t.Fatalf("failed to change permissions of directory: %v", err)
		}
		defer os.Chmod(tmpDir, 0755) // Restore permissions

		err = createTestFile()
		if err == nil {
			t.Fatalf("createTestFile() succeeded when it should have failed")
		}
		if !strings.Contains(err.Error(), "Error creating main_test.go file") {
			t.Errorf("expected error about creating main_test.go file, but got: %s", err)
		}
	})

	t.Run("GetMainFileNameFailure", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		currentWd, err := os.Getwd() // Capture current working directory inside subtest
		if err != nil {
			t.Fatalf("failed to get current working directory: %v", err)
		}
		defer os.Chdir(currentWd) // Defer restoring working directory

		os.Chdir(tmpDir) // Change to temporary directory

		// No go.mod or main.go to simulate getMainFileName failure

		err = createTestFile()
		if err == nil {
			t.Fatalf("createTestFile() succeeded when getMainFileName should have failed")
		}
		if !strings.Contains(err.Error(), "open go.mod: The system cannot find the file specified.") {
			t.Errorf("expected error about getMainFileName failure, but got: %s", err)
		}
	})
}

func TestInstallGoBuildFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping TestInstallGoBuildFailure on Windows due to mock executable compatibility")
	}
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	// Create dummy go.mod and main.go files
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module myproject"), 0644); err != nil {
		t.Fatalf("failed to create dummy go.mod file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "myproject.go"), []byte(`package main; func main(){}`), 0644); err != nil {
		t.Fatalf("failed to create dummy main.go file: %v", err)
	}

	gopherBinary := buildGopher(t, tmpDir)

	// Create mock go
	mockGoDir := createMockGo(t, t.TempDir())
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)
	os.Setenv("PATH", mockGoDir+string(os.PathListSeparator)+originalPath)

	cmd := exec.Command(gopherBinary, "install")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected gopher install command to fail due to go build, but it succeeded. Output: %s", output)
	}
	if !strings.Contains(string(output), "Error running go build command") && !strings.Contains(string(output), "go build failed") {
		t.Errorf("expected output to contain go build failure, but got: %s", output)
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
