Feature: Error Handling
  As an AI coding agent
  I want errors returned as JSON when in JSON mode
  So that I can handle failures programmatically

  Scenario: Command failure in JSON mode
    Given I am not in a git repository
    When I run "structured-cli git status --json"
    Then the output should be valid JSON
    And the JSON should contain key "error"
    And the JSON should contain key "exitCode"
    And the JSON should contain key "raw"

  Scenario: Command failure in passthrough mode
    Given I am not in a git repository
    When I run "structured-cli git status"
    Then the output should contain "fatal"

  Scenario: Non-zero exit code is preserved
    Given I am not in a git repository
    When I run "structured-cli git status --json"
    Then the exit code should be 128

  Scenario: Parser failure preserves raw output for debugging
    Given I am not in a git repository
    When I run "structured-cli git status --json"
    Then the output should be valid JSON
    And the JSON should contain key "raw" as a string
    And the JSON "raw" string should not be empty
