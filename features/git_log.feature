Feature: Git Log
  As an AI coding agent
  I want structured output from git log
  So that I can analyze commit history programmatically

  Background:
    Given I have a git repository with commits

  Scenario: git log - basic output
    When I run "structured-cli git log -n 3 --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "commits" as an array
    And the JSON "commits" array should have at most 3 items

  Scenario: git log - commit fields
    When I run "structured-cli git log -n 1 --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the first commit should have "hash" as a string
    And the first commit should have "author" as a string
    And the first commit should have "email" as a string
    And the first commit should have "date" as a string
    And the first commit should have "subject" as a string

  Scenario: git log - empty repository returns valid JSON with exit code
    Given I have an empty git repository
    When I run "structured-cli git log --json"
    Then the output should be valid JSON
    And the JSON should contain key "commits" as an array
