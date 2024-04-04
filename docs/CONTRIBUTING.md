# Contributing to Maestro 🎼

First off, thank you for considering contributing to Maestro! We appreciate your interest and support. This document
provides a set of guidelines for contributing to the project. By following these guidelines, you can help us maintain a
collaborative and welcoming environment for all contributors.

## Getting Started 🚀

1. Fork the repository on GitHub.
2. Clone your fork to your local machine.
3. Create a new branch for your changes: `git checkout -b my-feature-branch`.
4. Make your changes and commit them with descriptive commit messages.
5. Push your changes to your fork: `git push origin my-feature-branch`.
6. Open a pull request on the main repository.

## How to Contribute 💡

### Reporting Bugs 🐛

If you find a bug, please open an issue on the GitHub repository. When reporting a bug, please include:

- A clear and descriptive title
- Steps to reproduce the bug
- Expected behavior
- Actual behavior
- Any relevant error messages or logs

### Suggesting Enhancements 💡

If you have an idea for an enhancement or a new feature, please open an issue on the GitHub repository. When suggesting
an enhancement, please include:

- A clear and descriptive title
- A detailed description of the proposed enhancement
- Any relevant examples or use cases

### Submitting Pull Requests 🔀

When submitting a pull request, please ensure that your changes adhere to the following guidelines:

- Write clear and concise commit messages
- Follow the existing code style and conventions
- Include tests for your changes, if applicable
- Update the documentation, if necessary

## Development Setup 💻

To set up the development environment for Maestro, follow these steps:

1. Install Go 1.18 or higher.
2. Clone the repository: `git clone https://github.com/ethan-gallant/maestro.git`.
3. Navigate to the project directory: `cd maestro`.
4. Install the required dependencies: `go mod download`.
5. Run the tests to ensure everything is set up correctly: `go test ./...`.

## Code Style 🎨

Please follow the existing code style and conventions used in the project. We use [gofmt](https://golang.org/cmd/gofmt/)
to format the code.

## Testing 🧪

When contributing code changes, please include appropriate tests to ensure the correctness and reliability of the code.
Run the tests using the following command:

```shell
go test ./...
```

## Documentation 📚

If your changes require updates to the documentation, please make the necessary changes in the relevant Markdown files
or code comments.