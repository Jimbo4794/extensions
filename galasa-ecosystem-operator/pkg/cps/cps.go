package cps

import (
	etcd "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	galasav1alpha1 "github.com/galasa-dev/extensions/galasa-ecosystem-operator/pkg/apis/galasa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type CPS struct {
	ExposedService *corev1.Service
	Cluster        *etcd.EtcdCluster
}

func New(cr *galasav1alpha1.GalasaEcosystem) *CPS {
	return &CPS{
		ExposedService: generateExposedService(cr),
		Cluster:        generateEtcdCluster(cr),
	}
}

func generateExposedService(cr *galasav1alpha1.GalasaEcosystem) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-cps-external-service",
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceType(corev1.ServiceTypeNodePort),
			Ports: []corev1.ServicePort{
				{
					Name:       "etcd-client",
					TargetPort: intstr.FromInt(2379),
					Port:       2379,
				},
			},
			Selector: map[string]string{
				"etcd_cluster": cr.Name + "-etcd-cluster",
			},
		},
	}
}

func generateEtcdCluster(cr *galasav1alpha1.GalasaEcosystem) *etcd.EtcdCluster {
	return &etcd.EtcdCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-etcd-cluster",
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				"etcd.database.coreos.com/scope": "clusterwide",
			},
			Labels: map[string]string{
				"app": cr.Name + "-cps",
			},
		},
		Spec: etcd.ClusterSpec{
			Size:    cr.Spec.Propertystore.ClusterSize,
			Version: "3.2.13",
			Pod: &etcd.PodPolicy{
				EtcdEnv: []corev1.EnvVar{
					{
						Name:  "ETCDCTL_API",
						Value: "3",
					},
				},
				NodeSelector: cr.Spec.Propertystore.NodeSelector,
				PersistentVolumeClaimSpec: &corev1.PersistentVolumeClaimSpec{
					StorageClassName: cr.Spec.StorageClassName,
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(cr.Spec.Propertystore.Storage),
						},
					},
				},
			},
		},
	}
}
