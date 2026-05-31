# ROOT/functional/logic

Functional logic specifications — behavior, algorithms,
interfaces, and error conditions for each component.

# Public

## Namespaces

Some modules define records that are used by other
modules (e.g. `Frontmatter`, `FormatError`, `Chain`).
To make clear that a type is declared externally and
where it is declared, modules that export such records
declare a namespace:

    namespace: frontmatter

When a module references a record from another module,
it qualifies the name with the source namespace:

    format_errors: list of spectreevalidate.FormatError

This makes explicit that `FormatError` is not defined
in the current module — it comes from the module with
namespace `spectreevalidate`.

Records defined in the current module are used without
qualifier:

    staleness: list of StalenessEntry

Only modules that export records consumed by other
modules need a namespace declaration.
