package utility

import (
	"github.com/openshift/configure-alertmanager-operator/config"
	alertmanager "github.com/openshift/configure-alertmanager-operator/pkg/types"
)

// CreatePagerdutyRoute creates an AlertManager Route for PagerDuty in memory.
func CreatePagerdutyRoute() *alertmanager.Route {
	// order matters.
	// these are sub-routes.  if any matches it will not continue processing.
	// 1. route anything we want to silence to "null"
	// 2. route anything that should be a warning to "make-it-warning"
	// 3. route anything we want to go to PD
	pagerdutySubroutes := []*alertmanager.Route{
		// https://issues.redhat.com/browse/OSD-1966
		{Receiver: config.ReceiverNull, Match: map[string]string{"alertname": "KubeQuotaExceeded"}},
		// https://issues.redhat.com/browse/OSD-2382
		{Receiver: config.ReceiverNull, Match: map[string]string{"alertname": "UsingDeprecatedAPIAppsV1Beta1"}},
		// https://issues.redhat.com/browse/OSD-2382
		{Receiver: config.ReceiverNull, Match: map[string]string{"alertname": "UsingDeprecatedAPIAppsV1Beta2"}},
		// https://issues.redhat.com/browse/OSD-2382
		{Receiver: config.ReceiverNull, Match: map[string]string{"alertname": "UsingDeprecatedAPIExtensionsV1Beta1"}},
		// https://issues.redhat.com/browse/OSD-2980
		{Receiver: config.ReceiverNull, Match: map[string]string{"alertname": "CPUThrottlingHigh", "container": "registry-server"}},
		// https://issues.redhat.com/browse/OSD-3008
		{Receiver: config.ReceiverNull, Match: map[string]string{"alertname": "CPUThrottlingHigh", "container": "configmap-registry-server"}},
		// https://issues.redhat.com/browse/OSD-3010
		{Receiver: config.ReceiverNull, Match: map[string]string{"alertname": "NodeFilesystemSpaceFillingUp", "severity": "warning"}},
		// https://issues.redhat.com/browse/OSD-2611
		{Receiver: config.ReceiverNull, Match: map[string]string{"namespace": "openshift-customer-monitoring"}},
		// https://issues.redhat.com/browse/OSD-3220
		{Receiver: config.ReceiverNull, Match: map[string]string{"alertname": "SLAUptimeSRE"}},

		// https://issues.redhat.com/browse/OSD-1922
		{Receiver: config.ReceiverMakeItWarning, Match: map[string]string{"alertname": "KubeAPILatencyHigh", "severity": "critical"}},

		// general: route anything in core namespaces to PD
		{Receiver: config.ReceiverPagerduty, MatchRE: map[string]string{"namespace": alertmanager.PDRegex}},
		// fluentd: route any fluentd alert to PD
		// https://issues.redhat.com/browse/OSD-3326
		{Receiver: config.ReceiverPagerduty, Match: map[string]string{"job": "fluentd"}},
		{Receiver: config.ReceiverPagerduty, Match: map[string]string{"alertname": "FluentdNodeDown"}},
		// elasticsearch: route any ES alert to PD
		// https://issues.redhat.com/browse/OSD-3326
		{Receiver: config.ReceiverPagerduty, Match: map[string]string{"cluster": "elasticsearch"}},
	}

	return &alertmanager.Route{
		Receiver: config.DefaultReceiver,
		GroupByStr: []string{
			"alertname",
			"severity",
		},
		Continue: true,
		Routes:   pagerdutySubroutes,
	}
}

// CreatePagerdutyConfig creates an AlertManager PagerdutyConfig for PagerDuty in memory.
func CreatePagerdutyConfig(pagerdutyRoutingKey string) *alertmanager.PagerdutyConfig {
	return &alertmanager.PagerdutyConfig{
		NotifierConfig: alertmanager.NotifierConfig{VSendResolved: true},
		RoutingKey:     pagerdutyRoutingKey,
		Severity:       `{{ if .CommonLabels.severity }}{{ .CommonLabels.severity | toLower }}{{ else }}critical{{ end }}`,
		Description:    `{{ .CommonLabels.alertname }} {{ .CommonLabels.severity | toUpper }} ({{ len .Alerts }})`,
		Details: map[string]string{
			"link":         `{{ if .CommonAnnotations.link }}{{ .CommonAnnotations.link }}{{ else }}https://github.com/openshift/ops-sop/tree/master/v4/alerts/{{ .CommonLabels.alertname }}.md{{ end }}`,
			"link2":        `{{ if .CommonAnnotations.runbook }}{{ .CommonAnnotations.runbook }}{{ else }}{{ end }}`,
			"group":        `{{ .CommonLabels.alertname }}`,
			"component":    `{{ .CommonLabels.alertname }}`,
			"num_firing":   `{{ .Alerts.Firing | len }}`,
			"num_resolved": `{{ .Alerts.Resolved | len }}`,
			"resolved":     `{{ template "pagerduty.default.instances" .Alerts.Resolved }}`,
		},
	}

}

// CreatePagerdutyReceivers creates an AlertManager Receiver for PagerDuty in memory.
func CreatePagerdutyReceivers(pagerdutyRoutingKey string) []*alertmanager.Receiver {
	if pagerdutyRoutingKey == "" {
		return []*alertmanager.Receiver{}
	}

	receivers := []*alertmanager.Receiver{
		{
			Name:             config.ReceiverPagerduty,
			PagerdutyConfigs: []*alertmanager.PagerdutyConfig{CreatePagerdutyConfig(pagerdutyRoutingKey)},
		},
	}

	// make-it-warning overrides the severity
	pdconfig := CreatePagerdutyConfig(pagerdutyRoutingKey)
	pdconfig.Severity = "warning"
	receivers = append(receivers, &alertmanager.Receiver{
		Name:             config.ReceiverMakeItWarning,
		PagerdutyConfigs: []*alertmanager.PagerdutyConfig{pdconfig},
	})

	return receivers
}

// CreateWatchdogRoute creates an AlertManager Route for Watchdog (Dead Man's Snitch) in memory.
func CreateWatchdogRoute() *alertmanager.Route {
	return &alertmanager.Route{
		Receiver:       config.ReceiverWatchdog,
		RepeatInterval: "5m",
		Match:          map[string]string{"alertname": "Watchdog"},
	}
}

// CreateWatchdogReceivers creates an AlertManager Receiver for Watchdog (Dead Man's Sntich) in memory.
func CreateWatchdogReceivers(watchdogURL string) []*alertmanager.Receiver {
	if watchdogURL == "" {
		return []*alertmanager.Receiver{}
	}

	snitchconfig := &alertmanager.WebhookConfig{
		NotifierConfig: alertmanager.NotifierConfig{VSendResolved: true},
		URL:            watchdogURL,
	}

	return []*alertmanager.Receiver{
		{
			Name:           config.ReceiverWatchdog,
			WebhookConfigs: []*alertmanager.WebhookConfig{snitchconfig},
		},
	}
}

// CreateAlertManagerConfig creates an AlertManager Config in memory based on the provided input parameters.
func CreateAlertManagerConfig(pagerdutyRoutingKey string, watchdogURL string) *alertmanager.Config {
	routes := []*alertmanager.Route{}
	receivers := []*alertmanager.Receiver{}

	if pagerdutyRoutingKey != "" {
		routes = append(routes, CreatePagerdutyRoute())
		receivers = append(receivers, CreatePagerdutyReceivers(pagerdutyRoutingKey)...)
	}

	if watchdogURL != "" {
		routes = append(routes, CreateWatchdogRoute())
		receivers = append(receivers, CreateWatchdogReceivers(watchdogURL)...)
	}

	// always have the "null" receiver
	receivers = append(receivers, &alertmanager.Receiver{Name: config.ReceiverNull})

	amconfig := &alertmanager.Config{
		Global: &alertmanager.GlobalConfig{
			ResolveTimeout: "5m",
			PagerdutyURL:   config.PagerdutyURL,
		},
		Route: &alertmanager.Route{
			Receiver: config.DefaultReceiver,
			GroupByStr: []string{
				"job",
			},
			GroupWait:      "30s",
			GroupInterval:  "5m",
			RepeatInterval: "12h",
			Routes:         routes,
		},
		Receivers: receivers,
		Templates: []string{},
	}

	return amconfig
}
