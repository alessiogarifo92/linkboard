# Settimana 2: deploy dell'api

**Concetti:** Pod, ReplicaSet, Deployment, label/selector, rolling update, rollback.

- [ ] Scrivere a mano `api-deployment.yaml`: 3 repliche, immagine `linkboard-api:v0.1.0`, env `APP_VERSION`
- [ ] Per ora Postgres e Redis non esistono nel cluster: l'api va in crash loop. Osservarlo con `kubectl get pods -w` e `kubectl logs` (la soluzione arriva nella settimana 4; intanto si può puntare `DATABASE_URL` a un Postgres esterno)
- [ ] Rolling update: build con `APP_VERSION=v0.2.0`, `kubectl set image`, `kubectl rollout status`
- [ ] Rollback: `kubectl rollout undo` e `kubectl rollout history`
- [ ] Rompere: un selector che non combacia con le label del template; un'immagine inesistente (`ImagePullBackOff`)

**Deliverable:** 3 repliche, aggiornamento e rollback senza downtime.
