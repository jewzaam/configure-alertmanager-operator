// Copyright 2019 RedHat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"net/http"

	"github.com/openshift/configure-alertmanager-operator/config"
	alertmanager "github.com/openshift/configure-alertmanager-operator/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
)

const (
	// MetricsEndpoint is the port to export metrics on
	MetricsEndpoint = ":8080"
)

var (
	metricPDSecretExists = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "camo_secret_exists_pd",
		Help: "Pager Duty secret exists",
	}, []string{"name"})
	metricDMSSecretExists = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "camo_secret_exists_dms",
		Help: "Dead Man's Snitch secret exists",
	}, []string{"name"})
	metricAMSecretExists = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "camo_secret_exists_am",
		Help: "AlertManager Config secret exists",
	}, []string{"name"})
	metricAMSecretContainsPD = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "camo_secret_configured_pd",
		Help: "AlertManager Config contains configuration for Pager Duty",
	}, []string{"name"})
	metricAMSecretContainsDMS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "camo_secret_configured_dms",
		Help: "AlertManager Config contains configuration for Dead Man's Snitch",
	}, []string{"name"})

	metricsList = []prometheus.Collector{
		metricPDSecretExists,
		metricDMSSecretExists,
		metricAMSecretExists,
		metricAMSecretContainsPD,
		metricAMSecretContainsDMS,
	}
)

// StartMetrics register metrics and exposes them
func StartMetrics() {

	// Register metrics and start serving them on /metrics endpoint
	RegisterMetrics()
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(MetricsEndpoint, nil)
}

// RegisterMetrics for the operator
func RegisterMetrics() error {
	for _, metric := range metricsList {
		err := prometheus.Register(metric)
		if err != nil {
			return err
		}
	}
	return nil
}

// secretExists check if a secret exists in the list of secrets.  Returns 0 if it doesn't exist, 1 if it does exist.
func secretExists(secretname string, list *corev1.SecretList) int {
	for _, secret := range list.Items {
		if secret.Name == secretname {
			return 1
		}
	}

	return 0
}

// secretConfigured check if a secret is configured.  Returns 0 if it isn't configured, 1 if it is configured
func secretConfigured(secretname string, cfg *alertmanager.Config) int {
	for _, receiver := range cfg.Receivers {
		if receiver.Name == config.ReceiverPagerduty && secretname == config.SecretNamePD {
			return 1
		}

		if receiver.Name == config.ReceiverWatchdog && secretname == config.SecretNameDMS {
			return 1
		}
	}

	return 0
}

// UpdateSecretsMetrics updates all metrics related to the existance and contents of Secrets
// used by configure-alertmanager-operator.
func UpdateSecretsMetrics(list *corev1.SecretList, cfg *alertmanager.Config) {
	// does the PD secret exist?
	metricPDSecretExists.With(prometheus.Labels{"name": config.OperatorName}).Set(float64(secretExists(config.SecretNamePD, list)))

	// does the DMS secret exist?
	metricDMSSecretExists.With(prometheus.Labels{"name": config.OperatorName}).Set(float64(secretExists(config.SecretNameDMS, list)))

	// does the AM secret exist?
	metricAMSecretExists.With(prometheus.Labels{"name": config.OperatorName}).Set(float64(secretExists(config.SecretNameAlertmanager, list)))

	// is the PD secret configured?
	metricAMSecretContainsPD.With(prometheus.Labels{"name": config.OperatorName}).Set(float64(secretConfigured(config.SecretNamePD, cfg)))

	// is the DMS secret configured?
	metricAMSecretContainsDMS.With(prometheus.Labels{"name": config.OperatorName}).Set(float64(secretConfigured(config.SecretNameDMS, cfg)))
}
