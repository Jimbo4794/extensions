package cps

import (
	"testing"

	galasav1alpha1 "github.com/galasa-dev/extensions/galasa-ecosystem-operator/pkg/apis/galasa/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var instance = &galasav1alpha1.GalasaEcosystem{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test-ecosystem",
		Namespace: "test-namespace",
	},
	Spec: galasav1alpha1.GalasaEcosystemSpec{
		Propertystore: galasav1alpha1.PropertyStoreCluster{
			ClusterSize:  3,
			NodeSelector: map[string]string{"Node": "NodeName"},
			Storage:      "100Mi",
		},
	},
}

func TestServiceForm(t *testing.T) {
	service := generateExposedService(instance)
	if service.Name != "test-ecosystem-cps-external-service" {
		t.Error("Service name not generated correctly:" + service.Name)
	}
	if service.Namespace != "test-namespace" {
		t.Error("Service namespace incorrect:" + service.Namespace)
	}
}

func TestClusterForm(t *testing.T) {
	cluster := generateEtcdCluster(instance)
	if cluster.Name != "test-ecosystem-etcd-cluster" {
		t.Error("Cluster name not generated correctly:" + cluster.Name)
	}
	if cluster.Namespace != "test-namespace" {
		t.Error("Cluster namespace incorrect:" + cluster.Namespace)
	}
	if cluster.Labels["app"] != "test-ecosystem-cps" {
		t.Error("Labels for the cluster incorrect:" + cluster.Namespace)
	}
	if cluster.Spec.Size != 3 {
		t.Error("Cluster size incorrect:" + string(cluster.Spec.Size))
	}
}

func TestNewCluster(t *testing.T) {
	cps := New(instance)
	if cps.ExposedService == nil {
		t.Error("Exposed Service not created")
	}
	if cps.Cluster == nil {
		t.Error("Cluster not created")
	}
}
