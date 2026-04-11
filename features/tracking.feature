Feature: Usage Tracking
  As an AI coding agent user
  I want structured-cli to track my usage
  So that I can see statistics on token savings and command history

  Scenario: Tracking records commands after JSON output
    Given I have a clean tracking database
    And I have a git repository
    When I run "structured-cli --json git status"
    Then the exit code should be 0
    And the output should be valid JSON
    And the tracking database should have 1 command recorded

  Scenario: Tracking calculates token metrics
    Given I have a clean tracking database
    And I have a directory with files
    When I run "structured-cli --json ls"
    Then the exit code should be 0
    And the output should be valid JSON
    And the tracking database should have token metrics recorded

  Scenario: Stats command shows summary
    Given I have a tracking database with 10 recorded commands
    When I run "structured-cli stats"
    Then the exit code should be 0
    And the output should contain "Tokens Saved:"
    And the output should contain "10"

  Scenario: Stats --history shows recent commands
    Given I have a tracking database with 10 recorded commands
    When I run "structured-cli stats --history"
    Then the exit code should be 0
    And the output should contain "Recent Command History"
    And the output should contain "TIMESTAMP"

  Scenario: Stats --json outputs valid JSON
    Given I have a tracking database with 10 recorded commands
    When I run "structured-cli stats --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "total_commands"
    And the JSON should contain key "total_tokens_saved"
    And the JSON should contain key "avg_savings_percent"

  Scenario: 90-day cleanup removes old records
    Given I have a tracking database with old records from 100 days ago
    When I run "structured-cli --json git status" in the tracking context
    Then the exit code should be 0
    And the tracking database should have 1 command recorded
    And the old records should be removed

  Scenario: Disabled tracking with environment variable
    Given I have a clean tracking database
    And the environment variable "STRUCTURED_CLI_NO_TRACKING" is set to "1"
    And I have a git repository
    When I run "structured-cli --json git status"
    Then the exit code should be 0
    And the output should be valid JSON
    And no tracking database should be created
