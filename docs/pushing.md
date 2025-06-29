# 📤 Pushing a Package
```bash
upkg push ghcr.io/org/my-package:tag
```

This will:

1. Locate your packaged code (ecosystem specific).

2. Create an OCI artifact with the correct format.

3. Push it to the specified registry reference.

You must authenticate with the registry before pushing.