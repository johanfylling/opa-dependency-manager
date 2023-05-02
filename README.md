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

Without namespacing:

```bash
odm dep path/to/dependency
```

With namespacing:

```bash
odm dep path/to/dependency -n mynamespace
```

### Update dependencies

```bash
odm update
```