# Contributing to tfq

Thank you for your interest in contributing to tfq. This document provides guidelines and instructions for contributing to the project.

## Getting Started

To contribute to this project, follow these steps:

1.  Fork the repository on GitHub.
2.  Clone your fork to your local machine.
3.  Create a new branch for your changes. Use a descriptive name for your branch, such as `fix-issue-123` or `add-feature-abc`.

## Development Environment

This project is written in Go. Ensure you have a recent version of Go installed on your system.

The project uses a `Makefile` to automate common development tasks:

*   `make fmt`: Formats the Go source code.
*   `make fmtcheck`: Checks if the code is properly formatted.
*   `make build`: Builds the project using GoReleaser (requires GoReleaser to be installed).
*   `make unittest`: Runs unit tests.
*   `make test`: Runs all tests (requires environment variables, see Testing section).
*   `make generatemocks`: Generates mocks for testing.

## Submitting Changes

When you are ready to submit your changes:

1.  Ensure your code follows the project's coding standards and is properly formatted using `make fmt`.
2.  Commit your changes with clear and descriptive commit messages.
3.  Push your branch to your fork on GitHub.
4.  Open a Pull Request (PR) against the `main` branch of the original repository.
5.  Fill out the PR template questionnaire completely.
6.  Ensure that all CI checks pass.

## Testing

Before submitting a PR, ensure that all tests pass.

Some tests require access to a Terraform Enterprise or Terraform Cloud instance. The following environment variables must be set to run these tests:

*   `TFE_TOKEN`: A valid TFE/TFC API token.
*   `TFE_ORG`: The name of the TFE/TFC organization to use for testing.
*   `TFE_ADDRESS`: (Optional) The address of the TFE instance. Defaults to `https://app.terraform.io/`.

Run unit tests with:
```bash
make unittest
```

Run all tests with:
```bash
make test
```

## Coding Standards

*   Use standard Go formatting (`gofmt`).
*   Write clear, concise, and documented code.
*   Ensure new features or bug fixes include corresponding tests.
*   Avoid introducing new dependencies unless absolutely necessary.

## Reporting Issues

If you encounter a bug or have a feature request, please open an issue on the GitHub repository.

### Bug Reports

When reporting a bug, please include:

*   The version of tfq you are using.
*   Your operating system and environment details.
*   Steps to reproduce the issue.
*   Expected and actual behavior.
*   Relevant log output (use the `-l debug` flag for more detail).

### Feature Requests

When requesting a feature, please describe:

*   The problem you are trying to solve.
*   The proposed solution or interface.
*   Any alternatives you have considered.
