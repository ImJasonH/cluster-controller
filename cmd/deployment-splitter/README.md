Running the deployment splitter outside the cluster:

```
SYSTEM_NAMESPACE=default go run ./cmd/deployment-splitter \
  --kubeconfig=$GOPATH/src/github.com/smarterclayton/kcp/.kcp/data/admin.kubeconfig
```

- `SYSTEM_NAMESPACE` is required by knative.dev
- `--kubeconfig=` is required to connect to `kcp` using kubeconfig
