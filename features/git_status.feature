Feature: Git Status
  As an AI coding agent
  I want structured output from git status
  So that I can understand repository state without parsing text

  Background:
    Given I have a git repository

  Scenario: git status - clean repository
    Given the repository has no changes
    When I run "structured-cli git status --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "clean" with boolean value true
    And the JSON should contain key "branch" as a string
    And the JSON should contain key "staged" as an array
    And the JSON should contain key "modified" as an array
    And the JSON should contain key "untracked" as an array

  Scenario: git status - with untracked file
    Given I create an untracked file "test_untracked.txt"
    When I run "structured-cli git status --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "clean" with boolean value false
    And the JSON "untracked" array should contain "test_untracked.txt"

  Scenario: git status - with modified file
    Given I modify a tracked file
    When I run "structured-cli git status --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "clean" with boolean value false
    And the JSON "modified" array should not be empty
