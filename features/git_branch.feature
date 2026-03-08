Feature: Git Branch
  As an AI coding agent
  I want structured output from git branch
  So that I can manage branches programmatically

  Background:
    Given I have a git repository

  Scenario: git branch - list branches
    When I run "structured-cli git branch --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "branches" as an array
    And the JSON "branches" array should not be empty

  Scenario: git branch - current branch detection
    When I run "structured-cli git branch --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And one branch should have "current" equal to true

  Scenario: git branch - create new branch
    When I run "structured-cli git branch test-branch"
    Then the exit code should be 0
    When I run "structured-cli git branch --json"
    Then the output should be valid JSON
    And the JSON "branches" array should have at least 2 items
