# Repositories

A repository is a simple mapper of name to library location; 
where the name (preferably human-readable) can be used in a Rego project's `opa.project` file to link to a library.

## The `repository.yaml` file

The `repository.yaml` file is a simple YAML file that lists all curated libraries.

| Attribute          | Type     | Description                      |
|--------------------|----------|----------------------------------|
| `libraries.<name>` | `string` | The loication of the dependency. |

### Example

```yaml
libraries:
  test-assertions: git+https://github.com/anderseknert/rego-test-assertions
  example: git+https://github.com/johanfylling/odm-example-project
```