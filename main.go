package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/appscode/go/types"
	apps_util "github.com/appscode/kutil/apps/v1beta1"
	core_util "github.com/appscode/kutil/core/v1"

	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	ex_ct "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	// new config
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

	// init k8s client
	clientset := kubernetes.NewForConfigOrDie(config)
	if err != nil {
		panic(err)
	}

	// create secret
	secretName := "user-passwd"
	_, s, err := core_util.CreateOrPatchSecret(clientset,
		metav1.ObjectMeta{Namespace: core.NamespaceDefault, Name: secretName},
		func(in *core.Secret) *core.Secret {
			if in.Data == nil {
				in.Data = make(map[string][]byte)
			}
			in.Data["admin"] = []byte("admin")
			in.Data["password"] = []byte("password")
			return in
		})
	fmt.Println(s)
	fmt.Println(err)

	// create pvc
	pvcName := "testpvc"
	Size := "10Gi"
	_, pvcs, _ := core_util.CreateOrPatchPVC(clientset,
		metav1.ObjectMeta{Namespace: core.NamespaceDefault, Name: pvcName},
		func(p *core.PersistentVolumeClaim) *core.PersistentVolumeClaim {
			p.Spec = core.PersistentVolumeClaimSpec{
				AccessModes: []core.PersistentVolumeAccessMode{
					core.ReadWriteOnce,
				},
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse(Size),
					},
				},
			}
			return p
		})
	fmt.Println(pvcs)

	// create deployments
	deploymentName := "nginx"
	_, ds, err := apps_util.CreateOrPatchDeployment(clientset,
		metav1.ObjectMeta{Namespace: core.NamespaceDefault, Name: deploymentName},
		func(obj *apps.Deployment) *apps.Deployment {
			obj.Spec = apps.DeploymentSpec{
				Replicas: types.Int32P(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":         "deployment",
						"app-version": "v1",
					},
				},
				Template: core.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":         "deployment",
							"app-version": "v1",
						},
					},
					Spec: core.PodSpec{
						Volumes: []core.Volume{
							core.Volume{
								Name: "testpvc",
								VolumeSource: core.VolumeSource{
									PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
										ClaimName: "testpvc",
									},
								},
							},
						},
						Containers: []core.Container{
							core.Container{
								Name:  "nginx",
								Image: "nginx",
								Ports: []core.ContainerPort{
									{
										Name:          "http",
										ContainerPort: 80,
									},
								},
								VolumeMounts: []core.VolumeMount{
									{
										Name:      "testpvc",
										MountPath: "/mnt/testpvc",
									},
								},
							},
						},
					},
				},
			}
			return obj
		})
	fmt.Println(ds)
	if err != nil {
		fmt.Println(err)
	}

	// create svc
	svcName := "nginx"
	_, ss, err := core_util.CreateOrPatchService(clientset,
		metav1.ObjectMeta{Namespace: core.NamespaceDefault, Name: svcName},
		func(s *core.Service) *core.Service {
			s.Spec = core.ServiceSpec{
				ClusterIP: "None",
				Ports: []core.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   "TCP",
					},
				},
				Selector: map[string]string{
					"app":         "deployment",
					"app-version": "v1",
				},
			}
			return s
		})
	fmt.Println(ss)
	if err != nil {
		fmt.Println(err)
	}

	// create configmap
	configmapName := "nginx"
	_, cs, err := core_util.CreateOrPatchConfigMap(clientset,
		metav1.ObjectMeta{Namespace: core.NamespaceDefault, Name: configmapName},
		func(c *core.ConfigMap) *core.ConfigMap {
			c.Data = map[string]string{
				"nginx.conf": "nginx-cofnig",
			}
			return c
		})
	fmt.Println(cs)
	if err != nil {
		fmt.Println(err)
	}

	// create crd
	nginxCrd := extensionsobj.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "nginxs" + "." + "orangesys.io",
			Labels: map[string]string{},
		},
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group:   "orangesys.io",
			Version: "v1alpha1",
			Scope:   extensionsobj.NamespaceScoped,
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural: "nginxs",
				Kind:   "nginx",
			},
		},
	}
	crds := []*extensionsobj.CustomResourceDefinition{
		&nginxCrd,
	}

	// need "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset" client init
	clientSet := ex_ct.NewForConfigOrDie(config)
	crdClient := clientSet.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crd := range crds {
		if _, err := crdClient.Create(crd); err != nil && !apierrors.IsAlreadyExists(err) {
			fmt.Println(err)
		}
	}
}
