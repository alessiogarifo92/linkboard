// Package config legge la configurazione dalle variabili d'ambiente
// (in Kubernetes arrivano da ConfigMap e Secret, vedi settimana 4).
package config

import "os"

func Get(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
