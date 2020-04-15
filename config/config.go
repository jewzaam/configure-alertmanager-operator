// Copyright 2018 RedHat
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

package config

const (
	// OperatorName the name of this operator
	OperatorName string = "configure-alertmanager-operator"

	// OperatorNamespace the namespace the operator is deployed in
	OperatorNamespace string = "openshift-monitoring"

	// SecretNamePD the name of the PagerDuty secret
	SecretNamePD = "pd-secret"

	// SecretNameDMS the name of the Dead Man's Snitch secret
	SecretNameDMS = "dms-secret"

	// SecretNameAlertmanager the name of the AlertManager secret
	SecretNameAlertmanager = "alertmanager-main"

	// ReceiverNull anything routed to "null" receiver does not get routed to PD
	ReceiverNull = "null"

	// ReceiverMakeItWarning anything routed to "make-it-warning" receiver has severity=warning
	ReceiverMakeItWarning = "make-it-warning"

	// ReceiverPagerduty anything routed to "pagerduty" will alert/notify SREP
	ReceiverPagerduty = "pagerduty"

	// ReceiverWatchdog anything going to Dead Man's Snitch (watchdog)
	ReceiverWatchdog = "watchdog"

	// DefaultReceiver the default receiver used by the route used for pagerduty
	DefaultReceiver = ReceiverNull

	// PagerdutyURL global config for alertmanager config PagerdutyURL
	PagerdutyURL = "https://events.pagerduty.com/v2/enqueue"
)
