# Developer Documentation

This section contains comprehensive documentation for developers who want to contribute to, extend, or understand the internals of Waffles.

## Table of Contents

- [Architecture Overview](architecture.md) - System design and component relationships
- [Development Setup](setup.md) - How to set up a development environment
- [Code Organization](code-organization.md) - Project structure and module responsibilities
- [Contributing Guidelines](contributing.md) - How to contribute to the project
- [Testing Guide](testing.md) - Testing strategies and guidelines
- [Build System](build-system.md) - Build, release, and distribution processes
- [API Documentation](api.md) - Internal API reference and interfaces
- [Extension Guide](extensions.md) - How to extend Waffles functionality

## Quick Start for Developers

### Prerequisites
- Go 1.21+
- Git
- Make
- Python 3.8+ (for dependencies)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/toozej/waffles.git
cd waffles

# Install dependencies
make deps

# Build the project
make build

# Run tests
make test

# Install for local development
make install
```

### Making Changes

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Make changes and test
make test-all

# Commit with conventional commits
git commit -m "feat: add new feature description"

# Push and create pull request
git push origin feature/your-feature-name
```

## Project Overview

Waffles is a Go-based CLI tool that orchestrates an LLM toolchain for local development workflows. The architecture follows clean architecture principles with clear separation of concerns.

### Key Components

- **CLI Layer** (`cmd/waffles/`) - Command-line interface and user interaction
- **Internal Services** (`internal/`) - Core business logic and orchestration
- **Package Libraries** (`pkg/`) - Reusable components and utilities
- **Configuration** (`pkg/config/`) - Environment-based configuration management
- **Pipeline** (`pkg/pipeline/`) - Execution orchestration and workflow management
- **Database** (`pkg/logging/`) - Persistent storage and analytics
- **Repository Analysis** (`pkg/repo/`) - Project analysis and file detection

### Development Philosophy

- **Simplicity**: Keep implementations straightforward and maintainable
- **Testability**: Design for easy testing with clear interfaces
- **Configurability**: Make behavior configurable through environment variables
- **Reliability**: Handle errors gracefully and provide helpful feedback
- **Performance**: Optimize for developer productivity over raw performance

## Getting Involved

### Ways to Contribute

1. **Bug Reports**: Report issues with detailed reproduction steps
2. **Feature Requests**: Suggest improvements with clear use cases
3. **Code Contributions**: Submit pull requests for fixes and features
4. **Documentation**: Improve or expand documentation
5. **Testing**: Add test coverage or improve test quality
6. **Review**: Help review pull requests from other contributors

### Community Guidelines

- Follow the [Code of Conduct](../../CODE_OF_CONDUCT.md)
- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages
- Write clear, self-documenting code
- Include tests for new functionality
- Update documentation for user-facing changes

## Resources

- **GitHub Repository**: [github.com/toozej/waffles](https://github.com/toozej/waffles)
- **Issue Tracker**: [GitHub Issues](https://github.com/toozej/waffles/issues)
- **Discussions**: [GitHub Discussions](https://github.com/toozej/waffles/discussions)
- **CI/CD**: GitHub Actions workflows in [`.github/workflows/`](../../.github/workflows/)

## License

Waffles is released under the MIT License. See [LICENSE](../../LICENSE) for details.

---

For detailed information on any topic, see the specific documentation files linked above.