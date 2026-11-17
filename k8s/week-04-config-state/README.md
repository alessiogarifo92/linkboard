# Settimana 4: configurazione e stato

**Concetti:** ConfigMap, Secret, PersistentVolumeClaim, StatefulSet, Headless Service.

- [ ] `configmap.yaml`: `REDIS_ADDR`, `APP_VERSION`
- [ ] `secret.yaml`: `DATABASE_URL` e password di Postgres (mai committare valori reali: per lo studio va bene, ma va annotato in NOTES.md perché in produzione no)
- [ ] `postgres-statefulset.yaml` con `volumeClaimTemplates` e un Service headless
- [ ] `redis-deployment.yaml` + Service (Redis qui è una cache o una coda: perdere i dati è accettabile? Rispondere in NOTES.md)
- [ ] `worker-deployment.yaml`
- [ ] Rompere: `kubectl delete pod postgres-0` e verificare che i link ci siano ancora

**Deliverable:** stack completo nel cluster, DB persistente.
