package utility

import (
	"fmt"
	"testing"

	"github.com/openshift/configure-alertmanager-operator/config"
	alertmanager "github.com/openshift/configure-alertmanager-operator/pkg/types"
)

// utility class to test PD route creation
func verifyPagerdutyRoute(t *testing.T, route *alertmanager.Route) {
	AssertEquals(t, config.DefaultReceiver, route.Receiver, "Reciever Name")
	AssertEquals(t, true, route.Continue, "Continue")
	AssertEquals(t, []string{"alertname", "severity"}, route.GroupByStr, "GroupByStr")
	AssertGte(t, 1, len(route.Routes), "Number of Routes")

	// verify we have the core routes for namespace, ES, and fluentd
	hasNamespace := false
	hasElasticsearch := false
	hasFluentd := false
	for _, route := range route.Routes {
		if route.MatchRE["namespace"] == alertmanager.PDRegex {
			hasNamespace = true
		} else if route.Match["job"] == "fluentd" {
			hasFluentd = true
		} else if route.Match["cluster"] == "elasticsearch" {
			hasElasticsearch = true
		}
	}

	AssertTrue(t, hasNamespace, "No route for MatchRE on namespace")
	AssertTrue(t, hasElasticsearch, "No route for Match on cluster=elasticsearch")
	AssertTrue(t, hasFluentd, "No route for Match on job=fluentd")
}

func verifyNullReceiver(t *testing.T, receivers []*alertmanager.Receiver) {
	hasNull := false
	for _, receiver := range receivers {
		if receiver.Name == config.ReceiverNull {
			hasNull = true
			AssertEquals(t, 0, len(receiver.PagerdutyConfigs), "Empty PagerdutyConfigs")
		}
	}
	AssertTrue(t, hasNull, fmt.Sprintf("No '%s' receiver", config.ReceiverNull))
}

// utility function to verify Pagerduty Receivers
func verifyPagerdutyReceivers(t *testing.T, key string, receivers []*alertmanager.Receiver) {
	// there are at least 3 receivers: namespace, elasticsearch, and fluentd
	AssertGte(t, 2, len(receivers), "Number of Receivers")

	// verify structure of each
	hasMakeItWarning := false
	hasPagerduty := false
	for _, receiver := range receivers {
		switch receiver.Name {
		case config.ReceiverMakeItWarning:
			hasMakeItWarning = true
			AssertEquals(t, true, receiver.PagerdutyConfigs[0].NotifierConfig.VSendResolved, "VSendResolved")
			AssertEquals(t, key, receiver.PagerdutyConfigs[0].RoutingKey, "RoutingKey")
			AssertEquals(t, "warning", receiver.PagerdutyConfigs[0].Severity, "Severity")
		case config.ReceiverPagerduty:
			hasPagerduty = true
			AssertEquals(t, true, receiver.PagerdutyConfigs[0].NotifierConfig.VSendResolved, "VSendResolved")
			AssertEquals(t, key, receiver.PagerdutyConfigs[0].RoutingKey, "RoutingKey")
			AssertTrue(t, receiver.PagerdutyConfigs[0].Severity != "", "Non empty Severity")
			AssertNotEquals(t, "warning", receiver.PagerdutyConfigs[0].Severity, "Severity")
		}
	}

	AssertTrue(t, hasMakeItWarning, fmt.Sprintf("No '%s' receiver", config.ReceiverMakeItWarning))
	AssertTrue(t, hasPagerduty, fmt.Sprintf("No '%s' receiver", config.ReceiverPagerduty))
}

// utility function to verify watchdog route
func verifyWatchdogRoute(t *testing.T, route *alertmanager.Route) {
	AssertEquals(t, config.ReceiverWatchdog, route.Receiver, "Reciever Name")
	AssertEquals(t, "5m", route.RepeatInterval, "Repeat Interval")
	AssertEquals(t, "Watchdog", route.Match["alertname"], "Alert Name")
}

// utility to test watchdog receivers
func verifyWatchdogReceiver(t *testing.T, url string, receivers []*alertmanager.Receiver) {
	// there is 1 receiver
	AssertGte(t, 1, len(receivers), "Number of Receivers")

	// verify structure of each
	hasWatchdog := false
	for _, receiver := range receivers {
		if receiver.Name == config.ReceiverWatchdog {
			hasWatchdog = true
			AssertTrue(t, receiver.WebhookConfigs[0].VSendResolved, "VSendResolved")
			AssertEquals(t, url, receiver.WebhookConfigs[0].URL, "URL")
		}
	}

	AssertTrue(t, hasWatchdog, fmt.Sprintf("No '%s' receiver", config.ReceiverWatchdog))
}
