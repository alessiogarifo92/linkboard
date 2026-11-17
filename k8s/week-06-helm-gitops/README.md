# Settimana 6: packaging e GitOps

**Concetti:** Helm (chart, values, template) oppure Kustomize (base + overlay), Argo CD.

- [ ] Trasformare i manifest delle settimane 2-5 in un chart in `k8s/week-06-helm-gitops/chart/`
- [ ] `values.yaml` con repliche, immagine, risorse; `values-dev.yaml` per kind
- [ ] `helm template` e `helm lint` prima di ogni install
- [ ] Installare Argo CD e creare un'`Application` che punta a questo repo
- [ ] Pubblicare le immagini su GHCR con un workflow (estende `.github/workflows/ci.yml`)
- [ ] Rompere: modificare a mano una risorsa con `kubectl edit` e osservare Argo CD riportarla allo stato del repo (self-heal)

**Deliverable:** un `git push` aggiorna il cluster da solo.
