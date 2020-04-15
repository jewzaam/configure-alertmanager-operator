package secret

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/openshift/configure-alertmanager-operator/config"
	"github.com/openshift/configure-alertmanager-operator/pkg/metrics"
	alertmanager "github.com/openshift/configure-alertmanager-operator/pkg/types"
	"github.com/openshift/configure-alertmanager-operator/pkg/utility"

	yaml "gopkg.in/yaml.v2"
)

var log = logf.Log.WithName("secret_controller")

var (
	secretKeyPD = "PAGERDUTY_KEY"

	secretKeyDMS = "SNITCH_URL"
)

// Add creates a new Secret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSecret{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("secret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource (type "Secret").
	// For each Add/Update/Delete event, the reconcile loop will be sent a reconcile Request.
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileSecret{}

// ReconcileSecret reconciles a Secret object
type ReconcileSecret struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Secret object and makes changes based on the state read.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling Secret")

	// This operator is only interested in the 3 secrets listed below. Skip reconciling for all other secrets.
	switch request.Name {
	case config.SecretNamePD:
	case config.SecretNameDMS:
	case config.SecretNameAlertmanager:
	default:
		reqLogger.Info("Skip reconcile: No changes detected to alertmanager secrets.")
		return reconcile.Result{}, nil
	}
	log.Info("DEBUG: Started reconcile loop")

	// Get a list of all Secrets in the `openshift-monitoring` namespace.
	// This is used for determining which secrets are present so that the necessary
	// Alertmanager config changes can happen later.
	secretList := &corev1.SecretList{}
	opts := client.ListOptions{Namespace: request.Namespace}
	r.client.List(context.TODO(), &opts, secretList)

	// Check for the presence of specific secrets.
	pagerDutySecretExists := secretInList(config.SecretNamePD, secretList)
	snitchSecretExists := secretInList(config.SecretNameDMS, secretList)

	// Get the secret from the request.  If it's a secret we monitor, flag for reconcile.
	instance := &corev1.Secret{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)

	// if there was an error other than "not found" requeue
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("INFO: This secret has been deleted: %s", request.Name)
		} else {
			// Error and requeue in all other circumstances.
			// Don't requeue if a Secret is not found. It's valid to have an absent Pager Duty or DMS secret.
			log.Error(err, "Error reading object. Requeuing request")
			// NOTE originally updated metrics here, this has been removed
			return reconcile.Result{}, nil
		}
	}

	// do the work! collect secret info for PD and DMS
	pagerdutyRoutingKey := ""
	watchdogURL := ""
	// If a secret exists, add the necessary configs to Alertmanager.
	if pagerDutySecretExists {
		log.Info("INFO: Pager Duty secret exists")
		pagerdutyRoutingKey = readSecretKey(r, &request, config.SecretNamePD, secretKeyPD)
	}
	if snitchSecretExists {
		log.Info("INFO: Dead Man's Snitch secret exists")
		watchdogURL = readSecretKey(r, &request, config.SecretNameDMS, secretKeyDMS)
	}

	// create the desired alertmanager Config
	alertmanagerconfig := utility.CreateAlertManagerConfig(pagerdutyRoutingKey, watchdogURL)

	// write the alertmanager Config
	writeAlertManagerConfig(r, alertmanagerconfig)
	// Update metrics after all reconcile operations are complete.
	metrics.UpdateSecretsMetrics(secretList, alertmanagerconfig)
	reqLogger.Info("Finished reconcile for secret.")
	return reconcile.Result{}, nil
}

// secretInList takes the name of Secret, and a list of Secrets, and returns a Bool
// indicating if the name is present in the list
func secretInList(name string, list *corev1.SecretList) bool {
	for _, secret := range list.Items {
		if name == secret.Name {
			log.Info("DEBUG: Secret named", secret.Name, "found")
			return true
		}
	}
	log.Info("DEBUG: Secret", name, "not found")
	return false
}

// readSecretKey fetches the data from a Secret, such as a PagerDuty API key.
func readSecretKey(r *ReconcileSecret, request *reconcile.Request, secretname string, fieldname string) string {

	secret := &corev1.Secret{}

	// Define a new objectKey for fetching the secret key.
	objectKey := client.ObjectKey{
		Namespace: request.Namespace,
		Name:      secretname,
	}

	// Fetch the key from the secret object.
	r.client.Get(context.TODO(), objectKey, secret)
	secretkey := secret.Data[fieldname]

	return string(secretkey)
}

// writeAlertManagerConfig writes the updated alertmanager config to the `alertmanager-main` secret in namespace `openshift-monitoring`.
func writeAlertManagerConfig(r *ReconcileSecret, amconfig *alertmanager.Config) {
	amconfigbyte, marshalerr := yaml.Marshal(amconfig)
	if marshalerr != nil {
		log.Error(marshalerr, "ERROR: failed to marshal Alertmanager config")
	}
	// This is commented out because it prints secrets, but it might be useful for debugging when running locally.
	//log.Info("DEBUG: Marshalled Alertmanager config:", string(amconfigbyte))

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.SecretNameAlertmanager,
			Namespace: "openshift-monitoring",
		},
		Data: map[string][]byte{
			"alertmanager.yaml": amconfigbyte,
		},
	}

	// Write the alertmanager config into the alertmanager secret.
	err := r.client.Update(context.TODO(), secret)
	if err != nil {
		if errors.IsNotFound(err) {
			// couldn't update because it didn't exist.
			// create it instead.
			err = r.client.Create(context.TODO(), secret)
		}
	}

	if err != nil {
		log.Error(err, "ERROR: Could not write secret alertmanger-main", "namespace", secret.Namespace)
		return
	}
	log.Info("INFO: Secret alertmanager-main successfully updated")
}
