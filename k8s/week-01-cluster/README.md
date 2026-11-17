# Settimana 1: primo cluster

**Concetti:** architettura (control plane, kubelet, etcd), `kubectl`, kind o k3d.
**Teoria:** kubernetes.io → *Concepts → Overview* e *Cluster Architecture*.

- [ ] Installare `kubectl` e [kind](https://kind.sigs.k8s.io/) (oppure k3d)
- [ ] Creare il cluster con `kind-config.yaml` (sotto): 1 control plane e 2 worker, porte 80/443 esposte per l'Ingress della settimana 3
- [ ] `kubectl get nodes -o wide`, `kubectl get pods -A`: riconoscere i componenti di sistema
- [ ] Caricare le immagini nel cluster senza registry: `kind load docker-image linkboard-api:v0.1.0 --name linkboard`
- [ ] Rompere: `docker stop linkboard-worker` e osservare il nodo passare a `NotReady`

**Deliverable:** `kubectl get nodes` mostra 3 nodi `Ready`.

File da creare qui: `kind-config.yaml`.
