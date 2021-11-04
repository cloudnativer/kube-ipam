package kipam

import (
	"github.com/containernetworking/cni/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"context"
	"io/ioutil"
	"net"

	log "github.com/sirupsen/logrus"
)

type CommonArgs struct {
	IgnoreUnknown types.UnmarshallableBool `json:"ignoreunknown,omitempty"`
}

type K8sArgs struct {
	types.CommonArgs
	IP net.IP
	// 不能用直接用字符串
	K8S_POD_NAME               types.UnmarshallableString
	K8S_POD_NAMESPACE          types.UnmarshallableString
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString
}

func NewClient(kubeCfg string) (*kubernetes.Clientset, error) {
	kubeconfig, err := ioutil.ReadFile(kubeCfg)
	if err != nil {
		return nil, err
	}
	restConf, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConf)
}

func GetPodInfo(client *kubernetes.Clientset, podName, podNamespace string) (annotations map[string]string, err error) {
	pod, err := client.CoreV1().Pods(string(podNamespace)).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	log.Infof("pod info %+v", pod)
	return pod.Annotations, nil
}
