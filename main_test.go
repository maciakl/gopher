package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("file-not-found", func(t *testing.T) {
		_, err := getVersion("non-existent-file.go")
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestGetModule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("file-not-found", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		_, err := getModule()
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestGetModuleName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("file-not-found", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		_, err := getModuleName()
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestFindInFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("file-not-found", func(t *testing.T) {
		_, err := findInFile("non-existent-file.txt", "hello")
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestReplaceInFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("file-not-found", func(t *testing.T) {
		err := replaceInFile("non-existent-file.txt", "a", "b")
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("read-permission-denied", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "read_only.txt")
		content := "some content"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		// Change permissions to be non-readable (write-only for owner, no access for others)
		if err := os.Chmod(tmpFile, 0000); err != nil { // 0000 means no permissions for anyone
			// On Windows, changing permissions to 0000 might not fully prevent read,
			// but will likely cause an access denied error.
			// This test might behave differently on Windows vs Unix-like systems.
			t.Fatalf("failed to change file permissions: %v", err)
		}

		err := replaceInFile(tmpFile, "content", "changed")
		if err == nil {
			t.Error("expected an error due to read permission denied, got nil")
		}
		// Restore permissions for cleanup, though TempDir cleans up anyway
		os.Chmod(tmpFile, 0644)
	})

	t.Run("write-permission-denied", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "no_write.txt")
		content := "some content"
		if err := os.WriteFile(tmpFile, []byte(content), 0444); err != nil { // 0444 means read-only for all
			t.Fatal(err)
		}

		err := replaceInFile(tmpFile, "content", "changed")
		if err == nil {
			t.Error("expected an error due to write permission denied, got nil")
		}
		// Restore permissions for cleanup
		os.Chmod(tmpFile, 0644)
	})

}

func TestVersionBump(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("invalid-bump-type", func(t *testing.T) {
		err := versionBump("invalid")
		if err == nil {
			t.Error("expected an error for invalid bump type, got nil")
		}
	})
}

func TestCreateMakefile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("no-go-mod", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		err := createMakefile()
		if err == nil {
			t.Error("expected an error when go.mod is missing, got nil")
		}
	})
}

func TestCreateJustfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("no-go-mod", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		err := createJustfile()
		if err == nil {
			t.Error("expected an error when go.mod is missing, got nil")
		}
	})
}

func TestCreateMainFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("module-not-found", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		// createMainFile internally calls getModuleName, which will fail if go.mod is not present.
		// However, createMainFile sets a default name "main" if getModuleName returns an error.
		// So we expect success here.
		if err := createMainFile(); err != nil {
			t.Fatalf("createMainFile() failed: %v", err)
		}
		if _, err := os.Stat("main.go"); os.IsNotExist(err) {
			t.Error("main.go was not created")
		}
	})
}

func TestCreateTestFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("no-main-file", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		err := createTestFile()
		if err == nil {
			t.Error("expected an error when main file is missing, got nil")
		}
	})
}

func TestGetMainFileName(t *testing.T) {
	t.Run("main-exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)
		if err := os.WriteFile("main.go", []byte("package main"), 0644); err != nil {
			t.Fatal(err)
		}
		name, err := getMainFileName()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "main" {
			t.Errorf("expected 'main', got %q", name)
		}
	})

	t.Run("module-file-exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)
		if err := os.WriteFile("go.mod", []byte("module myproject"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("myproject.go", []byte("package main"), 0644); err != nil {
			t.Fatal(err)
		}
		name, err := getMainFileName()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "myproject" {
			t.Errorf("expected 'myproject', got %q", name)
		}
	})

	t.Run("no-main-or-module-file", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)
		_, err := getMainFileName()
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func TestCheck(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	t.Run("all-present", func(t *testing.T) {
		// Assume go, git, goreleaser are present in the original PATH
		// If not, this test will fail, indicating a problem with the test setup
		err := check()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("go-missing", func(t *testing.T) {
		t.Setenv("PATH", "") // Clear PATH to simulate missing executables
		err := check()
		if err == nil {
			t.Error("expected an error, got nil")
		}
		// Check for the actual error message from exec.LookPath
		if err != nil && !strings.Contains(err.Error(), "executable file not found in %PATH%") {
			t.Errorf("expected 'executable file not found in %%PATH%%' error, got %v", err)
		}
	})

	t.Run("git-missing", func(t *testing.T) {
		// Create a temporary directory and put a dummy 'go' executable in it
		tmpBinDir := t.TempDir()
		if runtime.GOOS == "windows" {
			// On Windows, executables usually need a .exe extension
			dummyGoPath := filepath.Join(tmpBinDir, "go.exe")
			if err := os.WriteFile(dummyGoPath, []byte("echo go"), 0755); err != nil {
				t.Fatal(err)
			}
		} else {
			dummyGoPath := filepath.Join(tmpBinDir, "go")
			if err := os.WriteFile(dummyGoPath, []byte("#!/bin/sh\necho go"), 0755); err != nil {
				t.Fatal(err)
			}
		}

		t.Setenv("PATH", tmpBinDir) // Set PATH to only contain the dummy go

		err := check()
		if err == nil {
			t.Error("expected an error, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "executable file not found in %PATH%") {
			t.Errorf("expected 'executable file not found in %%PATH%%' error, got %v", err)
		}
	})

	t.Run("goreleaser-missing", func(t *testing.T) {
		// Create a temporary directory and put dummy 'go' and 'git' executables in it
		tmpBinDir := t.TempDir()

		if runtime.GOOS == "windows" {
			dummyGoPath := filepath.Join(tmpBinDir, "go.exe")
			if err := os.WriteFile(dummyGoPath, []byte("echo go"), 0755); err != nil {
				t.Fatal(err)
			}
			dummyGitPath := filepath.Join(tmpBinDir, "git.exe")
			if err := os.WriteFile(dummyGitPath, []byte("echo git"), 0755); err != nil {
				t.Fatal(err)
			}
		} else {
			dummyGoPath := filepath.Join(tmpBinDir, "go")
			if err := os.WriteFile(dummyGoPath, []byte("#!/bin/sh\necho go"), 0755); err != nil {
				t.Fatal(err)
			}
			dummyGitPath := filepath.Join(tmpBinDir, "git")
			if err := os.WriteFile(dummyGitPath, []byte("#!/bin/sh\necho git"), 0755); err != nil {
				t.Fatal(err)
			}
		}

		t.Setenv("PATH", tmpBinDir) // Set PATH to contain dummy go and git

		err := check()
		if err == nil {
			t.Error("expected an error, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "executable file not found in %PATH%") {
			t.Errorf("expected 'executable file not found in %%PATH%%' error, got %v", err)
		}
	})
}
