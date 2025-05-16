# Contributing to k8s-cloud-installer

Thank you for your interest in contributing to the Kubernetes Cloud Installer project! This document provides guidelines and workflows for contributing.

## Code of Conduct

This project adheres to the Contributor Covenant code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Ways to Contribute

There are many ways to contribute to this project:

- Reporting bugs
- Suggesting enhancements
- Writing documentation
- Adding support for new cloud providers
- Improving existing code
- Adding tests

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork to your local machine
3. Create a new branch for your feature or bugfix
4. Make your changes
5. Run tests and ensure they pass
6. Commit your changes with clear, descriptive messages
7. Push to your fork
8. Open a pull request

## Development Environment Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/k8s-cloud-installer.git
cd k8s-cloud-installer

# Install Go (if not already installed)
# Visit https://golang.org/doc/install for instructions

# Build the project
go build -o k8s-installer cmd/installer/main.go
```

## Project Structure

```
k8s-cloud-installer/
├── cmd/                 # Command-line applications
│   └── installer/       # Main installer command
│       └── main.go      # Entry point
├── pkg/                 # Library code
│   ├── config/          # Configuration handling
│   ├── ssh/             # SSH client operations
│   ├── providers/       # Cloud provider implementations
│   └── installer/       # Kubernetes installation logic
├── docs/                # Documentation
├── examples/            # Example scripts
└── scripts/             # Utility scripts
```

## Adding a New Cloud Provider

To add support for a new cloud provider:

1. Create a new file in `pkg/providers/` (e.g., `newprovider.go`)
2. Implement the `Provider` interface defined in `providers.go`
3. Add the new provider to the `NewProvider()` function in `providers.go`
4. Update the `CloudProvider` type in `pkg/config/config.go`
5. Add documentation in `README.md` and create a specific guide in `docs/`

## Testing

Before submitting a pull request, make sure to test your changes:

```bash
# Run unit tests
go test ./...

# Build and manually test the installer
go build -o k8s-installer cmd/installer/main.go
./k8s-installer -host=<VM_IP> -key=<PATH_TO_KEY> -provider=<PROVIDER>
```

## Pull Request Process

1. Update the README.md and documentation with details of changes
2. Update the CHANGELOG.md with a description of the changes
3. Ensure all tests pass
4. Ensure the code follows the project's coding style
5. The PR will be merged once it receives approval from a maintainer

## Coding Style

This project follows the standard Go coding style. Please ensure your code:

- Is formatted with `gofmt`
- Passes `golint` and `go vet`
- Includes appropriate comments and documentation
- Has meaningful variable and function names

## License

By contributing to this project, you agree that your contributions will be licensed under the project's MIT License.

## Questions and Discussions

If you have questions or want to discuss ideas, please open an issue on GitHub. We welcome all feedback and suggestions!

Thank you for contributing to the Kubernetes Cloud Installer!