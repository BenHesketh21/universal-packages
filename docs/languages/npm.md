# NPM

Universal Packages currently supports pushing and pulling npm packages via `.tgz` tarballs.

---

## 🧪 Assumptions

- Uses standard `npm pack` output: `<name>-<version>.tgz`.
- Assumes the project contains a valid `package.json`.
- Pull places the `.tgz` file in a `vendor/` directory.

---

## 📥 Installing (Pull)

1. Pulls the `.tgz` from the OCI registry.
2. Adds a `file:` reference in your `package.json`, like so:

```json
"dependencies": {
  "your-package": "file:vendor/your-package-1.0.0.tgz"
}
```

You must run npm install manually to link the package.

## 📤 Publishing (Push)
You must first run:

```bash
npm pack
```

The CLI finds the resulting .tgz in the current directory.

It pushes the file as an OCI artifact with your given tag.

✅ Best Practices
Use .npmignore or the files field in package.json to control what’s included in the tarball.

Inspect the output of npm pack before publishing.

Match tag naming conventions to your registry/org layout.