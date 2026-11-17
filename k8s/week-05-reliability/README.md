# Settimana 5: affidabilità

**Concetti:** probe liveness/readiness/startup, requests/limits, HPA, metrics-server, Job/CronJob, terminazione graceful.

- [ ] Probe sull'api: liveness `GET /healthz`, readiness `GET /readyz` (controlla Postgres e Redis)
- [ ] Probe sul worker: liveness `GET :8081/healthz`
- [ ] requests/limits su tutti i container
- [ ] Installare metrics-server e creare un HPA sull'api (CPU al 60%)
- [ ] Test di carico: `hey -z 60s -c 50 http://linkboard.localhost/r/<code>` oppure k6
- [ ] `cleanup-cronjob.yaml`: ogni notte cancella i link più vecchi di 30 giorni (`psql` in un'immagine postgres)
- [ ] Terminazione: il worker si ferma solo quando `BRPOP` scade (fino a 5 s dopo SIGTERM). Misurarlo con `kubectl delete pod` e confrontarlo con `terminationGracePeriodSeconds`
- [ ] Rompere: liveness su `/readyz` e poi spegnere Redis. Perché si riavviano tutti i pod dell'api?

**Deliverable:** l'api scala sotto carico, i pod non pronti escono dal Service.
