package main

import (
	"flag"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	s := map[string][]byte{
		"admin":    []byte("admin"),
		"password": []byte("passwd"),
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	secretName := "user-passwd"
	pvcName := "pv-couchdb"
	// deploymentsClient := clientset.AppsV1beta1().Deployments(apiv1.NamespaceDefault)
	_, err = clientset.CoreV1().Secrets(apiv1.NamespaceDefault).Get(secretName, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		secret := &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: apiv1.NamespaceDefault,
			},
			Data: s,
		}
		clientset.CoreV1().Secrets(apiv1.NamespaceDefault).Create(secret)
		// clientset.CoreV1().Secrets(apiv1.NamespaceDefault).Update(secret)
	}
	// pvcSpec.StorageClassName = ""
	// StorageClass := "standard"

	// Dynamic Provisioning and Storage Classes in Kubernetes
	// http://blog.kubernetes.io/2017/03/dynamic-provisioning-and-storage-classes-kubernetes.html

	Size := "10Gi"

	pvc := &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvcName,
			// Annotations: map[string]string{
			// 	"volume.beta.kubernetes.io/storage-class": StorageClass,
			// },
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			// VolumeName: pvcName,
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			// StorageClassName: &StorageClass,
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceStorage: resource.MustParse(Size),
				},
			},
		},
	}
	clientset.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault).Create(pvc)
}
