package cps

import (
	galasav1alpha1 "github.com/galasa-dev/extensions/galasa-ecosystem-operator/pkg/apis/galasa/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var instanceWithProps = &galasav1alpha1.GalasaEcosystem{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test-ecosystem",
		Namespace: "test-namespace",
	},
	Spec: galasav1alpha1.GalasaEcosystemSpec{
		Propertystore: galasav1alpha1.PropertyStoreCluster{
			InitProps: map[string]string{
				"PropName":       "PropValue",
				"SecondPropName": "SecondPropValue",
			},
		},
	},
}

//TODO need to MOCK the executor
