package utility

import (
	"testing"

	"github.com/openshift/configure-alertmanager-operator/config"
)

func Test_CreatePagerdutyRoute(t *testing.T) {
	// test the structure of the Route is sane
	route := CreatePagerdutyRoute()

	verifyPagerdutyRoute(t, route)
}

func Test_CreatePagerdutyReceivers_WithoutKey(t *testing.T) {
	AssertEquals(t, 0, len(CreatePagerdutyReceivers("")), "Number of Receivers")
}

func Test_CreatePagerdutyReceivers_WithKey(t *testing.T) {
	key := "abcdefg1234567890"

	receivers := CreatePagerdutyReceivers(key)

	verifyPagerdutyReceivers(t, key, receivers)
}

func Test_CreateWatchdogRoute(t *testing.T) {
	// test the structure of the Route is sane
	route := CreateWatchdogRoute()

	verifyWatchdogRoute(t, route)
}

func Test_CreateWatchdogReceivers_WithoutURL(t *testing.T) {
	AssertEquals(t, 0, len(CreateWatchdogReceivers("")), "Number of Receivers")
}

func Test_CreateWatchdogReceivers_WithKey(t *testing.T) {
	url := "http://whatever/something"

	receivers := CreateWatchdogReceivers(url)

	verifyWatchdogReceiver(t, url, receivers)
}

func Test_CreateAlertManagerConfig_WithoutKey_WithoutURL(t *testing.T) {
	pdKey := ""
	wdURL := ""

	cfg := CreateAlertManagerConfig(pdKey, wdURL)

	// verify static things
	AssertEquals(t, "5m", cfg.Global.ResolveTimeout, "Global.ResolveTimeout")
	AssertEquals(t, config.PagerdutyURL, cfg.Global.PagerdutyURL, "Global.PagerdutyURL")
	AssertEquals(t, config.DefaultReceiver, cfg.Route.Receiver, "Route.Receiver")
	AssertEquals(t, "30s", cfg.Route.GroupWait, "Route.GroupWait")
	AssertEquals(t, "5m", cfg.Route.GroupInterval, "Route.GroupInterval")
	AssertEquals(t, "12h", cfg.Route.RepeatInterval, "Route.RepeatInterval")
	AssertEquals(t, 0, len(cfg.Route.Routes), "Route.Routes")
	AssertEquals(t, 1, len(cfg.Receivers), "Receivers")

	verifyNullReceiver(t, cfg.Receivers)
}

func Test_CreateAlertManagerConfig_WithKey_WithoutURL(t *testing.T) {
	pdKey := "poiuqwer78902345"
	wdURL := ""

	cfg := CreateAlertManagerConfig(pdKey, wdURL)

	// verify static things
	AssertEquals(t, "5m", cfg.Global.ResolveTimeout, "Global.ResolveTimeout")
	AssertEquals(t, config.PagerdutyURL, cfg.Global.PagerdutyURL, "Global.PagerdutyURL")
	AssertEquals(t, config.DefaultReceiver, cfg.Route.Receiver, "Route.Receiver")
	AssertEquals(t, "30s", cfg.Route.GroupWait, "Route.GroupWait")
	AssertEquals(t, "5m", cfg.Route.GroupInterval, "Route.GroupInterval")
	AssertEquals(t, "12h", cfg.Route.RepeatInterval, "Route.RepeatInterval")
	AssertEquals(t, 1, len(cfg.Route.Routes), "Route.Routes")
	AssertEquals(t, 3, len(cfg.Receivers), "Receivers")

	verifyNullReceiver(t, cfg.Receivers)

	verifyPagerdutyRoute(t, cfg.Route.Routes[0])
	verifyPagerdutyReceivers(t, pdKey, cfg.Receivers)
}

func Test_CreateAlertManagerConfig_WithKey_WithURL(t *testing.T) {
	pdKey := "poiuqwer78902345"
	wdURL := "http://theinterwebs"

	cfg := CreateAlertManagerConfig(pdKey, wdURL)

	// verify static things
	AssertEquals(t, "5m", cfg.Global.ResolveTimeout, "Global.ResolveTimeout")
	AssertEquals(t, config.PagerdutyURL, cfg.Global.PagerdutyURL, "Global.PagerdutyURL")
	AssertEquals(t, config.DefaultReceiver, cfg.Route.Receiver, "Route.Receiver")
	AssertEquals(t, "30s", cfg.Route.GroupWait, "Route.GroupWait")
	AssertEquals(t, "5m", cfg.Route.GroupInterval, "Route.GroupInterval")
	AssertEquals(t, "12h", cfg.Route.RepeatInterval, "Route.RepeatInterval")
	AssertEquals(t, 2, len(cfg.Route.Routes), "Route.Routes")
	AssertEquals(t, 4, len(cfg.Receivers), "Receivers")

	verifyNullReceiver(t, cfg.Receivers)

	verifyPagerdutyRoute(t, cfg.Route.Routes[0])
	verifyPagerdutyReceivers(t, pdKey, cfg.Receivers)

	verifyWatchdogRoute(t, cfg.Route.Routes[1])
	verifyWatchdogReceiver(t, wdURL, cfg.Receivers)
}

func Test_CreateAlertManagerConfig_WithoutKey_WithURL(t *testing.T) {
	pdKey := ""
	wdURL := "http://theinterwebs"

	cfg := CreateAlertManagerConfig(pdKey, wdURL)

	// verify static things
	AssertEquals(t, "5m", cfg.Global.ResolveTimeout, "Global.ResolveTimeout")
	AssertEquals(t, config.PagerdutyURL, cfg.Global.PagerdutyURL, "Global.PagerdutyURL")
	AssertEquals(t, config.DefaultReceiver, cfg.Route.Receiver, "Route.Receiver")
	AssertEquals(t, "30s", cfg.Route.GroupWait, "Route.GroupWait")
	AssertEquals(t, "5m", cfg.Route.GroupInterval, "Route.GroupInterval")
	AssertEquals(t, "12h", cfg.Route.RepeatInterval, "Route.RepeatInterval")
	AssertEquals(t, 1, len(cfg.Route.Routes), "Route.Routes")
	AssertEquals(t, 2, len(cfg.Receivers), "Receivers")

	verifyNullReceiver(t, cfg.Receivers)
	verifyWatchdogRoute(t, cfg.Route.Routes[0])
	verifyWatchdogReceiver(t, wdURL, cfg.Receivers)
}
