package connect

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Connect VPN connect interface
type Connect struct {
	Kubeconfig   string
	Namespace    string
	Name         string
	Image        string
	Swap         string
	Expose       string
	Port         int
	DisableDNS   bool
	podIP        string
	swapReplicas *int32
	labels       map[string]string
	cidrs        []string
	PodCIDR      string
	Debug        bool
	PidFile      string
}

// InitDaemon
func (c *Connect) InitDaemon() (err error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	cidrs, err := getCirds(clientset, c.PodCIDR)
	if err != nil {
		return
	}

	c.cidrs = cidrs
	c.createDaemon(clientset)
	return
}

// InitSwap prepare swap deployment
func (c *Connect) InitSwap() (err error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.createSwap(clientset)
	return
}

// InitMesh prepare swap deployment
func (c *Connect) InitMesh() (err error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.createMesh(clientset)
	return
}

func (c *Connect) createMesh(clientset *kubernetes.Clientset) (err error) {
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	origin, err := deploymentsClient.Get(c.Swap, metav1.GetOptions{})
	if err != nil {
		return err
	}

	meshVersion := strings.ToLower(randStringRunes(5))
	c.Name = origin.GetObjectMeta().GetName() + "-kt-" + meshVersion
	c.labels = origin.Spec.Selector.MatchLabels
	c.labels["kt"] = c.Name
	c.labels["version"] = meshVersion
	c.labels["kt-component"] = "mesh"
	c.labels["control-by"] = "kt"

	podIP, err := createAndWait(clientset, c.Namespace, c.Name, c.labels, c.Image)
	if err != nil {
		return err
	}

	log.Printf("-----------------------------------------------------------\n")
	log.Printf("|    Mesh Version '%s' You can update Istio rule       |\n", meshVersion)
	log.Printf("-----------------------------------------------------------\n")

	c.podIP = podIP
	return
}

func (c *Connect) createDaemon(clientset *kubernetes.Clientset) (err error) {
	c.Name = c.Name + "-" + strings.ToLower(randStringRunes(5))
	c.labels = map[string]string{
		"kt":           c.Name,
		"kt-component": "connect",
		"control-by":   "kt",
	}
	podIP, err := createAndWait(clientset, c.Namespace, c.Name, c.labels, c.Image)
	if err != nil {
		return err
	}
	c.podIP = podIP
	return
}

func (c *Connect) createSwap(clientset *kubernetes.Clientset) (err error) {
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	origin, err := deploymentsClient.Get(c.Swap, metav1.GetOptions{})
	if err != nil {
		return err
	}

	c.swapReplicas = origin.Spec.Replicas

	replicas := int32(0)
	origin.Spec.Replicas = &replicas
	d, err := deploymentsClient.Update(origin)
	if err != nil {
		log.Printf("Fails scale deployment %s to zero\n", origin.GetObjectMeta().GetName())
		return err
	}
	log.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	log.Printf("Scale deployment %s to zero\n", origin.GetObjectMeta().GetName())

	c.Name = origin.GetObjectMeta().GetName() + "-kt-" + strings.ToLower(randStringRunes(5))

	c.labels = origin.Spec.Selector.MatchLabels
	c.labels["kt"] = c.Name
	c.labels["kt-component"] = "exchange"
	c.labels["control-by"] = "kt"

	podIP, err := createAndWait(clientset, c.Namespace, c.Name, c.labels, c.Image)
	if err != nil {
		return err
	}
	c.podIP = podIP
	return
}

func createAndWait(
	clientset *kubernetes.Clientset,
	namespace string, name string,
	labels map[string]string, image string,
) (podIP string, err error) {

	localIPAddress := util.GetOutboundIP()
	log.Printf("Client address %s", localIPAddress)
	labels["remoteAddress"] = localIPAddress

	client := clientset.AppsV1().Deployments(namespace)
	deployment := buildDeployment(namespace, name, labels, image)
	result, err := client.Create(deployment)
	log.Printf("Deploying proxy deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)

	if err != nil {
		return
	}

	pod, err := waitPodReady(namespace, name, clientset)

	if err != nil {
		return
	}
	log.Printf("Success deploy proxy deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)
	podIP = pod.Status.PodIP
	return
}

func buildDeployment(namespace string, name string, labels map[string]string, image string) *appsv1.Deployment {
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

func getCirds(clientset *kubernetes.Clientset, podCIDR string) (cidrs []string, err error) {
	cidrs, err = getPodCirds(clientset, podCIDR)
	if err != nil {
		return
	}
	serviceCird, err := getServiceCird(clientset)
	if err != nil {
		return
	}
	cidrs = append(cidrs, serviceCird)
	return
}

func getPodCirds(clientset *kubernetes.Clientset, podCIDR string) (cidrs []string, err error) {
	cidrs = []string{}

	if len(podCIDR) != 0 {
		cidrs = append(cidrs, podCIDR)
		return
	}

	nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		log.Printf("Fails to get node info of cluster")
		return nil, err
	}

	for _, node := range nodeList.Items {
		if node.Spec.PodCIDR != "" && len(node.Spec.PodCIDR) != 0 {
			cidrs = append(cidrs, node.Spec.PodCIDR)
		}
	}
	return
}

func getServiceCird(clientset *kubernetes.Clientset) (cidr string, err error) {
	serviceList, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
	if err != nil {
		log.Printf("Fails to get service info of cluster")
		return "", err
	}

	cluserIps := []string{}
	for _, service := range serviceList.Items {
		if service.Spec.ClusterIP != "" && service.Spec.ClusterIP != "None" {
			cluserIps = append(cluserIps, service.Spec.ClusterIP)
		}
	}

	sample := cluserIps[0]
	cidr = strings.Join(append(strings.Split(sample, ".")[:2], []string{"0", "0"}...), ".") + "/16"
	return
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
			log.Printf("Pods not ready......\n")
		} else {
			pod = pods.Items[0]
			log.Printf("Pod status is %s\n", pod.Status.Phase)
			if pod.Status.Phase == "Running" {
				break
			}
		}

		time.Sleep(time.Duration(2) * time.Second)
	}
	return
}

// Start start vpn connection
func (c *Connect) Start() (err error) {
	err = prepareSSHPrivateKey()
	if err != nil {
		return
	}
	err = c.forwardPorxy()
	if err != nil {
		return
	}
	time.Sleep(time.Duration(5) * time.Second)
	err = c.startVPN()
	if err != nil {
		return
	}
	log.Printf("KT proxy start successful\n")
	return
}

// RemotePortForwardToPod
func (c *Connect) RemotePortForwardToPod() (err error) {
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(c.podIP))
	if err != nil {
		return
	}
	portforward := util.PortForward(c.Kubeconfig, c.Namespace, c.Name, localSSHPort)
	err = BackgroundRun(portforward, "exchange port forward to local", c.Debug)
	if err != nil {
		return
	}

	time.Sleep(time.Duration(2) * time.Second)

	fmt.Printf("SSH Remote port-forward POD %s 22 to 127.0.0.1:%d starting\n", c.podIP, localSSHPort)
	cmd := util.SSHRemotePortForward(c.Expose, "127.0.0.1", c.Expose, localSSHPort)
	return BackgroundRun(cmd, "ssh remote port-forward", c.Debug)
}

// ExposeToLocal remote port forward to local
func (c *Connect) ExposeToLocal() (err error) {
	if c.Expose == "" {
		return
	}
	fmt.Printf("SSH Remote port-forward starting\n")
	cmd := util.SSHRemotePortForward(c.Expose, "127.0.0.1", c.Expose, c.Port)
	return BackgroundRun(cmd, "ssh remote port-forward", c.Debug)
}

// Exit cleanup proxy deployment in proxy
func (c *Connect) Exit() {
	os.Remove(c.PidFile)
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	log.Println("Cleanup proxy deplyment " + c.Name)
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	deletePolicy := metav1.DeletePropagationForeground
	deploymentsClient.Delete(c.Name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if c.Swap != "" {
		origin, err := deploymentsClient.Get(c.Swap, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Fail to get swap deployment %s in cluster", c.Swap)
		}
		log.Println("Recover origin deplyment " + c.Swap)
		origin.Spec.Replicas = c.swapReplicas
		var one = int32(1)
		origin.Spec.Replicas = &one

		d, err := deploymentsClient.Update(origin)
		if err != nil {
			log.Printf("Fail to revert deployment %s in cluster, please mannual scale it.", c.Swap)
		}
		log.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}

}

func prepareSSHPrivateKey() (err error) {
	err = ioutil.WriteFile("/tmp/kt_id_rsa", pk, 400)
	if err != nil {
		log.Printf("Fails create temp ssh private key")
	}
	return
}

func (c *Connect) command(cmdLine string) *exec.Cmd {
	if c.Debug {
		log.Printf("[DEBUG] %s \n", cmdLine)
	}
	cmdArgs := strings.Fields(cmdLine)
	return exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)]...)
}

func (c *Connect) forwardPorxy() (err error) {
	return BackgroundRun(util.PortForward(c.Kubeconfig, c.Namespace, c.Name, c.Port), "port-forward", c.Debug)
}

func (c *Connect) startVPN() (err error) {
	return BackgroundRun(util.SSHUttle("127.0.0.1", c.Port, c.podIP, c.DisableDNS, c.cidrs), "vpn(sshuttle)", c.Debug)
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
