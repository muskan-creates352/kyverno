# AGENTS.md — cmd/cli/kubectl-kyverno/commands/create/user-info

This package implements the `kyverno create user-info` (and `kubectl-kyverno create user-info`) CLI command, which scaffolds a Kyverno `UserInfo` manifest conforming to `cli.kyverno.io/v1alpha1`.

## Command Contract and Semantics

1. **Manifest Structure**:
   - Outputs a standard Kubernetes-style manifest with `apiVersion: cli.kyverno.io/v1alpha1` and `kind: UserInfo`.
   - `metadata.name` defaults to `user-info` and can be overridden via the `-n, --name` flag.
   - Embeds user, group, role, and cluster-role assignments under `userInfo`, `roles`, and `clusterRoles`.

2. **Validation**:
   - The `--name` flag must not be empty. If an empty name is supplied, a validation error is returned.

3. **Failure-Safe Output Writing**:
   - Flag validation and template rendering execute in-memory before writing the output file (`-o/--output`).
   - Output files are written and closed via a temporary file in the destination directory and atomically renamed into place. If validation, write, or close fails, pre-existing output files remain intact and untruncated.
   - Resolved symlink destinations and pre-existing file permissions are preserved intact (dangling symlinks return an error, and newly created files default to restrictive 0600 permissions).
