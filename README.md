# OPA Dependency Manager (ODM)

ODM is a tool for managing dependencies for [Open Policy Agent](https://www.openpolicyagent.org/) (OPA) projects.

__NOTE__: This is an experimental project not officially supported by the OPA team or Styra. 

```bash
odm init my_project
cd my_project
odm dep git+https://github.com/anderseknert/rego-test-assertions
mkdir src

cat <<EOF > src/policy.rego
package main

import data.test.assert

foo := 42

test_foo {
    assert.equals(42, foo)
}
EOF

odm update
odm test
```

An example project can be found [here](https://github.com/johanfylling/odm-example-project).

## Build

```bash
go build
```

## Run

Where you have your `.rego` project/files.

### Setup new project

```bash
odm init
```

### Add a dependency

```bash
odm dep <dependency path>
```

In `opa.project`:

```yaml
dependencies:
  - path: <dependency path>
```

#### Local dependency

Local dependencies can be specified with relative or absolute paths, or URLs.:

* `file://<path>`
* `./<path>`
* `<path>`
* `../<path>`
* `~/<path>`

#### Git dependency

Git dependencies are URLs prefixed with `git+`:

* `git+http://<path>`
* `git+https://<path>`
* `git+ssh://<path>`

[//]: # (A branch, tag or commit can be specified with the `#` separator:)

[//]: # ()
[//]: # (* `git+https://<path>#<branch>`)

[//]: # (* `git+https://<path>#<tag>`)

[//]: # (* `git+https://<path>#<commit>`)

#### Namespacing

```bash
odm dep path/to/dependency -n mynamespace
```

In `opa.project`:

```yaml
dependencies:
  - path: path/to/dependency
    namespace: mynamespace
```

When a dependency is namespaced, all contained Rego packages will be prefixed with the namespace.
E.g.: a dependency with the following package structure:

```
foo
 +-- bar
 |   +-- baz
 +-- qux   
```

when namespaced with `utils`, it will have the following structure:

```
utils
 +-- foo
     +-- bar
     |   +-- baz
     +-- qux   
```

Transitive dependencies will be namespaced as well.
Any transitive dependency already namespaced by its enclosing dependency project will have it's packages prefixed by the namespace assigned by the enclosing project, and then by the namespace defined in the main project, recursively.

### Update dependencies

```bash
odm update
```

### Evaluating policies

Example:
```bash
odm eval -- 'data.main.allow'
```

if a `source` folder is specified in `opa.project`, it will be automatically included in the evaluation.

### Testing policies

Example:
```bash
odm test -- -d policy.rego
```

if a `source` folder is specified in `opa.project`, it will be automatically included in the evaluation.

## The `opa.project` file

The `opa.project` file is a YAML file that contains the project configuration.

Example:

```yaml
name: <project name>
source: <source path>
dependencies:
  - path: <dependency path>
    namespace: <namespace>
```

### `name`

The name of the project.

### `source`

The path to the source folder.
If specified, the source folder will be automatically included in the `eval` and `test` commands.

### `dependencies`

A list of dependencies.

#### `path`

The path to the dependency.

#### `namespace`

The namespace to use for the dependency.
