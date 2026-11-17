# Linkboard

Un URL shortener con statistiche, costruito per **imparare Kubernetes** un concetto alla volta.
Il codice dell'app è volutamente semplice: il lavoro vero sta nella cartella [`k8s/`](k8s/), una settimana per cartella.

Piano di studio completo: pagina Notion *"Kubernetes: piano di studio con progetto Linkboard"*.

## Architettura

```
browser ──► frontend (nginx) ──► api (Go) ──► PostgreSQL   (link e contatori)
                                   │
                                   └─ LPUSH ─► Redis ─ BRPOP ─► worker (Go) ──► PostgreSQL
```

| Servizio | Cosa fa | Porta | Endpoint utili |
|---|---|---|---|
| `api` | Crea e risolve i link, accoda i click | 8080 | `POST /api/links`, `GET /api/links`, `GET /r/{code}`, `/healthz`, `/readyz`, `/version` |
| `worker` | Consuma i click e aggiorna i contatori | 8081 | `/healthz` |
| `frontend` | Dashboard statica (stile minimale) | 80 | `/` |
| `postgres` | Tabella `links` | 5432 | |
| `redis` | Lista `linkboard:clicks` | 6379 | |

Già pronto per Kubernetes: configurazione solo da variabili d'ambiente, liveness e readiness separate, `/version` che mostra quale pod ha risposto, arresto graceful su SIGTERM, immagini distroless non-root.

## Settimana 0: avvio con Docker Compose

Dalla cartella del repo clonato:

```bash
docker compose up --build
```

Apri http://localhost:8000, accorcia un URL, clicca il link: dopo qualche secondo il worker aggiorna il contatore.

## Sviluppo

```bash
make verify             # vet + gofmt + test unitari: da eseguire prima di ogni push
make test-integration   # anche i test su Postgres/Redis veri (vedi Makefile)
make images             # build delle tre immagini Docker
```

| Variabile | Default | Usata da |
|---|---|---|
| `DATABASE_URL` | `postgres://linkboard:linkboard@localhost:5432/linkboard` | api, worker |
| `REDIS_ADDR` | `localhost:6379` | api, worker |
| `PORT` | `8080` | api |
| `HEALTH_PORT` | `8081` | worker |
| `APP_VERSION` | `dev` | api (`/version`) |

## Percorso

| Settimana | Cartella | Obiettivo |
|---|---|---|
| 1 | [`k8s/week-01-cluster`](k8s/week-01-cluster) | Cluster kind locale |
| 2 | [`k8s/week-02-deployment`](k8s/week-02-deployment) | Deployment, rolling update, rollback |
| 3 | [`k8s/week-03-networking`](k8s/week-03-networking) | Service e Ingress |
| 4 | [`k8s/week-04-config-state`](k8s/week-04-config-state) | ConfigMap, Secret, StatefulSet |
| 5 | [`k8s/week-05-reliability`](k8s/week-05-reliability) | Probe, risorse, HPA, CronJob |
| 6 | [`k8s/week-06-helm-gitops`](k8s/week-06-helm-gitops) | Helm e Argo CD |
| 7 | [`k8s/week-07-observability-security`](k8s/week-07-observability-security) | Prometheus, RBAC, NetworkPolicy |
| 8 | [`k8s/week-08-cloud`](k8s/week-08-cloud) | Cluster gestito e TLS (facoltativa) |

Ogni settimana: teoria breve → manifest scritti a mano → rompere qualcosa di proposito → annotare in [`NOTES.md`](NOTES.md).
