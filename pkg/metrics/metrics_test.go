package metrics

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/openshift/configure-alertmanager-operator/config"
	"github.com/openshift/configure-alertmanager-operator/pkg/utility"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func assertEquals(t *testing.T, want interface{}, got interface{}, message string) {
	if reflect.DeepEqual(got, want) {
		return
	}

	if len(message) == 0 {
		message = fmt.Sprintf("Expected '%v' but got '%v'", want, got)
	} else {
		message = fmt.Sprintf("%s: Expected '%v' but got '%v'", message, want, got)
	}
	t.Fatal(message)
}

func createSecret(secretname string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretname,
			Namespace: config.OperatorNamespace,
		},
	}

}

func Test_secretExists_false(t *testing.T) {
	secrets := &corev1.SecretList{}

	exists := secretExists(config.SecretNamePD, secrets)

	assertEquals(t, 0, exists, "")
}

func Test_secretExists_true(t *testing.T) {
	secrets := &corev1.SecretList{
		Items: []corev1.Secret{
			createSecret(config.SecretNamePD),
		},
	}

	exists := secretExists(config.SecretNamePD, secrets)

	assertEquals(t, 1, exists, "")
}

func Test_secretConfigured_PD_false(t *testing.T) {
	cfg := utility.CreateAlertManagerConfig("", "")

	exists := secretConfigured(config.SecretNamePD, cfg)

	assertEquals(t, 0, exists, "")
}

func Test_secretConfigured_PD_true(t *testing.T) {
	cfg := utility.CreateAlertManagerConfig("asdf", "")

	exists := secretConfigured(config.SecretNamePD, cfg)

	assertEquals(t, 1, exists, "")
}

func Test_secretConfigured_DMS_false(t *testing.T) {
	cfg := utility.CreateAlertManagerConfig("", "")

	exists := secretConfigured(config.SecretNamePD, cfg)

	assertEquals(t, 0, exists, "")
}

func Test_secretConfigured_DMS_true(t *testing.T) {
	cfg := utility.CreateAlertManagerConfig("", "asdf")

	exists := secretConfigured(config.SecretNameDMS, cfg)

	assertEquals(t, 1, exists, "")
}
