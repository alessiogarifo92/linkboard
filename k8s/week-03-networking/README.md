# Settimana 3: esporla

**Concetti:** Service (ClusterIP, NodePort), DNS interno (`api.default.svc.cluster.local`), Ingress (ingress-nginx).

- [ ] `api-service.yaml` (ClusterIP sulla porta 8080) e `frontend-deployment.yaml` + `frontend-service.yaml`
- [ ] Installare ingress-nginx per kind
- [ ] `ingress.yaml` per `linkboard.localhost`: `/api`, `/r`, `/version` → Service api; `/` → frontend
- [ ] Il `nginx.conf` del frontend fa proxy verso `http://api:8080`: con l'Ingress non serve più, quindi va semplificato o tenuto e capito perché funziona comunque (DNS del Service `api`)
- [ ] Load balancing: `for i in $(seq 10); do curl -s linkboard.localhost/version; done` e osservare cambiare il nome del pod
- [ ] Rompere: un `targetPort` sbagliato nel Service (`kubectl get endpoints` vuoto)

**Deliverable:** `http://linkboard.localhost` raggiungibile dal browser.
