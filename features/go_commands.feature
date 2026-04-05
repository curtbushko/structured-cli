Feature: Go Commands
  As an AI coding agent
  I want structured output from Go toolchain commands
  So that I can build and test Go code programmatically

  Background:
    Given I have a Go project

  # go build scenarios
  Scenario: go build - successful build
    When I run "structured-cli --json --disable-filter=small go build ./..."
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value true
    And the JSON should contain key "errors" as an array
    And the JSON "errors" array should be empty

  Scenario: go build - build error
    Given the Go project has a syntax error
    When I run "structured-cli --json --disable-filter=small go build ./..."
    Then the output should be valid JSON
    And the JSON should contain key "success" with boolean value false
    And the JSON should contain key "errors" as an array
    And the JSON "errors" array should not be empty

  Scenario: go build - passthrough mode
    When I run "structured-cli go build ./..."
    Then the output should not be JSON

  # go test scenarios
  Scenario: go test - run passing tests
    Given the Go project has tests
    When I run "structured-cli --json --disable-filter=small go test -json ./..."
    Then the output should be valid JSON
    And the JSON should contain key "passed"
    And the JSON should contain key "failed"
    And the JSON should contain key "packages" as an array

  Scenario: go test - with coverage
    Given the Go project has tests
    When I run "structured-cli --json --disable-filter=small go test -json -cover ./..."
    Then the output should be valid JSON
    And the JSON should contain key "coverage"

  Scenario: go test - passthrough mode
    Given the Go project has tests
    When I run "structured-cli go test ./..."
    Then the output should not be JSON

  # go vet scenarios
  Scenario: go vet - clean code
    When I run "structured-cli --json --disable-filter=small go vet ./..."
    Then the output should be valid JSON
    And the JSON should contain key "issues" as an array
    And the JSON "issues" array should be empty

  Scenario: go vet - with issues
    Given the Go project has vet issues
    When I run "structured-cli --json --disable-filter=small go vet ./..."
    Then the output should be valid JSON
    And the JSON should contain key "issues" as an array
    And the JSON "issues" array should not be empty

  Scenario: go vet - passthrough mode
    When I run "structured-cli go vet ./..."
    Then the output should not be JSON

  # go fmt scenarios
  Scenario: go fmt - properly formatted code
    When I run "structured-cli --json --disable-filter=small gofmt -l ."
    Then the output should be valid JSON
    And the JSON should contain key "unformatted" as an array
    And the JSON "unformatted" array should be empty

  Scenario: go fmt - code needs formatting
    Given the Go project has format issues
    When I run "structured-cli --json --disable-filter=small gofmt -l ."
    Then the output should be valid JSON
    And the JSON should contain key "unformatted" as an array
    And the JSON "unformatted" array should not be empty

  Scenario: go fmt - passthrough mode
    When I run "structured-cli gofmt -l ."
    Then the output should not be JSON
