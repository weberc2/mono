# Kube Edit Secret

Allows a user to edit a secret in their editor (configurable via the `$EDITOR`
environment variable) in plaintext rather than base64 (the latter is no less
secure and infinitely more cumbersome)

## USAGE

```bash
$ kes --namespace default hello-world-secret
```

**NOTE** `--namespace` is optional; if omitted, the namespace will be fetched
from kubeconfig.