# evict pods from busy nodes

# design

- run as cronjob
- list nodes
- filter nodes, busy or idle(by threshold?, <30% idle, >50% busy)
- if balance is not needed, return
- filter pods(evictable) from busy nodes
- evict pods(handle to scheduler), set anti-affinity maybe?
