Feature: Unsupported Commands
  As a developer
  I want graceful handling of commands without parsers
  So that the tool doesn't break on obscure subcommands

  Background:
    Given I have a git repository

  Scenario: Unsupported subcommand in JSON mode returns fallback
    When I run "structured-cli --json --disable-filter=small git stash list"
    Then the output should be valid JSON
    And the JSON should contain key "raw"
    And the JSON should contain key "parsed" with boolean value false
    And the JSON should contain key "exitCode"

  Scenario: Unsupported subcommand in passthrough mode
    When I run "structured-cli git stash list"
    Then the output should not be JSON
