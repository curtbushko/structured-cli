Feature: NPM Commands
  As an AI coding agent
  I want structured output from npm commands
  So that I can manage Node.js dependencies programmatically

  # npm list scenarios
  @npm
  Scenario: npm list - show dependency tree
    Given I have a Node.js project
    When I run "structured-cli --json --disable-filter=small npm list"
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value true
    And the JSON should contain key "dependencies" as an array

  @npm
  Scenario: npm list - empty project (no dependencies)
    Given I have a Node.js project with no dependencies
    When I run "structured-cli --json --disable-filter=small npm list"
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value true
    And the JSON should contain key "dependencies" as an array
    And the JSON "dependencies" array should be empty

  @npm
  Scenario: npm list - passthrough mode
    Given I have a Node.js project
    When I run "structured-cli npm list"
    Then the output should not be JSON

  # npm outdated scenarios
  @npm
  Scenario: npm outdated - no outdated packages
    Given I have a Node.js project
    When I run "structured-cli --json --disable-filter=small npm outdated"
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value true
    And the JSON should contain key "packages" as an array
    And the JSON "packages" array should be empty

  @npm
  Scenario: npm outdated - passthrough mode
    Given I have a Node.js project
    When I run "structured-cli npm outdated"
    Then the output should not be JSON
