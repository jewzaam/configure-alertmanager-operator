package secret

import (
	"context"
	"testing"

	"github.com/openshift/configure-alertmanager-operator/config"
	alertmanager "github.com/openshift/configure-alertmanager-operator/pkg/types"
	"github.com/openshift/configure-alertmanager-operator/pkg/utility"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// readAlertManagerConfig fetches the AlertManager configuration from its default location.
// This is equivalent to `oc get secrets -n openshift-monitoring alertmanager-main`.
// It specifically extracts the .data "alertmanager.yaml" field, and loads it into a resource
// of type Config, enabling it to be marshalled and unmarshalled as needed.
func readAlertManagerConfig(r *ReconcileSecret, request *reconcile.Request) *alertmanager.Config {
	amconfig := &alertmanager.Config{}

	secret := &corev1.Secret{}

	// Define a new objectKey for fetching the alertmanager config.
	objectKey := client.ObjectKey{
		Namespace: request.Namespace,
		Name:      config.SecretNameAlertmanager,
	}

	// Fetch the alertmanager config and load it into an alertmanager.Config struct.
	r.client.Get(context.TODO(), objectKey, secret)
	secretdata := secret.Data["alertmanager.yaml"]
	err := yaml.Unmarshal(secretdata, &amconfig)
	if err != nil {
		panic(err)
	}

	return amconfig
}

// createSecret creates a fake Secret to use in testing.
func createSecret(reconciler *ReconcileSecret, secretname string, secretkey string, secretdata string) {
	newsecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretname,
			Namespace: config.OperatorNamespace,
		},
		Data: map[string][]byte{
			secretkey: []byte(secretdata),
		},
	}
	reconciler.client.Create(context.TODO(), newsecret)
}

// createReconciler creates a fake ReconcileSecret for testing.
func createReconciler() *ReconcileSecret {
	return &ReconcileSecret{
		client: fake.NewFakeClient(),
		scheme: nil,
	}
}

// createNamespace creates a fake `openshift-monitoring` namespace for testing.
func createNamespace(reconciler *ReconcileSecret, t *testing.T) {
	err := reconciler.client.Create(context.TODO(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: config.OperatorNamespace}})
	if err != nil {
		// exit the test if we can't create the namespace. Every test depends on this.
		t.Errorf("Couldn't create the required namespace for the test. Encountered error: %s", err)
		panic("Exiting due to fatal error")
	}
}

// Create the reconcile request for the specified secret.
func createReconcileRequest(reconciler *ReconcileSecret, secretname string) *reconcile.Request {
	return &reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      secretname,
			Namespace: config.OperatorNamespace,
		},
	}
}

// Test_updateAlertManagerConfig tests writing to the Alertmanager config.
func Test_createPagerdutySecret_Create(t *testing.T) {
	pdKey := "asdaidsgadfi9853"
	wdURL := "http://theinterwebs/asdf"

	configExpected := utility.CreateAlertManagerConfig(pdKey, wdURL)

	// prepare environment
	reconciler := createReconciler()
	createNamespace(reconciler, t)
	createSecret(reconciler, config.SecretNamePD, secretKeyPD, pdKey)
	createSecret(reconciler, config.SecretNameDMS, secretKeyDMS, wdURL)

	// reconcile (one event should config everything)
	req := createReconcileRequest(reconciler, "pd-secret")
	reconciler.Reconcile(*req)

	// read config and a copy for comparison
	configActual := readAlertManagerConfig(reconciler, req)

	utility.AssertEquals(t, configExpected, configActual, "Config Deep Comparison")
}

// Test updating the config and making sure it is updated as expected
func Test_createPagerdutySecret_Update(t *testing.T) {
	pdKey := "asdaidsgadfi9853"
	wdURL := "http://theinterwebs/asdf"

	configExpected := utility.CreateAlertManagerConfig(pdKey, wdURL)

	// prepare environment
	reconciler := createReconciler()
	createNamespace(reconciler, t)
	createSecret(reconciler, config.SecretNamePD, secretKeyPD, pdKey)

	// reconcile (one event should config everything)
	req := createReconcileRequest(reconciler, config.SecretNamePD)
	reconciler.Reconcile(*req)

	// verify what we have configured is NOT what we expect at the end (we have updates to do still)
	configActual := readAlertManagerConfig(reconciler, req)
	utility.AssertNotEquals(t, configExpected, configActual, "Config Deep Comparison")

	// update environment
	createSecret(reconciler, config.SecretNameDMS, secretKeyDMS, wdURL)
	req = createReconcileRequest(reconciler, config.SecretNameDMS)
	reconciler.Reconcile(*req)

	// read config and compare
	configActual = readAlertManagerConfig(reconciler, req)

	utility.AssertEquals(t, configExpected, configActual, "Config Deep Comparison")
}

func Test_ReconcileSecrets(t *testing.T) {
	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}
	tests := []struct {
		name        string
		dmsExists   bool
		pdExists    bool
		amExists    bool
		otherExists bool
	}{
		{
			name:        "Test reconcile with NO secrets.",
			dmsExists:   false,
			pdExists:    false,
			amExists:    false,
			otherExists: false,
		},
		{
			name:        "Test reconcile with dms-secret only.",
			dmsExists:   true,
			pdExists:    false,
			amExists:    false,
			otherExists: false,
		},
		{
			name:        "Test reconcile with pd-secret only.",
			dmsExists:   false,
			pdExists:    true,
			amExists:    false,
			otherExists: false,
		},
		{
			name:        "Test reconcile with alertmanager-main only.",
			dmsExists:   false,
			pdExists:    false,
			amExists:    true,
			otherExists: false,
		},
		{
			name:        "Test reconcile with 'other' secret only.",
			dmsExists:   false,
			pdExists:    false,
			amExists:    false,
			otherExists: true,
		},
		{
			name:        "Test reconcile with pd & dms secrets.",
			dmsExists:   true,
			pdExists:    true,
			amExists:    false,
			otherExists: false,
		},
		{
			name:        "Test reconcile with pd & am secrets.",
			dmsExists:   false,
			pdExists:    true,
			amExists:    true,
			otherExists: false,
		},
		{
			name:        "Test reconcile with am & dms secrets.",
			dmsExists:   true,
			pdExists:    false,
			amExists:    true,
			otherExists: false,
		},
		{
			name:        "Test reconcile with pd, dms, and am secrets.",
			dmsExists:   true,
			pdExists:    true,
			amExists:    true,
			otherExists: false,
		},
	}
	for _, tt := range tests {
		reconciler := createReconciler()
		createNamespace(reconciler, t)

		pdKey := ""
		wdURL := ""

		// Create the secrets for this specific test.
		if tt.amExists {
			writeAlertManagerConfig(reconciler, utility.CreateAlertManagerConfig("", ""))
		}
		if tt.dmsExists {
			wdURL = "https://hjklasdf09876"
			createSecret(reconciler, config.SecretNameDMS, secretKeyDMS, wdURL)
		}
		if tt.otherExists {
			createSecret(reconciler, "other", "key", "asdfjkl")
		}
		if tt.pdExists {
			pdKey = "asdfjkl123"
			createSecret(reconciler, config.SecretNamePD, secretKeyPD, pdKey)
		}

		configExpected := utility.CreateAlertManagerConfig(pdKey, wdURL)

		req := createReconcileRequest(reconciler, config.SecretNameAlertmanager)
		reconciler.Reconcile(*req)

		// load the config and check it
		configActual := readAlertManagerConfig(reconciler, req)

		// NOTE compare of the objects will fail when no secrets are created for some reason, so using .String()
		utility.AssertEquals(t, configExpected.String(), configActual.String(), tt.name)
	}
}
