# 📥 Pulling a Package

```bash
upkg install ghcr.io/org/my-package:tag
```

This will:

1. Pull an OCI artifact from the registry.

2. Extract the contents into a known local format.

3. Update your local project to reference the pulled package (ecosystem-specific).

See [ecosystems](./ecosystems) for ecosystem-specific integration details.