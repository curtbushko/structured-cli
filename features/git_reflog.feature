Feature: Git Reflog
  As an AI coding agent
  I want structured output from git reflog
  So that I can analyze reference log entries programmatically

  Background:
    Given I have a git repository with commits

  Scenario: git reflog - basic output
    When I run "structured-cli git reflog --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "entries" as an array

  Scenario: git reflog - entries contain required fields
    When I run "structured-cli git reflog --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON "entries" array should not be empty
    And the first reflog entry should have "hash" as a string
    And the first reflog entry should have "index"
    And the first reflog entry should have "action" as a string
    And the first reflog entry should have "message" as a string

  Scenario: git reflog - passthrough mode
    When I run "structured-cli git reflog"
    Then the exit code should be 0
    And the output should not be JSON
    And the output should contain "HEAD@{0}"
