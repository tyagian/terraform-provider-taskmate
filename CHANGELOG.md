# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of TaskMate Terraform Provider
- Full CRUD operations for tasks
- Data sources for querying single and multiple tasks
- Import support for existing tasks
- Token-based authentication
- Comprehensive examples and documentation
- CI/CD workflows for testing and releases
- Production-ready configuration files
- Documentation generation with tfplugindocs

### Features
- `taskmate_task` resource for managing tasks
- `taskmate_task` data source for querying a single task
- `taskmate_tasks` data source for listing all tasks
- Environment variable support for configuration
- Sensitive token handling
- Import discovery tools

## [1.0.0] - TBD

Initial release.

### Added
- Task resource with full CRUD operations
- Task data sources (single and list)
- Token-based authentication
- Import functionality
- Comprehensive examples
- Documentation
- CI/CD workflows
- Makefile for common tasks
- Linting configuration
- Release automation with GoReleaser

[Unreleased]: https://github.com/YOUR_USERNAME/terraform-provider-taskmate/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/YOUR_USERNAME/terraform-provider-taskmate/releases/tag/v1.0.0
