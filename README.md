# OPA Dependency Manager (ODM)

ODM is a tool for managing dependencies for [Open Policy Agent](https://www.openpolicyagent.org/) (OPA) projects.

__NOTE__: This is an experimental project not officially supported by the OPA team or Styra. 

```bash
$ odm init my_project
$ cd my_project
$ odm depend --no-namespace rego-test-assertions \
      git+https://github.com/anderseknert/rego-test-assertions
$ mkdir src

$ cat <<EOF > src/policy.rego
package main

import data.test.assert

foo := 42

test_foo {
    assert.equals(42, foo)
}
EOF

$ odm test
```

An example project can be found [here](https://github.com/johanfylling/odm-example-project).

## Running

Where you have your `.rego` project/files.

### Setup new project

```bash
$ odm init [project name]
```

### Add a dependency

```bash
$ odm depend <dependency name> <dependency path>
```

In `opa.project`:

```yaml
dependencies:
  <dependency name>: <dependency path>
```

#### Local dependency

Local dependencies can be specified with relative or absolute paths, or URLs.:

* `file:/<path>`

Examples:

* Absolute path: `file://tmp/my/dependency`
* Relative path: `file:/../my/dependency`

#### Git dependency

Git dependencies are URLs prefixed with `git+`:

* `git+http://<path>[#tag|branch|commit]]`
* `git+https://<path>[#tag|branch|commit]]`
* `git+ssh://<path>[#tag|branch|commit]]`

Examples:

* GitHub dependency at `HEAD` of repo: `git+https://github.com/johanfylling/odm-example-dependency.git`
* GitHub dependency at `v1.0` tag: `git+https://github.com/johanfylling/odm-example-dependency.git#v1.0`
* GitHub dependency at `foo` branch: `git+https://github.com/johanfylling/odm-example-dependency.git#foo`
* GitHub dependency at `88c5cde` commit: `git+https://github.com/johanfylling/odm-example-dependency.git#88c5cde`

### Update dependencies

```bash
$ odm update
```

### Evaluating policies

Example:
```bash
$ odm eval -- 'data.main.allow'
```

if a `source` folder is specified in `opa.project`, it will be automatically included in the evaluation.

### Testing policies

Example:
```bash
$ odm test -- -d policy.rego
```

if a `source` folder is specified in `opa.project`, it will be automatically included in the evaluation.

## Namespacing

By default, dependencies are namespaced by their declared name.

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
Any transitive dependency already namespaced by its enclosing dependency project will have its packages prefixed by the namespace assigned by the enclosing project, and then by the namespace defined in the main project, recursively.

### Custom namespace

```bash
$ odm dep my_dep file:/path/to/dependency -n mynamespace
```

In `opa.project`:

```yaml
dependencies:
  my_dep: 
    path: file:/path/to/dependency
    namespace: mynamespace
```

### Disabling namespacing

```bash
$ odm dep my_dep file:/path/to/dependency --no-namespace
```

In `opa.project`:

```yaml
dependencies:
  my_dep: 
    path: file:/path/to/dependency
    namespace: false
```

## The `opa.project` file

The `opa.project` file is a YAML file that contains the project configuration.

Example:

```yaml
name: <project name>
source: <source path>
dependencies:
  <dependency name>: <dependency path>
```

### Attributes

| Attribute                       | Type                 | Default                 | Description                                                                                                                                                                                                 |
|---------------------------------|----------------------|-------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`                          | `string`             | none                    | The name of the project.                                                                                                                                                                                    |
| `source`                        | `string`, `[]string` | none                    | The path to the source folder. If specified, the source directory will be automatically included in the `eval` and `test` commands. Can either be the path of a single directory, or a list of directories. |
| `tests`                         | `string`, `[]string` | none                    | The path to the test folder. If specified, the test directory will be automatically included in the `test` command. Can either be the path of a single directory, or a list of directories.                 |
| `dependencies`                  | `map`                |                         | A map of dependency declaration, keyed by their name.                                                                                                                                                       |
| `dependencies.<name>`           | `map`, `string`      | none                    | A dependency declaration. A short form is supported, where the dependency value is its location as a string.                                                                                                |
| `dependencies.<name>.location`  | `string`             | none                    | The location of the dependency.                                                                                                                                                                             |
| `dependencies.<name>.namespace` | `string`, `bool`     | `true`                  | If a `string`: the namespace to use for the dependency.  If a `bool`: if `true`, use the dependency `name` as namespace; if `false`, don't namesapace the dependency.                                       |
| `build`                         | `map`                |                         | Settings for building bundles.                                                                                                                                                                              |
| `build.output`                  | `string`             | `./build/bundle.tar.gz` | The location of the target bundle.                                                                                                                                                                          |
| `build.target`                  | `string`             | `rego`                  | The target bundle format. E.g. `rego`, `wasm`, or `plan`                                                                                                                                                    |
| `build.entrypoints`             | `[]string`           | `[]`                    | List of entrypoints.                                                                                                                                                                                        |
