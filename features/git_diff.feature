Feature: Git Diff
  As an AI coding agent
  I want structured output from git diff
  So that I can analyze code changes programmatically

  Background:
    Given I have a git repository

  Scenario: git diff - no changes
    Given the repository has no changes
    When I run "structured-cli git diff --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "files" as an array
    And the JSON "files" array should be empty

  Scenario: git diff - with unstaged changes
    Given I modify a tracked file
    When I run "structured-cli git diff --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "files" as an array
    And the JSON "files" array should not be empty

  Scenario: git diff - with staged changes
    Given I modify a tracked file
    And I stage the modified file
    When I run "structured-cli git diff --staged --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "files" as an array
    And the JSON "files" array should not be empty

  Scenario: git diff - file has path, additions, deletions
    Given I modify a tracked file
    When I run "structured-cli git diff --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the first file should have "path" as a string
