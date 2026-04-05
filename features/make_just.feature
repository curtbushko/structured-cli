Feature: Make and Just Commands
  As an AI coding agent
  I want structured output from make and just commands
  So that I can run build tasks programmatically

  Scenario: make - successful target
    Given I have a Makefile with target "test"
    When I run "structured-cli --json --disable-filter=small make test"
    Then the output should be valid JSON
    And the JSON should contain key "success"
    And the JSON should contain key "exit_code"

  Scenario: make - target failure
    Given I have a Makefile with failing target "broken"
    When I run "structured-cli --json --disable-filter=small make broken"
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value false
    And the JSON should contain key "exit_code"

  Scenario: make - passthrough mode
    Given I have a Makefile with target "test"
    When I run "structured-cli make test"
    Then the output should not be JSON

  Scenario: just - successful recipe
    Given I have a justfile with recipe "build"
    When I run "structured-cli --json --disable-filter=small just build"
    Then the output should be valid JSON
    And the JSON should contain key "success"
    And the JSON should contain key "exit_code"

  Scenario: just - recipe failure
    Given I have a justfile with failing recipe "broken"
    When I run "structured-cli --json --disable-filter=small just broken"
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value false

  Scenario: just - list recipes
    Given I have a justfile with recipe "build"
    When I run "structured-cli --json --disable-filter=small just --list"
    Then the output should be valid JSON
    And the JSON should contain key "recipes" as an array
