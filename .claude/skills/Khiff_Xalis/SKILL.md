```markdown
# Khiff_Xalis Development Patterns

> Auto-generated skill from repository analysis

## Overview
This skill teaches the core development patterns and conventions used in the Khiff_Xalis Go codebase. You'll learn how to structure files, write imports and exports, follow commit message conventions, and organize tests. This guide also provides suggested commands for common workflows to streamline your development process.

## Coding Conventions

### File Naming
- Use **camelCase** for file names.

  **Example:**
  ```
  userService.go
  dataFetcher.go
  ```

### Import Style
- Use **relative imports** for referencing local packages.

  **Example:**
  ```go
  import (
      "fmt"
      "../utils"
  )
  ```

### Export Style
- Use **named exports**. In Go, exported identifiers start with an uppercase letter.

  **Example:**
  ```go
  // Exported function
  func FetchData() {}

  // Exported type
  type User struct {}
  ```

### Commit Messages
- Use **conventional commits** with the `feat` prefix for new features.
- Keep commit messages concise (average: 72 characters).

  **Example:**
  ```
  feat: add user authentication middleware
  ```

## Workflows

### Feature Development
**Trigger:** When adding a new feature  
**Command:** `/feature`

1. Create a new branch for the feature.
2. Implement the feature following the coding conventions.
3. Write or update tests in corresponding `*.test.*` files.
4. Commit changes using the `feat` prefix.
5. Open a pull request for review.

### Importing Local Packages
**Trigger:** When you need to use code from another local package  
**Command:** `/import-local`

1. Use a relative import path to reference the package.
2. Ensure the imported identifiers are exported (start with uppercase).

   **Example:**
   ```go
   import "../utils"
   ```

### Writing Tests
**Trigger:** When adding or updating functionality  
**Command:** `/write-test`

1. Create a test file matching the pattern `*.test.*` (e.g., `userService.test.go`).
2. Write test functions using Go's testing conventions.
3. Run tests to ensure correctness.

## Testing Patterns

- Test files follow the pattern: `*.test.*` (e.g., `orderService.test.go`).
- The testing framework is not explicitly specified; use Go's built-in `testing` package unless otherwise noted.

  **Example:**
  ```go
  // userService.test.go
  package userService

  import "testing"

  func TestFetchData(t *testing.T) {
      // test logic here
  }
  ```

## Commands

| Command        | Purpose                                         |
|----------------|-------------------------------------------------|
| /feature       | Start a new feature development workflow        |
| /import-local  | Import a local package using relative imports   |
| /write-test    | Create and run tests for new or updated code    |
```
