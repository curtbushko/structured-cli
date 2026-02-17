Feature: Output Modes
  As a developer
  I want to control structured-cli output format
  So that I can use it as a drop-in replacement or get structured data

  Background:
    Given I have a git repository

  Scenario: Passthrough mode by default
    When I run "structured-cli git status"
    Then the exit code should be 0
    And the output should not be JSON
    And the output should contain "On branch"

  Scenario: JSON output with --json flag
    When I run "structured-cli git status --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "branch"

  Scenario: JSON output with environment variable
    Given the environment variable "STRUCTURED_CLI_JSON" is set to "true"
    When I run "structured-cli git status"
    Then the exit code should be 0
    And the output should be valid JSON

  Scenario: --json flag position before command
    When I run "structured-cli --json git status"
    Then the exit code should be 0
    And the output should be valid JSON

  Scenario: --json flag position after command
    When I run "structured-cli git status --json"
    Then the exit code should be 0
    And the output should be valid JSON
