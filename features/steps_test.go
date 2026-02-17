package features

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"
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
	tempDir      string
	repoDir      string
	output       string
	exitCode     int
	envVars      map[string]string
	binaryPath   string
	emptyRepoDir string
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

func (tc *testContext) theJSONShouldContainKeyAsAnArray(_ context.Context, key string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(tc.output), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	val, ok := data[key]
	if !ok {
		return fmt.Errorf("JSON does not contain key %q\nJSON: %s", key, tc.output)
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

	arr, ok := data[key].([]interface{})
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

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := newTestContext()

	// Background steps
	ctx.Step(`^I have a git repository$`, tc.iHaveAGitRepository)
	ctx.Step(`^I have a git repository with commits$`, tc.iHaveAGitRepositoryWithCommits)

	// Given steps
	ctx.Step(`^the repository has no changes$`, tc.theRepositoryHasNoChanges)
	ctx.Step(`^I create an untracked file "([^"]*)"$`, tc.iCreateAnUntrackedFile)
	ctx.Step(`^I modify a tracked file$`, tc.iModifyATrackedFile)
	ctx.Step(`^I have an empty git repository$`, tc.iHaveAnEmptyGitRepository)
	ctx.Step(`^the environment variable "([^"]*)" is set to "([^"]*)"$`, tc.theEnvironmentVariableIsSetTo)

	// When steps
	ctx.Step(`^I run "([^"]*)"$`, tc.iRun)

	// Then steps
	ctx.Step(`^the exit code should be (\d+)$`, tc.theExitCodeShouldBe)
	ctx.Step(`^the output should be valid JSON$`, tc.theOutputShouldBeValidJSON)
	ctx.Step(`^the output should not be JSON$`, tc.theOutputShouldNotBeJSON)
	ctx.Step(`^the output should contain "([^"]*)"$`, tc.theOutputShouldContain)
	ctx.Step(`^the JSON should contain key "([^"]*)" with boolean value (true|false)$`, tc.theJSONShouldContainKeyWithBooleanValue)
	ctx.Step(`^the JSON should contain key "([^"]*)" as a string$`, tc.theJSONShouldContainKeyAsAString)
	ctx.Step(`^the JSON should contain key "([^"]*)" as an array$`, tc.theJSONShouldContainKeyAsAnArray)
	ctx.Step(`^the JSON should contain key "([^"]*)"$`, tc.theJSONShouldContainKey)
	ctx.Step(`^the JSON "([^"]*)" array should contain "([^"]*)"$`, tc.theJSONArrayShouldContain)
	ctx.Step(`^the JSON "([^"]*)" array should not be empty$`, tc.theJSONArrayShouldNotBeEmpty)
	ctx.Step(`^the JSON "([^"]*)" array should be empty$`, tc.theJSONArrayShouldBeEmpty)
	ctx.Step(`^the JSON "([^"]*)" array should have at most (\d+) items$`, tc.theJSONArrayShouldHaveAtMostNItems)
	ctx.Step(`^the first commit should have "([^"]*)" as a string$`, tc.theFirstCommitShouldHaveAsAString)

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
