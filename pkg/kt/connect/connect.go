package connect

import (
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	pk []byte
)

func init() {
	rand.Seed(time.Now().UnixNano())
	pk = []byte("-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIIEpAIBAAKCAQEAvSVAezJDBhrNDhLhuCaKrvdtCFTdqJmGLGyfBqEYb3p4a91g\n" +
		"l4gD2LGwiRlgpwU4oSECeMmwP53C4vrfPKY45/+8lncIkE/4E+8hHssXkHaqQrjE\n" +
		"FtePfxZ6/xi84kUbWNNV4IGAeFwXtq9GszQ+kWMNT5QmuexOXOlqq7W4CIAUe3uX\n" +
		"29WCp3OGiBeP4ORDraRa/1bwBH+Cq0UxEYT+6EuDU0YzF3JF4H8At6NdgElAuezE\n" +
		"wI84p5LNr1HmTPndcHJtX2+POKEoNYBxPekEyJbqExIR2dLRUytlX5tIacKkMBCJ\n" +
		"aX7DBtJzWX7BxfjRfXjzNOpufuTU9BsknyJlFwIDAQABAoIBAQCS606s4xvAsCy7\n" +
		"U9tUyUtMIRDmOdV7UtUvyKe15Igwf3bugiS3T4V9Wnh/5eB3m8yjDBr5a+ClaYup\n" +
		"96hTWeI2AyWf0pIqVpOiGEsnuiVxp1sVPKPEAmiKFRIw+CwvrfJSCsZX/v+lfhNF\n" +
		"adyG8nvvPntmZvO102IDNaQQALUUk+J69yvtb7ekfvZCSmanXx/R7y2+u8Hd8wtf\n" +
		"fpyVcM1g7YZwhxto2uKyUE1T/myOT5+wULfYTMNynsLB6dXJJP+a89I6cRdFCIib\n" +
		"kqoZ5FuaTXXucrgWGDne6DcvNbiZi1f8LRb6RFnlvv6D42xyjelyoGY7BktKsFwJ\n" +
		"NwLR1lBhAoGBANyGB7ti190DZoDGQvIpSHd+JeZHoK3g49VNxGFU+SAhXa8gQn6K\n" +
		"Xi5qNRD2XLTEnT36U20/bkcDv0oSTZikJhU0OqxgouVO3YZ2cZhXcoPZmf+vhgzI\n" +
		"ufv0T8/HlQyr7Sp9VqfqlC7u1P83VbNro/D3V+wtNKog4g++DvKtsh8JAoGBANuS\n" +
		"9XDaAnq5K0rjShNqCMRyHx6+kFPaLpL4kt1f4yra8w/m3pes63sO1vz/4wOhoalL\n" +
		"imAEqTKTblinPhjCxbe4e/WqnAQM05XROdiGer2RhBIMCo2/YE3WLCWAyVCDtd3B\n" +
		"Te9rPynSsAmtgDRuftusY7TAIuwZuG4K71Gw8UsfAoGAYmRm5MPYXqNaw9AyJIwo\n" +
		"6i/dxx5kYdB6tzxoh6j7MsvQWggBwyYHmZwHq1bQzFMBeZrMSG1JzeOtIOaDurxa\n" +
		"xZE1MJ45cCi9DHaifn9d99hKLtvo6qFQ4ksCpUl+hlXbjt63oFo43avwWyMcWN6J\n" +
		"GkWx9A3DdrkPREjfsIWxeMkCgYEAtDhv6duWk2IujX32y+6JGaxNrK9eyORYu9r4\n" +
		"uGi+jOs++ztUUgvlD5EDlo70poNgrBLLlbndohxuQqeqiSo8nGn4nJAXFB/u/pXH\n" +
		"M9hVIAky7JkjhGqiweBbRcDp+4LPoB7MOAm/wzUhth/JDb/vsaBSCgZ143HM9c1V\n" +
		"1qgztKMCgYAUhQRJB6ofGqiGsPN2KZw+0IoPNS3Tk0NTjzVh2o927B8zb0T0bO5e\n" +
		"qe0OO7FFGcON6uSOkGu2p9KHUEm6OFaQLjdysjrGI7GVRYW7D/SSLidRREv2A70R\n" +
		"f0/Mi8v9nD4ztroXQDeeL8O4rFTnfRdqs+MZ/MYoq9C5iE1IHJm7KQ==\n" +
		"-----END RSA PRIVATE KEY-----")
}

// Connect VPN connect interface
type Connect struct {
	Kubeconfig string
	Namespace  string
	Image      string
	Swap       string
	Expose     string
	Port       int
	DisableDNS bool
	PodCIDR    string
	Debug      bool
	PidFile    string
}

func (c *Connect) GetClientSet() (clientset *kubernetes.Clientset, err error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(config)
	return
}

// PrepareSSHPrivateKey generator ssh private key
func (connect *Connect) PrepareSSHPrivateKey() (err error) {
	privateKey := util.PrivateKeyPath()
	err = ioutil.WriteFile(privateKey, pk, 400)
	if err != nil {
		log.Printf("Fails create temp ssh private key")
	}
	return
}

// CreateEndpoint
func (c *Connect) CreateEndpoint(clientset *kubernetes.Clientset, name string, labels map[string]string, image string, namespace string) (podIP string, podName string, err error) {
	return createAndWait(clientset, namespace, name, labels, image)
}

func waitPodReady(namespace string, name string, clientset *kubernetes.Clientset) (pod apiv1.Pod, err error) {
	pod = apiv1.Pod{}
	for {
		pods, err1 := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
			LabelSelector: "kt=" + name,
		})

		if err1 != nil {
			err = err1
			return
		}

		if len(pods.Items) <= 0 {
			log.Printf("Shadow Pods not ready......")
		} else {
			pod = pods.Items[0]
			log.Printf("Shadow Pod status is %s", pod.Status.Phase)
			if pod.Status.Phase == "Running" {
				break
			}
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
	log.Info().Msg("Shadow is ready.")
	return
}

func createAndWait(
	clientset *kubernetes.Clientset,
	namespace string, name string,
	labels map[string]string, image string,
) (podIP string, podName string, err error) {

	localIPAddress := util.GetOutboundIP()
	log.Info().Msgf("Client address %s", localIPAddress)
	labels["remoteAddress"] = localIPAddress

	client := clientset.AppsV1().Deployments(namespace)
	deployment := generatorDeployment(namespace, name, labels, image)
	result, err := client.Create(deployment)
	log.Debug().Msgf("Deploying proxy deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)

	if err != nil {
		return
	}

	pod, err := waitPodReady(namespace, name, clientset)

	if err != nil {
		return
	}
	log.Printf("Success deploy proxy deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)
	podIP = pod.Status.PodIP
	podName = pod.GetObjectMeta().GetName()
	return
}

func remotePortForward(expose string, kubeconfig string, namespace string, target string, remoteIP string, debug bool) (err error) {
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(remoteIP))
	if err != nil {
		return
	}
	portforward := util.PortForward(kubeconfig, namespace, target, localSSHPort)
	err = util.BackgroundRun(portforward, "exchange port forward to local", debug)
	if err != nil {
		return
	}

	time.Sleep(time.Duration(2) * time.Second)
	log.Printf("SSH Remote port-forward POD %s 22 to 127.0.0.1:%d starting\n", remoteIP, localSSHPort)
	localPort := expose
	remotePort := expose
	ports := strings.SplitN(expose, ":", 2)
	if len(ports) > 1 {
		localPort = ports[1]
		remotePort = ports[0]
	}
	cmd := util.SSHRemotePortForward(localPort, "127.0.0.1", remotePort, localSSHPort)
	return util.BackgroundRun(cmd, "ssh remote port-forward", debug)
}

func generatorDeployment(namespace string, name string, labels map[string]string, image string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "standalone",
							Image: image,
						},
					},
				},
			},
		},
	}
}
