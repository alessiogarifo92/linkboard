# Decisioni di progetto

Registro delle decisioni prese con l'utente (project-lifecycle). Le modifiche si fanno qui, non ri-decidendole nei ticket.

| Data | Fase | Decisione | Perché |
|---|---|---|---|
| 2026-11-17 | 1. Idea | Progetto di studio Kubernetes: URL shortener "Linkboard" con api, worker, frontend, Postgres, Redis | Abbastanza piccolo da scrivere in pochi giorni, con tutti i pezzi di un sistema reale (stateless, stateful, coda, job) |
| 2026-11-17 | 2. Stack e costi | Go 1.24 (stdlib `net/http`, pgx v5, go-redis v9), nginx per il frontend, PostgreSQL 17, Redis 7. Cluster locale con kind | Immagini piccole e avvio istantaneo rendono visibili rolling update e probe. **Costo: 0 €** fino alla settimana 8 (cloud facoltativo, da distruggere a fine prova) |
| 2026-11-17 | 3. Architettura | api stateless + coda Redis (LPUSH/BRPOP) + worker; lo schema DB lo crea l'api all'avvio | Separare il redirect (veloce) dal conteggio (asincrono) dà un worker da scalare. Le migrazioni come Job sono un esercizio della settimana 5 |
| 2026-11-17 | 3.5. Design UI | **Rimandata**: stile minimale di default | Il frontend è solo una dashboard di studio; l'obiettivo è Kubernetes, non la UI |
| 2026-11-17 | R. Repo | `alessiogarifo92/linkboard`, pubblico | Minuti di GitHub Actions illimitati; utile come portfolio |
