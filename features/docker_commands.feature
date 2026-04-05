Feature: Docker Commands
  As an AI coding agent
  I want structured output from Docker commands
  So that I can manage containers and images programmatically

  # docker ps scenarios
  @docker
  Scenario: docker ps - list containers
    Given docker is available
    When I run "structured-cli --json --disable-filter=small docker ps -a"
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value true
    And the JSON should contain key "containers" as an array

  @docker
  Scenario: docker ps - passthrough mode
    Given docker is available
    When I run "structured-cli docker ps -a"
    Then the output should not be JSON

  # docker images scenarios
  @docker
  Scenario: docker images - list images
    Given docker is available
    When I run "structured-cli --json --disable-filter=small docker images"
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value true
    And the JSON should contain key "images" as an array

  @docker
  Scenario: docker images - passthrough mode
    Given docker is available
    When I run "structured-cli docker images"
    Then the output should not be JSON
