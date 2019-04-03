# evict pods from busy nodes

# design

- run as cronjob
- list nodes
- filter nodes, busy or idle(by threshold?, <30% idle, >50% busy)
- if balance is not needed, return
- filter pods(evictable) from busy nodes
- evict pods(handle to scheduler)

# quick start
```bash
make img
# push img or use stepdc/podacrobat:latest
kubectl apply -f hack/k8s/rbac.yaml
kubectl apply -f hack/k8s/cronjob.yaml
```
