package main

import (
	"flag"
	"fmt"
	"path/filepath"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// CreateSecret is CreateSecret
func CreateSecret(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*core.Secret) *core.Secret) (*core.Secret, string, error) {
	cur, err := c.CoreV1().Secrets(core.NamespaceDefault).Get(meta.Name, metav1.GetOptions{})
	fmt.Println(cur.Name)
	if kerr.IsNotFound(err) {
		fmt.Printf("create secret %v", meta.Name)
		out, err := c.CoreV1().Secrets(meta.Namespace).Create(transform(&core.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: core.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, "create", err
	} else if err != nil {
		return nil, "unchange", err
		// fmt.Printf("update secret %v", meta.Name)
		// c.CoreV1().Secrets(core.NamespaceDefault).Update(secret)
	}
	return nil, "exist", nil
}

// func PatchSecret(c kubernetes.Interface, cur *core.Secret, transform func(*core.Secret) *core.Secret) (*core.Secret, string, error) {
// 	return PatchSecretObject(c, cur, transform(cur.DeepCopy()))
// }

// func PatchSecretObject(c kubernetes.Interface, cur, mod *core.Secret) {
// 	curJson, err := json.Marshal(cur)
// 	if err != nil {
// 		return
// 	}

// 	modJson, err := json.Marshal(mod)
// 	if err != nil {
// 		return
// 	}
// }

func main() {

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
	clientset := kubernetes.NewForConfigOrDie(config)
	if err != nil {
		panic(err)
	}
	secretName := "user-passwd"
	s := map[string][]byte{
		"admin":    []byte("admin"),
		"password": []byte("passwd"),
	}
	_, status, _ := CreateSecret(clientset,
		metav1.ObjectMeta{Name: secretName, Namespace: core.NamespaceDefault},
		func(in *core.Secret) *core.Secret {
			if in.Data == nil {
				in.Data = s
			}
			return in
		})
	fmt.Println(status)
}
