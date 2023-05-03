# OPA Dependency Managed (ODM)

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

#### Without namespacing

```bash
odm dep path/to/dependency
```

#### With namespacing

```bash
odm dep path/to/dependency -n mynamespace
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

E.g.:
```bash
odm eval -- -d policy.rego 'data.main.allow'
```
