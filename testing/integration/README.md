Auto-regression testcase
---

## Pre-requirement

- `kubectl` installed and configured to a test cluster
- `ktctl` compiled and installed to a folder in PATH variable
- `sudo` command is available without password

## Usage

```bash
testing/integration/go.sh [options]
```

**Available options**

- `--keep-proof` do not clean up resource in cluster when failed
- `--cleanup-only` only clean up resource in cluster created by test
