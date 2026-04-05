Feature: Git Show
  As an AI coding agent
  I want structured output from git show
  So that I can analyze commit details programmatically

  Background:
    Given I have a git repository with commits

  Scenario: git show - default output shows commit details
    When I run "structured-cli git show --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "commit"
    And the JSON should contain key "diff"

  Scenario: git show - commit has required fields
    When I run "structured-cli git show --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON "commit" object should have "hash" as a string
    And the JSON "commit" object should have "author" as a string
    And the JSON "commit" object should have "date" as a string
    And the JSON "commit" object should have "subject" as a string

  Scenario: git show - diff contains files array
    When I run "structured-cli git show --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON "diff" object should have "files" as an array

  Scenario: git show - passthrough mode
    When I run "structured-cli git show"
    Then the exit code should be 0
    And the output should not be JSON
    And the output should contain "commit"
