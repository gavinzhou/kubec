package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"

	core_util "github.com/appscode/kutil/core/v1"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// CreateSecret is CreateSecret
func CreateSecret(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*core.Secret) *core.Secret) (*core.Secret, string, error) {
	cur, err := c.CoreV1().Secrets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
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
	return PatchSecret(c, cur, transform)

}

func PatchSecret(c kubernetes.Interface, cur *core.Secret, transform func(*core.Secret) *core.Secret) (*core.Secret, string, error) {
	return PatchSecretObject(c, cur, transform(cur.DeepCopy()))
}

func PatchSecretObject(c kubernetes.Interface, cur, mod *core.Secret) (*core.Secret, string, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, "cur-unchanged", err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, "mod-unchanged", err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, core.Secret{})
	if err != nil {
		return nil, "patch-unchanged", err
	}

	if len(patch) == 0 || string(patch) == "{}" {
		return cur, "unchanged", nil
	}
	fmt.Printf("Patching Secret %s/%s", cur.Namespace, cur.Name)
	out, err := c.CoreV1().Secrets(cur.Namespace).Patch(cur.Name, types.StrategicMergePatchType, patch)
	return out, "Patched", err
}

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
	data := map[string][]byte{
		"admin":  []byte("admin"),
		"passwd": []byte("passwd"),
	}
	// _, status, err := CreateSecret(clientset,
	// 	metav1.ObjectMeta{Name: secretName, Namespace: core.NamespaceDefault},
	// 	func(in *core.Secret) *core.Secret {
	// 		if in.Data == nil {
	// 			in.Data = data // can not change data
	// 		}
	// 		// in.Data["admin"] = []byte("admin")
	// 		// in.Data["passwd"] = []byte("passwd")
	// 		// this OK
	// 		return in
	// 	})
	// fmt.Println(status)
	// fmt.Println(err)
	_, s, err := core_util.CreateOrPatchSecret(clientset,
		metav1.ObjectMeta{Namespace: core.NamespaceDefault, Name: secretName},
		func(in *core.Secret) *core.Secret {
			if in.Data == nil {
				// in.Data = make(map[string][]byte)
				in.Data = data
			}
			// in.Data["admin"] = []byte("admin")
			// in.Data["password"] = []byte("password")
			return in
		})
	fmt.Println(s)
	fmt.Println(err)
}
