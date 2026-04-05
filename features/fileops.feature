Feature: File Operations
  As an AI coding agent
  I want structured output from file operations
  So that I can navigate and search filesystems programmatically

  Scenario: ls - list current directory
    Given I have a directory with files
    When I run "structured-cli ls -la --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "entries" as an array
    And the JSON "entries" array should not be empty

  Scenario: ls - list specific path
    Given I have a directory with files
    When I run "structured-cli ls -la subdir --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "entries" as an array

  Scenario: ls - passthrough mode
    Given I have a directory with files
    When I run "structured-cli ls"
    Then the exit code should be 0
    And the output should not be JSON

  Scenario: cat - read file contents
    Given I have a file "test.txt" with content "Hello World"
    When I run "structured-cli cat test.txt --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "content" as a string
    And the JSON should contain key "bytes"

  Scenario: wc - word count
    Given I have a file "test.txt" with content "Hello World"
    When I run "structured-cli wc test.txt --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "files" as an array

  Scenario: head - read first lines
    Given I have a file "test.txt" with multiple lines
    When I run "structured-cli head -n 5 test.txt --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "content" as a string
    And the JSON should contain key "lineCount"

  Scenario: tail - read last lines
    Given I have a file "test.txt" with multiple lines
    When I run "structured-cli tail -n 5 test.txt --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "content" as a string
    And the JSON should contain key "lineCount"

  Scenario: find - search by name
    Given I have a directory with files
    When I run "structured-cli find . -name *.txt --json"
    Then the output should be valid JSON
    And the JSON should contain key "files" as an array

  Scenario: find - search by type
    Given I have a directory with files
    When I run "structured-cli find . -type f --json"
    Then the output should be valid JSON
    And the JSON should contain key "files" as an array

  Scenario: grep - search in files
    Given I have a file "test.txt" with content "Hello World"
    When I run "structured-cli grep Hello test.txt --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "matches" as an array

  Scenario: grep - pattern matching with line numbers
    Given I have a file "test.txt" with multiple lines
    When I run "structured-cli grep -n Line test.txt --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "matches" as an array
    And the JSON should contain key "count"

  Scenario: du - disk usage
    Given I have a directory with files
    When I run "structured-cli du -sh . --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "entries" as an array

  Scenario: df - disk free
    Given I have a directory with files
    When I run "structured-cli df -h . --json"
    Then the exit code should be 0
    And the output should be valid JSON
    And the JSON should contain key "filesystems" as an array
