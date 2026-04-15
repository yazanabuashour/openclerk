# OpenClerk Agent Guide

## When to use OpenClerk

Use OpenClerk when an agent needs to store, search, and retrieve local user data without running a daemon or depending on an external service.

Prefer these backends:

- `fts` for exact or lexical search over canonical markdown documents
- `hybrid` when local vector-style scoring improves recall and a simple local embedding provider is enough
- `graph` when document-to-document evidence links matter
- `records` when the agent needs stable entity lookup with facts and citations

## Preferred runtime

Use [`client/local`](client/local) as the default runtime entrypoint. It opens the SQLite-backed store in process and returns the generated backend client without binding a port.

The OpenAPI contract in [`openapi/v1/openclerk.yaml`](openapi/v1/openclerk.yaml) remains the source of truth for operations, schemas, and generated request and response types.

## Default storage

Unless the caller overrides the paths in [`client/local.Config`](client/local/local.go), data is stored under:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk
```

The embedded runtime creates:

- `openclerk.sqlite`
- `vault/`

## Minimal flow

```go
client, runtime, err := local.OpenRecords(local.Config{})
if err != nil {
	return err
}
defer runtime.Close()

create, err := client.CreateDocumentWithResponse(ctx, records.CreateDocumentRequest{
	Path:  "health/labs/glucose.md",
	Title: "Fasting glucose",
	Body:  "---\nentity_type: lab_result\nentity_name: Fasting glucose\nentity_id: fasting-glucose\n---\n# Fasting glucose\n\n## Facts\n- value: 92 mg/dL\n",
})
if err != nil {
	return err
}
if create.JSON201 == nil {
	return fmt.Errorf("create failed: %s", string(create.Body))
}

entity, err := client.GetRecordEntityWithResponse(ctx, "fasting-glucose")
if err != nil {
	return err
}
if entity.JSON200 == nil {
	return fmt.Errorf("lookup failed: %s", string(entity.Body))
}
```

## Practical defaults

- Use explicit `DataDir` overrides in tests and demos to avoid polluting the default XDG location.
- Use release tags for reproducible installs.
- Treat `cmd/openclerkd` as internal compatibility infrastructure, not the primary runtime path.
