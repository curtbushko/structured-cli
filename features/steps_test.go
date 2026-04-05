package features

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"
	_ "modernc.org/sqlite" // SQLite driver for tracking database checks
)

// binaryPath holds the path to the built binary, set once at test start
var (
	builtBinaryPath string
	buildOnce       sync.Once
	buildErr        error
)

// buildBinary builds the structured-cli binary once for all tests
func buildBinary() (string, error) {
	buildOnce.Do(func() {
		cwd, err := os.Getwd()
		if err != nil {
			buildErr = err
			return
		}
		projectRoot := filepath.Dir(cwd)
		if projectRoot == "" || projectRoot == "." {
			projectRoot = cwd
		}

		builtBinaryPath = filepath.Join(projectRoot, "bin", "structured-cli")
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", builtBinaryPath, "./cmd/structured-cli")
		cmd.Dir = projectRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("failed to build structured-cli: %w\n%s", err, out)
			return
		}
	})
	return builtBinaryPath, buildErr
}

// testContext holds state across scenario steps
type testContext struct {
	tempDir        string
	repoDir        string
	output         string
	exitCode       int
	envVars        map[string]string
	binaryPath     string
	emptyRepoDir   string
	testFile       string
	nonGitDir      string
	trackingDBPath string // Path to the tracking database for E2E tests
	trackingDir    string // Temp directory for XDG_DATA_HOME
}

func newTestContext() *testContext {
	return &testContext{
		envVars: make(map[string]string),
	}
}

// Step definitions

func (tc *testContext) iHaveAGitRepository(ctx context.Context) error {
	// Get the pre-built binary path
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	// Create an isolated test repository
	tc.tempDir, err = os.MkdirTemp("", "structured-cli-test-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.tempDir

	// Initialize git repo
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = tc.repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %w\n%s", err, out)
	}

	// Configure git user
	cmd = exec.CommandContext(ctx, "git", "config", "user.email", "test@test.com")
	cmd.Dir = tc.repoDir
	_ = cmd.Run()
	cmd = exec.CommandContext(ctx, "git", "config", "user.name", "Test User")
	cmd.Dir = tc.repoDir
	_ = cmd.Run()

	// Create initial commit so we have a valid repo
	readmeFile := filepath.Join(tc.repoDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test Repository\n"), 0o644); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "add", "README.md")
	cmd.Dir = tc.repoDir
	_ = cmd.Run()

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", "initial commit")
	cmd.Dir = tc.repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w\n%s", err, out)
	}

	return nil
}

func (tc *testContext) iHaveAGitRepositoryWithCommits(ctx context.Context) error {
	// First create a basic repository
	if err := tc.iHaveAGitRepository(ctx); err != nil {
		return err
	}

	// Add more commits
	for i := 2; i <= 5; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		filePath := filepath.Join(tc.repoDir, filename)
		content := fmt.Sprintf("Content of file %d\n", i)

		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write file %d: %w", i, err)
		}

		cmd := exec.CommandContext(ctx, "git", "add", filename)
		cmd.Dir = tc.repoDir
		_ = cmd.Run()

		cmd = exec.CommandContext(ctx, "git", "commit", "-m", fmt.Sprintf("Add file %d", i))
		cmd.Dir = tc.repoDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git commit %d failed: %w\n%s", i, err, out)
		}
	}

	return nil
}

func (tc *testContext) theRepositoryHasNoChanges(_ context.Context) error {
	// The test repo is already clean after initial commit
	// Just verify there are no uncommitted changes
	return nil
}

func (tc *testContext) iCreateAnUntrackedFile(_ context.Context, filename string) error {
	filePath := filepath.Join(tc.repoDir, filename)
	return os.WriteFile(filePath, []byte("test content"), 0o644)
}

func (tc *testContext) iModifyATrackedFile(_ context.Context) error {
	// Modify the README.md that was created in the test repo
	readmeFile := filepath.Join(tc.repoDir, "README.md")
	content, err := os.ReadFile(readmeFile)
	if err != nil {
		return fmt.Errorf("read README.md: %w", err)
	}

	// Append content to make it modified
	newContent := append(content, []byte("\nModified line\n")...)
	if err := os.WriteFile(readmeFile, newContent, 0o644); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}

	return nil
}

func (tc *testContext) iHaveAnEmptyGitRepository(ctx context.Context) error {
	var err error
	tc.emptyRepoDir, err = os.MkdirTemp("", "structured-cli-empty-*")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = tc.emptyRepoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %w\n%s", err, out)
	}

	tc.repoDir = tc.emptyRepoDir
	return nil
}

func (tc *testContext) iAmNotInAGitRepository(_ context.Context) error {
	var err error
	tc.nonGitDir, err = os.MkdirTemp("", "structured-cli-nongit-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.nonGitDir

	// Get the pre-built binary path
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath
	return nil
}

func (tc *testContext) iHaveADirectoryWithFiles(ctx context.Context) error {
	// Get the pre-built binary path
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	tc.tempDir, err = os.MkdirTemp("", "structured-cli-files-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.tempDir

	// Create some test files
	if err := os.WriteFile(filepath.Join(tc.tempDir, "file1.txt"), []byte("content 1"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tc.tempDir, "file2.txt"), []byte("content 2"), 0o644); err != nil {
		return err
	}
	return os.Mkdir(filepath.Join(tc.tempDir, "subdir"), 0o755)
}

func (tc *testContext) iHaveAFileWithContent(_ context.Context, filename, content string) error {
	// Get the pre-built binary path
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	if tc.tempDir == "" {
		tc.tempDir, err = os.MkdirTemp("", "structured-cli-file-*")
		if err != nil {
			return err
		}
		tc.repoDir = tc.tempDir
	}

	tc.testFile = filepath.Join(tc.repoDir, filename)
	return os.WriteFile(tc.testFile, []byte(content), 0o644)
}

func (tc *testContext) iHaveAFileWithMultipleLines(_ context.Context, filename string) error {
	// Get the pre-built binary path
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	if tc.tempDir == "" {
		tc.tempDir, err = os.MkdirTemp("", "structured-cli-multiline-*")
		if err != nil {
			return err
		}
		tc.repoDir = tc.tempDir
	}

	// Create file with 20 lines
	var content strings.Builder
	for i := 1; i <= 20; i++ {
		content.WriteString(fmt.Sprintf("Line %d of the test file\n", i))
	}

	tc.testFile = filepath.Join(tc.repoDir, filename)
	return os.WriteFile(tc.testFile, []byte(content.String()), 0o644)
}

func (tc *testContext) iStageTheModifiedFile(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = tc.repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, out)
	}
	return nil
}

func (tc *testContext) iHaveAMakefileWithTarget(_ context.Context, target string) error {
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	tc.tempDir, err = os.MkdirTemp("", "structured-cli-make-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.tempDir

	makefileContent := fmt.Sprintf(".PHONY: %s\n%s:\n\t@echo \"Running %s\"\n", target, target, target)
	return os.WriteFile(filepath.Join(tc.tempDir, "Makefile"), []byte(makefileContent), 0o644)
}

func (tc *testContext) iHaveAMakefileWithFailingTarget(_ context.Context, target string) error {
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	var err2 error
	tc.tempDir, err2 = os.MkdirTemp("", "structured-cli-make-fail-*")
	if err2 != nil {
		return err2
	}
	tc.repoDir = tc.tempDir

	makefileContent := fmt.Sprintf(".PHONY: %s\n%s:\n\t@exit 1\n", target, target)
	return os.WriteFile(filepath.Join(tc.tempDir, "Makefile"), []byte(makefileContent), 0o644)
}

func (tc *testContext) iHaveAJustfileWithRecipe(_ context.Context, recipe string) error {
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	tc.tempDir, err = os.MkdirTemp("", "structured-cli-just-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.tempDir

	justfileContent := fmt.Sprintf("%s:\n    @echo \"Running %s\"\n", recipe, recipe)
	return os.WriteFile(filepath.Join(tc.tempDir, "justfile"), []byte(justfileContent), 0o644)
}

func (tc *testContext) iHaveAJustfileWithFailingRecipe(_ context.Context, recipe string) error {
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	var err2 error
	tc.tempDir, err2 = os.MkdirTemp("", "structured-cli-just-fail-*")
	if err2 != nil {
		return err2
	}
	tc.repoDir = tc.tempDir

	justfileContent := recipe + ":\n    @exit 1\n"
	return os.WriteFile(filepath.Join(tc.tempDir, "justfile"), []byte(justfileContent), 0o644)
}

func (tc *testContext) iHaveAGoProject(_ context.Context) error {
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	tc.tempDir, err = os.MkdirTemp("", "structured-cli-go-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.tempDir

	// Create go.mod
	goModContent := `module testproject

go 1.21
`
	if err := os.WriteFile(filepath.Join(tc.tempDir, "go.mod"), []byte(goModContent), 0o644); err != nil {
		return err
	}

	// Create a valid main.go
	mainGoContent := `package main

func main() {
	println("hello")
}
`
	return os.WriteFile(filepath.Join(tc.tempDir, "main.go"), []byte(mainGoContent), 0o644)
}

func (tc *testContext) theGoProjectHasASyntaxError(_ context.Context) error {
	// Overwrite main.go with syntax error
	brokenMainContent := `package main

func main() {
	println("missing quote)
}
`
	return os.WriteFile(filepath.Join(tc.repoDir, "main.go"), []byte(brokenMainContent), 0o644)
}

func (tc *testContext) theGoProjectHasTests(_ context.Context) error {
	// Create a simple test file
	testContent := `package main

import "testing"

func TestAdd(t *testing.T) {
	result := 1 + 1
	if result != 2 {
		t.Errorf("expected 2, got %d", result)
	}
}
`
	return os.WriteFile(filepath.Join(tc.repoDir, "main_test.go"), []byte(testContent), 0o644)
}

func (tc *testContext) theGoProjectHasVetIssues(_ context.Context) error {
	// Create code with a vet issue (unreachable code)
	vetIssueContent := `package main

import "fmt"

func main() {
	fmt.Printf("%s", 123) // Printf format mismatch: %s expects string, got int
}
`
	return os.WriteFile(filepath.Join(tc.repoDir, "main.go"), []byte(vetIssueContent), 0o644)
}

func (tc *testContext) theGoProjectHasFormatIssues(_ context.Context) error {
	// Create code with formatting issues (wrong indentation)
	badFormatContent := `package main

func main() {
println("hello")
}
`
	return os.WriteFile(filepath.Join(tc.repoDir, "main.go"), []byte(badFormatContent), 0o644)
}

// Docker step definitions

func (tc *testContext) dockerIsAvailable(ctx context.Context) error {
	// Get the pre-built binary path
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	// Create a temp directory for running commands
	tc.tempDir, err = os.MkdirTemp("", "structured-cli-docker-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.tempDir

	// Check if docker is available and daemon is running
	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		return godog.ErrPending // Skip test if docker not available
	}

	return nil
}

// Node.js project step definitions

func (tc *testContext) iHaveANodejsProject(_ context.Context) error {
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	tc.tempDir, err = os.MkdirTemp("", "structured-cli-npm-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.tempDir

	// Create package.json with a minimal valid structure
	packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "Test project for npm E2E tests",
  "main": "index.js",
  "dependencies": {}
}
`
	return os.WriteFile(filepath.Join(tc.tempDir, "package.json"), []byte(packageJSON), 0o644)
}

func (tc *testContext) iHaveANodejsProjectWithNoDependencies(_ context.Context) error {
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	tc.tempDir, err = os.MkdirTemp("", "structured-cli-npm-empty-*")
	if err != nil {
		return err
	}
	tc.repoDir = tc.tempDir

	// Create package.json with no dependencies
	packageJSON := `{
  "name": "empty-project",
  "version": "1.0.0",
  "description": "Empty test project"
}
`
	return os.WriteFile(filepath.Join(tc.tempDir, "package.json"), []byte(packageJSON), 0o644)
}

func (tc *testContext) theEnvironmentVariableIsSetTo(_ context.Context, name, value string) error {
	tc.envVars[name] = value
	return nil
}

func (tc *testContext) iRun(ctx context.Context, command string) error {
	// Parse the command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return errors.New("empty command")
	}

	// Replace "structured-cli" with the actual binary path
	if parts[0] == "structured-cli" {
		parts[0] = tc.binaryPath
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = tc.repoDir

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range tc.envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	tc.output = stdout.String() + stderr.String()

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		tc.exitCode = exitErr.ExitCode()
	} else if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	} else {
		tc.exitCode = 0
	}

	return nil
}

func (tc *testContext) theExitCodeShouldBe(_ context.Context, expected int) error {
	if tc.exitCode != expected {
		return fmt.Errorf("expected exit code %d, got %d\nOutput: %s", expected, tc.exitCode, tc.output)
	}
	return nil
}

func (tc *testContext) theOutputShouldBeValidJSON(_ context.Context) error {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(tc.output), &js); err != nil {
		return fmt.Errorf("output is not valid JSON: %w\nOutput: %s", err, tc.output)
	}
	return nil
}

func (tc *testContext) theOutputShouldNotBeJSON(_ context.Context) error {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(tc.output), &js); err == nil {
		return errors.New("expected non-JSON output, but got valid JSON")
	}
	return nil
}

func (tc *testContext) theOutputShouldContain(_ context.Context, expected string) error {
	if !strings.Contains(tc.output, expected) {
		return fmt.Errorf("output does not contain %q\nOutput: %s", expected, tc.output)
	}
	return nil
}

func (tc *testContext) theJSONShouldContainKeyWithBooleanValue(_ context.Context, key string, valueStr string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	val, ok := data[key]
	if !ok {
		return fmt.Errorf("JSON does not contain key %q\nJSON: %s", key, tc.output)
	}

	boolVal, ok := val.(bool)
	if !ok {
		return fmt.Errorf("key %q is not a boolean, got %T", key, val)
	}

	expected := valueStr == "true"
	if boolVal != expected {
		return fmt.Errorf("key %q has value %v, expected %v", key, boolVal, expected)
	}

	return nil
}

func (tc *testContext) theJSONShouldContainKeyAsAString(_ context.Context, key string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	val, ok := data[key]
	if !ok {
		return fmt.Errorf("JSON does not contain key %q\nJSON: %s", key, tc.output)
	}

	if _, ok := val.(string); !ok {
		return fmt.Errorf("key %q is not a string, got %T", key, val)
	}

	return nil
}

func (tc *testContext) theJSONStringShouldNotBeEmpty(_ context.Context, key string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	val, ok := data[key]
	if !ok {
		return fmt.Errorf("JSON does not contain key %q\nJSON: %s", key, tc.output)
	}

	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("key %q is not a string, got %T", key, val)
	}

	if str == "" {
		return fmt.Errorf("string %q is empty", key)
	}

	return nil
}

func (tc *testContext) theJSONShouldContainKeyAsAnArray(_ context.Context, key string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	val, ok := data[key]
	if !ok {
		return fmt.Errorf("JSON does not contain key %q\nJSON: %s", key, tc.output)
	}

	// Accept both array and null (null is a valid "empty" array in JSON from Go nil slices)
	if val == nil {
		return nil // null is acceptable as an empty array
	}
	if _, ok := val.([]interface{}); !ok {
		return fmt.Errorf("key %q is not an array, got %T", key, val)
	}

	return nil
}

func (tc *testContext) theJSONShouldContainKey(_ context.Context, key string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if _, ok := data[key]; !ok {
		return fmt.Errorf("JSON does not contain key %q\nJSON: %s", key, tc.output)
	}

	return nil
}

func (tc *testContext) theJSONArrayShouldContain(_ context.Context, key, expected string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	arr, ok := data[key].([]interface{})
	if !ok {
		return fmt.Errorf("key %q is not an array", key)
	}

	for _, item := range arr {
		if str, ok := item.(string); ok && str == expected {
			return nil
		}
	}

	return fmt.Errorf("array %q does not contain %q", key, expected)
}

func (tc *testContext) theJSONArrayShouldNotBeEmpty(_ context.Context, key string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	arr, ok := data[key].([]interface{})
	if !ok {
		return fmt.Errorf("key %q is not an array", key)
	}

	if len(arr) == 0 {
		return fmt.Errorf("array %q is empty", key)
	}

	return nil
}

func (tc *testContext) theJSONArrayShouldBeEmpty(_ context.Context, key string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	val, ok := data[key]
	if !ok {
		return fmt.Errorf("key %q not found in JSON", key)
	}

	// Accept null as empty
	if val == nil {
		return nil
	}

	arr, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("key %q is not an array", key)
	}

	if len(arr) != 0 {
		return fmt.Errorf("array %q is not empty, has %d items", key, len(arr))
	}

	return nil
}

func (tc *testContext) theJSONArrayShouldHaveAtMostNItems(_ context.Context, key string, maxItems int) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	arr, ok := data[key].([]interface{})
	if !ok {
		return fmt.Errorf("key %q is not an array", key)
	}

	if len(arr) > maxItems {
		return fmt.Errorf("array %q has %d items, expected at most %d", key, len(arr), maxItems)
	}

	return nil
}

func (tc *testContext) theFirstCommitShouldHaveAsAString(_ context.Context, field string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	commits, ok := data["commits"].([]interface{})
	if !ok || len(commits) == 0 {
		return errors.New("no commits found in JSON")
	}

	firstCommit, ok := commits[0].(map[string]interface{})
	if !ok {
		return errors.New("first commit is not an object")
	}

	val, ok := firstCommit[field]
	if !ok {
		return fmt.Errorf("first commit does not have field %q", field)
	}

	if _, ok := val.(string); !ok {
		return fmt.Errorf("field %q is not a string, got %T", field, val)
	}

	return nil
}

func (tc *testContext) theFirstFileShouldHaveAsAString(_ context.Context, field string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	files, ok := data["files"].([]interface{})
	if !ok || len(files) == 0 {
		return errors.New("no files found in JSON")
	}

	firstFile, ok := files[0].(map[string]interface{})
	if !ok {
		return errors.New("first file is not an object")
	}

	val, ok := firstFile[field]
	if !ok {
		return fmt.Errorf("first file does not have field %q", field)
	}

	if _, ok := val.(string); !ok {
		return fmt.Errorf("field %q is not a string, got %T", field, val)
	}

	return nil
}

func (tc *testContext) oneBranchShouldHaveCurrentEqualToTrue(_ context.Context) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	branches, ok := data["branches"].([]interface{})
	if !ok {
		return errors.New("branches is not an array")
	}

	for _, branch := range branches {
		branchMap, ok := branch.(map[string]interface{})
		if !ok {
			continue
		}
		if current, ok := branchMap["current"].(bool); ok && current {
			return nil
		}
	}

	return errors.New("no branch has current=true")
}

func (tc *testContext) theJSONArrayShouldHaveAtLeastNItems(_ context.Context, key string, minItems int) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	arr, ok := data[key].([]interface{})
	if !ok {
		return fmt.Errorf("key %q is not an array", key)
	}

	if len(arr) < minItems {
		return fmt.Errorf("array %q has %d items, expected at least %d", key, len(arr), minItems)
	}

	return nil
}

// Tracking-specific step definitions

func (tc *testContext) iHaveACleanTrackingDatabase(_ context.Context) error {
	// Get the pre-built binary path
	binaryPath, err := buildBinary()
	if err != nil {
		return err
	}
	tc.binaryPath = binaryPath

	// Create a temporary directory for XDG_DATA_HOME
	tc.trackingDir, err = os.MkdirTemp("", "structured-cli-tracking-*")
	if err != nil {
		return err
	}

	// Set XDG_DATA_HOME to isolate tracking database
	tc.envVars["XDG_DATA_HOME"] = tc.trackingDir

	// Set the expected tracking database path
	tc.trackingDBPath = filepath.Join(tc.trackingDir, "structured-cli", "tracking.db")

	return nil
}

func (tc *testContext) iHaveATrackingDatabaseWithNRecordedCommands(ctx context.Context, n int) error {
	if err := tc.iHaveACleanTrackingDatabase(ctx); err != nil {
		return err
	}

	// Also need a git repository to run commands against
	if err := tc.iHaveAGitRepository(ctx); err != nil {
		return err
	}

	// Reset envVars to preserve the XDG_DATA_HOME we set
	xdgDataHome := tc.trackingDir
	tc.envVars["XDG_DATA_HOME"] = xdgDataHome

	// Run n commands to populate the database
	for i := 0; i < n; i++ {
		if err := tc.iRun(ctx, "structured-cli --json git status"); err != nil {
			return fmt.Errorf("failed to run command %d: %w", i+1, err)
		}
	}

	return nil
}

func (tc *testContext) iHaveATrackingDatabaseWithOldRecords(ctx context.Context, daysAgo int) error {
	if err := tc.iHaveACleanTrackingDatabase(ctx); err != nil {
		return err
	}

	// Also need a git repository
	if err := tc.iHaveAGitRepository(ctx); err != nil {
		return err
	}

	// Reset envVars to preserve the XDG_DATA_HOME we set
	xdgDataHome := tc.trackingDir
	tc.envVars["XDG_DATA_HOME"] = xdgDataHome

	// Run a few commands to create the database
	for i := 0; i < 3; i++ {
		if err := tc.iRun(ctx, "structured-cli --json git status"); err != nil {
			return fmt.Errorf("failed to run command %d: %w", i+1, err)
		}
	}

	// Update the timestamps to be old (simulate old records)
	db, err := sql.Open("sqlite", tc.trackingDBPath)
	if err != nil {
		return fmt.Errorf("failed to open tracking database: %w", err)
	}
	defer func() { _ = db.Close() }()

	oldTime := time.Now().AddDate(0, 0, -daysAgo)
	_, err = db.ExecContext(ctx, "UPDATE commands SET timestamp = ?", oldTime)
	if err != nil {
		return fmt.Errorf("failed to update timestamps: %w", err)
	}

	return nil
}

func (tc *testContext) iRunInTheTrackingContext(ctx context.Context, command string) error {
	// Run a command that will trigger cleanup of old records
	return tc.iRun(ctx, command)
}

func (tc *testContext) theTrackingDatabaseShouldHaveNCommandRecorded(ctx context.Context, n int) error {
	// Wait briefly for any async writes
	time.Sleep(100 * time.Millisecond)

	db, err := sql.Open("sqlite", tc.trackingDBPath)
	if err != nil {
		return fmt.Errorf("failed to open tracking database: %w", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commands").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count commands: %w", err)
	}

	if count != n {
		return fmt.Errorf("expected %d commands recorded, got %d", n, count)
	}

	return nil
}

func (tc *testContext) theTrackingDatabaseShouldHaveTokenMetricsRecorded(ctx context.Context) error {
	// Wait briefly for any async writes
	time.Sleep(100 * time.Millisecond)

	db, err := sql.Open("sqlite", tc.trackingDBPath)
	if err != nil {
		return fmt.Errorf("failed to open tracking database: %w", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	var rawTokens, parsedTokens int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(raw_tokens), 0), COALESCE(SUM(parsed_tokens), 0) FROM commands",
	).Scan(&count, &rawTokens, &parsedTokens)
	if err != nil {
		return fmt.Errorf("failed to get token metrics: %w", err)
	}

	if count == 0 {
		return errors.New("no commands recorded in tracking database")
	}

	// Verify that token counts are recorded (they should be positive for any command)
	if rawTokens == 0 && parsedTokens == 0 {
		return errors.New("no token metrics recorded (both raw and parsed are 0)")
	}

	return nil
}

func (tc *testContext) theOldRecordsShouldBeRemoved(ctx context.Context) error {
	// Wait briefly for any async writes
	time.Sleep(100 * time.Millisecond)

	db, err := sql.Open("sqlite", tc.trackingDBPath)
	if err != nil {
		return fmt.Errorf("failed to open tracking database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Check if any records are older than 90 days
	cutoff := time.Now().AddDate(0, 0, -90)
	var oldCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commands WHERE timestamp < ?", cutoff).Scan(&oldCount)
	if err != nil {
		return fmt.Errorf("failed to count old commands: %w", err)
	}

	if oldCount > 0 {
		return fmt.Errorf("expected old records to be removed, but found %d", oldCount)
	}

	return nil
}

func (tc *testContext) noTrackingDatabaseShouldBeCreated(_ context.Context) error {
	// Wait briefly for any async writes
	time.Sleep(100 * time.Millisecond)

	// Check that the tracking database file does not exist
	if _, err := os.Stat(tc.trackingDBPath); err == nil {
		return fmt.Errorf("tracking database was created at %s when it should not have been", tc.trackingDBPath)
	}

	return nil
}

func (tc *testContext) theJSONObjectShouldHaveAsAString(_ context.Context, objKey, field string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	obj, ok := data[objKey].(map[string]interface{})
	if !ok {
		return fmt.Errorf("key %q is not an object\nJSON: %s", objKey, tc.output)
	}

	val, ok := obj[field]
	if !ok {
		return fmt.Errorf("object %q does not have field %q", objKey, field)
	}

	if _, ok := val.(string); !ok {
		return fmt.Errorf("field %q in object %q is not a string, got %T", field, objKey, val)
	}

	return nil
}

func (tc *testContext) theJSONObjectShouldHaveAsAnArray(_ context.Context, objKey, field string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	obj, ok := data[objKey].(map[string]interface{})
	if !ok {
		return fmt.Errorf("key %q is not an object\nJSON: %s", objKey, tc.output)
	}

	val, ok := obj[field]
	if !ok {
		return fmt.Errorf("object %q does not have field %q", objKey, field)
	}

	if _, ok := val.([]interface{}); !ok {
		return fmt.Errorf("field %q in object %q is not an array, got %T", field, objKey, val)
	}

	return nil
}

func (tc *testContext) theFirstBlameLineShouldHaveAsAString(_ context.Context, field string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	lines, ok := data["lines"].([]interface{})
	if !ok || len(lines) == 0 {
		return errors.New("no lines found in JSON")
	}

	firstLine, ok := lines[0].(map[string]interface{})
	if !ok {
		return errors.New("first line is not an object")
	}

	val, ok := firstLine[field]
	if !ok {
		return fmt.Errorf("first blame line does not have field %q", field)
	}

	if _, ok := val.(string); !ok {
		return fmt.Errorf("field %q is not a string, got %T", field, val)
	}

	return nil
}

func (tc *testContext) theFirstBlameLineShouldHave(_ context.Context, field string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	lines, ok := data["lines"].([]interface{})
	if !ok || len(lines) == 0 {
		return errors.New("no lines found in JSON")
	}

	firstLine, ok := lines[0].(map[string]interface{})
	if !ok {
		return errors.New("first line is not an object")
	}

	if _, ok := firstLine[field]; !ok {
		return fmt.Errorf("first blame line does not have field %q", field)
	}

	return nil
}

func (tc *testContext) theFirstReflogEntryShouldHaveAsAString(_ context.Context, field string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	entries, ok := data["entries"].([]interface{})
	if !ok || len(entries) == 0 {
		return errors.New("no entries found in JSON")
	}

	firstEntry, ok := entries[0].(map[string]interface{})
	if !ok {
		return errors.New("first entry is not an object")
	}

	val, ok := firstEntry[field]
	if !ok {
		return fmt.Errorf("first reflog entry does not have field %q", field)
	}

	if _, ok := val.(string); !ok {
		return fmt.Errorf("field %q is not a string, got %T", field, val)
	}

	return nil
}

func (tc *testContext) theFirstReflogEntryShouldHave(_ context.Context, field string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	entries, ok := data["entries"].([]interface{})
	if !ok || len(entries) == 0 {
		return errors.New("no entries found in JSON")
	}

	firstEntry, ok := entries[0].(map[string]interface{})
	if !ok {
		return errors.New("first entry is not an object")
	}

	if _, ok := firstEntry[field]; !ok {
		return fmt.Errorf("first reflog entry does not have field %q", field)
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := newTestContext()

	// Background steps
	ctx.Step(`^I have a git repository$`, tc.iHaveAGitRepository)
	ctx.Step(`^I have a git repository with commits$`, tc.iHaveAGitRepositoryWithCommits)

	// Given steps - Git
	ctx.Step(`^the repository has no changes$`, tc.theRepositoryHasNoChanges)
	ctx.Step(`^I create an untracked file "([^"]*)"$`, tc.iCreateAnUntrackedFile)
	ctx.Step(`^I modify a tracked file$`, tc.iModifyATrackedFile)
	ctx.Step(`^I have an empty git repository$`, tc.iHaveAnEmptyGitRepository)
	ctx.Step(`^I am not in a git repository$`, tc.iAmNotInAGitRepository)
	ctx.Step(`^I stage the modified file$`, tc.iStageTheModifiedFile)
	ctx.Step(`^the environment variable "([^"]*)" is set to "([^"]*)"$`, tc.theEnvironmentVariableIsSetTo)

	// Given steps - File operations
	ctx.Step(`^I have a directory with files$`, tc.iHaveADirectoryWithFiles)
	ctx.Step(`^I have a file "([^"]*)" with content "([^"]*)"$`, tc.iHaveAFileWithContent)
	ctx.Step(`^I have a file "([^"]*)" with multiple lines$`, tc.iHaveAFileWithMultipleLines)

	// Given steps - Make/Just
	ctx.Step(`^I have a Makefile with target "([^"]*)"$`, tc.iHaveAMakefileWithTarget)
	ctx.Step(`^I have a Makefile with failing target "([^"]*)"$`, tc.iHaveAMakefileWithFailingTarget)
	ctx.Step(`^I have a justfile with recipe "([^"]*)"$`, tc.iHaveAJustfileWithRecipe)
	ctx.Step(`^I have a justfile with failing recipe "([^"]*)"$`, tc.iHaveAJustfileWithFailingRecipe)

	// Given steps - Go
	ctx.Step(`^I have a Go project$`, tc.iHaveAGoProject)
	ctx.Step(`^the Go project has a syntax error$`, tc.theGoProjectHasASyntaxError)
	ctx.Step(`^the Go project has tests$`, tc.theGoProjectHasTests)
	ctx.Step(`^the Go project has vet issues$`, tc.theGoProjectHasVetIssues)
	ctx.Step(`^the Go project has format issues$`, tc.theGoProjectHasFormatIssues)

	// Given steps - Node.js/NPM
	ctx.Step(`^I have a Node\.js project$`, tc.iHaveANodejsProject)
	ctx.Step(`^I have a Node\.js project with no dependencies$`, tc.iHaveANodejsProjectWithNoDependencies)

	// Given steps - Docker
	ctx.Step(`^docker is available$`, tc.dockerIsAvailable)

	// When steps
	ctx.Step(`^I run "([^"]*)"$`, tc.iRun)

	// Then steps
	ctx.Step(`^the exit code should be (\d+)$`, tc.theExitCodeShouldBe)
	ctx.Step(`^the output should be valid JSON$`, tc.theOutputShouldBeValidJSON)
	ctx.Step(`^the output should not be JSON$`, tc.theOutputShouldNotBeJSON)
	ctx.Step(`^the output should contain "([^"]*)"$`, tc.theOutputShouldContain)
	ctx.Step(`^the JSON should contain key "([^"]*)" with boolean value (true|false)$`, tc.theJSONShouldContainKeyWithBooleanValue)
	ctx.Step(`^the JSON should contain key "([^"]*)" as a string$`, tc.theJSONShouldContainKeyAsAString)
	ctx.Step(`^the JSON "([^"]*)" string should not be empty$`, tc.theJSONStringShouldNotBeEmpty)
	ctx.Step(`^the JSON should contain key "([^"]*)" as an array$`, tc.theJSONShouldContainKeyAsAnArray)
	ctx.Step(`^the JSON should contain key "([^"]*)"$`, tc.theJSONShouldContainKey)
	ctx.Step(`^the JSON "([^"]*)" array should contain "([^"]*)"$`, tc.theJSONArrayShouldContain)
	ctx.Step(`^the JSON "([^"]*)" array should not be empty$`, tc.theJSONArrayShouldNotBeEmpty)
	ctx.Step(`^the JSON "([^"]*)" array should be empty$`, tc.theJSONArrayShouldBeEmpty)
	ctx.Step(`^the JSON "([^"]*)" array should have at most (\d+) items$`, tc.theJSONArrayShouldHaveAtMostNItems)
	ctx.Step(`^the JSON "([^"]*)" array should have at least (\d+) items$`, tc.theJSONArrayShouldHaveAtLeastNItems)
	ctx.Step(`^the first commit should have "([^"]*)" as a string$`, tc.theFirstCommitShouldHaveAsAString)
	ctx.Step(`^the first file should have "([^"]*)" as a string$`, tc.theFirstFileShouldHaveAsAString)
	ctx.Step(`^one branch should have "current" equal to true$`, tc.oneBranchShouldHaveCurrentEqualToTrue)
	ctx.Step(`^the JSON "([^"]*)" object should have "([^"]*)" as a string$`, tc.theJSONObjectShouldHaveAsAString)
	ctx.Step(`^the JSON "([^"]*)" object should have "([^"]*)" as an array$`, tc.theJSONObjectShouldHaveAsAnArray)
	ctx.Step(`^the first blame line should have "([^"]*)" as a string$`, tc.theFirstBlameLineShouldHaveAsAString)
	ctx.Step(`^the first blame line should have "([^"]*)"$`, tc.theFirstBlameLineShouldHave)
	ctx.Step(`^the first reflog entry should have "([^"]*)" as a string$`, tc.theFirstReflogEntryShouldHaveAsAString)
	ctx.Step(`^the first reflog entry should have "([^"]*)"$`, tc.theFirstReflogEntryShouldHave)

	// Given steps - Tracking
	ctx.Step(`^I have a clean tracking database$`, tc.iHaveACleanTrackingDatabase)
	ctx.Step(`^I have a tracking database with (\d+) recorded commands$`, tc.iHaveATrackingDatabaseWithNRecordedCommands)
	ctx.Step(`^I have a tracking database with old records from (\d+) days ago$`, tc.iHaveATrackingDatabaseWithOldRecords)

	// When steps - Tracking
	ctx.Step(`^I run "([^"]*)" in the tracking context$`, tc.iRunInTheTrackingContext)

	// Then steps - Tracking
	ctx.Step(`^the tracking database should have (\d+) command recorded$`, tc.theTrackingDatabaseShouldHaveNCommandRecorded)
	ctx.Step(`^the tracking database should have token metrics recorded$`, tc.theTrackingDatabaseShouldHaveTokenMetricsRecorded)
	ctx.Step(`^the old records should be removed$`, tc.theOldRecordsShouldBeRemoved)
	ctx.Step(`^no tracking database should be created$`, tc.noTrackingDatabaseShouldBeCreated)

	// Cleanup after each scenario
	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		// Clean up test files
		if tc.repoDir != "" {
			testFile := filepath.Join(tc.repoDir, "test_untracked.txt")
			os.Remove(testFile)
		}
		if tc.tempDir != "" {
			os.RemoveAll(tc.tempDir)
		}
		if tc.emptyRepoDir != "" {
			os.RemoveAll(tc.emptyRepoDir)
		}
		if tc.nonGitDir != "" {
			os.RemoveAll(tc.nonGitDir)
		}
		if tc.testFile != "" {
			os.Remove(tc.testFile)
		}
		// Clean up tracking directory
		if tc.trackingDir != "" {
			os.RemoveAll(tc.trackingDir)
		}
		return ctx, nil
	})
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"./"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
