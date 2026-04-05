Feature: Git Blame
  As an AI coding agent
  I want structured output from git blame
  So that I can analyze file attribution programmatically

  Background:
    Given I have a git repository

  Scenario: git blame - basic output
    When I run "structured-cli git blame --porcelain README.md --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "file" as a string
    And the JSON should contain key "lines" as an array

  Scenario: git blame - lines contain attribution
    When I run "structured-cli git blame --porcelain README.md --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON "lines" array should not be empty
    And the first blame line should have "hash" as a string
    And the first blame line should have "lineNumber"
    And the first blame line should have "author" as a string
    And the first blame line should have "content" as a string

  Scenario: git blame - passthrough mode
    When I run "structured-cli git blame README.md"
    Then the exit code should be 0
    And the output should not be JSON
