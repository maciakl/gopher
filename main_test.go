package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetName(t *testing.T) {
	testCases := []struct {
		uri      string
		expected string
	}{
		{"github.com/user/repo", "repo"},
		{"user/repo", "repo"},
		{"repo", "repo"},
	}

	for _, tc := range testCases {
		t.Run(tc.uri, func(t *testing.T) {
			if got := getName(tc.uri); got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestGetUsername(t *testing.T) {
	testCases := []struct {
		uri      string
		expected string
	}{
		{"github.com/user/repo", "user"},
		{"user/repo", "user"},
	}

	for _, tc := range testCases {
		t.Run(tc.uri, func(t *testing.T) {
			if got := getUsername(tc.uri); got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestIncString(t *testing.T) {
	testCases := []struct {
		s        string
		expected string
	}{
		{"1", "2"},
		{"9", "10"},
		{"0", "1"},
	}

	for _, tc := range testCases {
		t.Run(tc.s, func(t *testing.T) {
			if got := incString(tc.s); got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "main.go")
	content := `package main
const version = "1.2.3"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	version, err := getVersion(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "1.2.3" {
		t.Errorf("expected version %q, got %q", "1.2.3", version)
	}
}

func TestGetModule(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "go.mod")
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	content := `module github.com/user/repo`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	module, err := getModule()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if module != "github.com/user/repo" {
		t.Errorf("expected module %q, got %q", "github.com/user/repo", module)
	}
}

func TestGetModuleName(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "go.mod")
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)
	content := `module github.com/user/repo`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	moduleName, err := getModuleName()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if moduleName != "repo" {
		t.Errorf("expected module name %q, got %q", "repo", moduleName)
	}
}

func TestFindInFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world\nfind me here\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	line, err := findInFile(tmpFile, "find me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "find me here" {
		t.Errorf("expected to find %q, got %q", "find me here", line)
	}

	line, err = findInFile(tmpFile, "not here")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "" {
		t.Errorf("expected to find nothing, got %q", line)
	}
}

func TestReplaceInFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := replaceInFile(tmpFile, "world", "gopher")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	newContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(newContent) != "hello gopher" {
		t.Errorf("expected content %q, got %q", "hello gopher", string(newContent))
	}
}

func TestVersionBump(t *testing.T) {
	tmpDir := t.TempDir()
	mainGo := filepath.Join(tmpDir, "main.go")
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module gopher"), 0644); err != nil {
		t.Fatal(err)
	}
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	content := `package main
const version = "1.2.3"
`
	if err := os.WriteFile(mainGo, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		bumpType string
		expected string
	}{
		{"patch", "1.2.4"},
		{"minor", "1.3.0"},
		{"major", "2.0.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.bumpType, func(t *testing.T) {
			// Reset file content for each test case
			content := `package main
const version = "1.2.3"
`
			if err := os.WriteFile(mainGo, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
			err := versionBump(tc.bumpType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			version, err := getVersion(mainGo)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != tc.expected {
				t.Errorf("expected version %q, got %q", tc.expected, version)
			}
		})
	}
}

func TestCreateMakefile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// create a dummy go.mod
	if err := os.WriteFile("go.mod", []byte("module myproject"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := createMakefile(); err != nil {
		t.Fatalf("createMakefile() failed: %v", err)
	}

	if _, err := os.Stat("Makefile"); os.IsNotExist(err) {
		t.Error("Makefile was not created")
	}
}

func TestCreateJustfile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// create a dummy go.mod
	if err := os.WriteFile("go.mod", []byte("module myproject"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := createJustfile(); err != nil {
		t.Fatalf("createJustfile() failed: %v", err)
	}

	if _, err := os.Stat("Justfile"); os.IsNotExist(err) {
		t.Error("Justfile was not created")
	}
}

func TestCreateMainFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	if err := createMainFile(); err != nil {
		t.Fatalf("createMainFile() failed: %v", err)
	}

	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		t.Error("main.go was not created")
	}
}

func TestCreateTestFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// create a dummy main.go so getMainFileName() works
	if err := os.WriteFile("main.go", []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := createTestFile(); err != nil {
		t.Fatalf("createTestFile() failed: %v", err)
	}

	if _, err := os.Stat("main_test.go"); os.IsNotExist(err) {
		t.Error("main_test.go was not created")
	}
}
