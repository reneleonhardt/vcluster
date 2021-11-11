package kubeconfig

import (
	"context"
	"github.com/loft-sh/vcluster/pkg/util/clienthelper"
	"github.com/loft-sh/vcluster/pkg/util/loghelper"
	"github.com/loft-sh/vcluster/pkg/util/translate"
	"github.com/pkg/errors"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WriteKubeConfig(ctx context.Context, client client.Client, secretName, secretNamespace string, config *api.Config) error {
	out, err := clientcmd.Write(*config)
	if err != nil {
		return err
	}

	err = os.MkdirAll("/root/.kube", 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("/root/.kube/config", out, 0666)
	if err != nil {
		return err
	}

	if secretName != "" {
		kubeConfigSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: secretNamespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"config": out,
			},
		}

		// set owner reference
		if translate.Owner != nil && translate.Owner.GetNamespace() == kubeConfigSecret.Namespace {
			kubeConfigSecret.OwnerReferences = translate.GetOwnerReference()
		}

		err = clienthelper.Apply(ctx, client, kubeConfigSecret, loghelper.New("apply-secret"))
		if err != nil {
			return errors.Wrap(err, "apply generated kube config secret")
		}
	}

	return nil
}
