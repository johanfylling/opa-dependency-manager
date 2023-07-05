# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0]

- Flattened dependency directory structure ([#21](https://github.com/johanfylling/opa-dependency-manager/issues/21))
- Added `list source` command for enumerating source directories
- Added support for listing multiple `source` and `test` directories in `opa.project`

## [0.2.0]

- Configurable OPA executable path through optional `OPA_PATH` environment variable ([#1](https://github.com/johanfylling/opa-dependency-manager/issues/1))
- Refactored `opa.project` yaml structure ([#4](https://github.com/johanfylling/opa-dependency-manager/pull/14))
- Info/debug printing to `stderr` ([#6](https://github.com/johanfylling/opa-dependency-manager/issues/6))
- Added `build` command for building OPA bundles ([#8](https://github.com/johanfylling/opa-dependency-manager/issues/8))
- Respecting `src` and `tests` project parameters for `eval`, `test`, and `build` commands ([#15](https://github.com/johanfylling/opa-dependency-manager/issues/15))
- Override opa executable through `OPA_PATH` environment variable ([#1](https://github.com/johanfylling/opa-dependency-manager/issues/1))

## [0.1.0]

- Initial release
