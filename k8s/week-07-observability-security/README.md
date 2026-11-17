# Settimana 7: osservabilità e sicurezza

**Concetti:** kube-prometheus-stack (Prometheus, Grafana), log, Namespace, ServiceAccount, RBAC, NetworkPolicy, securityContext.

- [ ] Installare kube-prometheus-stack con Helm e aprire Grafana
- [ ] (Estensione al codice) aggiungere `/metrics` all'api con `prometheus/client_golang` e creare un `ServiceMonitor`
- [ ] Spostare l'app nel namespace `linkboard`
- [ ] ServiceAccount per l'app con `automountServiceAccountToken: false`
- [ ] NetworkPolicy default-deny, poi aprire solo: frontend → api, api/worker → postgres e redis
- [ ] Nota: la CNI di default di kind non applica le NetworkPolicy. Serve Calico o Cilium (ricreare il cluster)
- [ ] `securityContext`: `runAsNonRoot`, `readOnlyRootFilesystem` (le immagini distroless lo permettono già)

**Deliverable:** dashboard con le metriche; il frontend non può parlare col DB.
