## Development

Go 1.9+ required

## Dependencies

```bash
  dep ensure -vendor-only -v
```

### Setup

```bash
minikube start --vm-driver xhyve --insecure-registry your.registry
go install github.com/codegangsta/gin
```

> If you already had a VM ensure to run `minikube delete`

### Run

```bash
make run
```
### Development

Environment varialbes:
```
# if you will generate toolbelt boxes
export AWS_ACCESS_KEY=<aws-access-key>
export AWS_SECRET_KEY=<aws-secret-key>

# was used to pull images from xumak registry (for minikube)
DEVELOPMENT=false

# vault config
export VAULT_ADDR=https://127.0.0.1:8200
export VAULT_TOKEN=TOKEN-HERE
export VAULT_SKIP_VERIFY=true

# grid DNS
export GRID_EXTERNAL_DOMAIN=test.grid.xumak.io

# ingress controller config
export INGRESS_CLASS=contour
export INGRESS_FORCE_SSL_REDIRECT=true

# certManager and issuer config
export CERT_MANAGER_ISSUER=letsencrypt-prod-dns
export CERT_MANAGER_DNS_PROVIDER=prod-dns
```

To have Vault in the localhost
```
# to get de active pod
kubectl -n bedrock get vault grid-vault -o jsonpath='{.status.vaultStatus.active}'
kubectl -n bedrock port-forward <active-pod> 8200
```

Copyright © 2016 Tikal Technologies, Inc.