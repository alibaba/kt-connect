package cluster

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"strconv"
	"strings"
	"time"
)

// PodMetaAndSpec ...
type PodMetaAndSpec struct {
	Meta  *ResourceMeta
	Image string
	Envs  map[string]string
}

// GetPod ...
func (k *Kubernetes) GetPod(name string, namespace string) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetPodsByLabel get pods by label
func (k *Kubernetes) GetPodsByLabel(labels map[string]string, namespace string) (*coreV1.PodList, error) {
	return k.Clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// CreatePod create pod
func (k *Kubernetes) CreatePod(metaAndSpec *PodMetaAndSpec) error {
	if _, err := k.Clientset.CoreV1().Pods(metaAndSpec.Meta.Namespace).
		Create(context.TODO(), createPod(metaAndSpec), metav1.CreateOptions{}); err != nil {
		return err
	}
	SetupHeartBeat(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, k.UpdatePodHeartBeat)
	return nil
}

// UpdatePod ...
func (k *Kubernetes) UpdatePod(pod *coreV1.Pod) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
}

// RemovePod remove pod instances
func (k *Kubernetes) RemovePod(name, namespace string) (err error) {
	podsClient := k.Clientset.CoreV1().Pods(namespace)
	deletePolicy := metav1.DeletePropagationBackground
	return podsClient.Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// CreateShadowPod create shadow pod
func (k *Kubernetes) CreateShadowPod(metaAndSpec *PodMetaAndSpec, sshcm string) error {
	pod := createPod(metaAndSpec)
	pod.Spec.Containers[0].VolumeMounts = []coreV1.VolumeMount{
		{
			Name:      "ssh-public-key",
			MountPath: fmt.Sprintf("/root/%s", common.SshAuthKey),
		},
	}
	pod.Spec.Volumes = []coreV1.Volume{
		getSSHVolume(sshcm),
	}
	if _, err := k.Clientset.CoreV1().Pods(metaAndSpec.Meta.Namespace).
		Create(context.TODO(), pod, metav1.CreateOptions{}); err != nil {
		return err
	}
	SetupHeartBeat(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, k.UpdatePodHeartBeat)
	return nil
}

// WaitPodReady ...
func (k *Kubernetes) WaitPodReady(name, namespace string, timeoutSec int) (*coreV1.Pod, error) {
	return k.waitPodReady(name, namespace, timeoutSec, 0)
}

// WaitPodTerminate ...
func (k *Kubernetes) WaitPodTerminate(name, namespace string) (*coreV1.Pod, error) {
	return k.waitPodTerminate(name, namespace, 0)
}

func (k *Kubernetes) UpdatePodHeartBeat(name, namespace string) {
	log.Debug().Msgf("Heartbeat pod %s ticked at %s", name, formattedTime())
	if _, err := k.Clientset.CoreV1().Pods(namespace).
		Patch(context.TODO(), name, types.JSONPatchType, []byte(resourceHeartbeatPatch()), metav1.PatchOptions{}); err != nil {
		log.Warn().Err(err).Msgf("Failed to update pod heart beat")
	}
}

func (k *Kubernetes) ExecInPod(containerName, podName, namespace string, cmd ...string) (string, string, error) {
	req := k.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)
	req.VersionedParams(&coreV1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	log.Debug().Msgf("Execute command %v in %s:%s", cmd, podName, containerName)
	err := execute("POST", req.URL(), opt.Get().RuntimeStore.RestConfig, nil, &stdout, &stderr, false)
	stdoutMsg := util.RemoveColor(strings.TrimSpace(stdout.String()))
	stderrMsg := util.RemoveColor(strings.TrimSpace(stderr.String()))
	rawErrMsg := util.ExtractErrorMessage(stderrMsg)
	if err == nil && rawErrMsg != "" {
		err = fmt.Errorf(rawErrMsg)
	}
	return stdoutMsg, stderrMsg, err
}

// AddEphemeralContainer add ephemeral container to specified pod
func (k *Kubernetes) AddEphemeralContainer(containerName string, name string,
	envs map[string]string) (string, error) {
	pod, err := k.GetPod(name, opt.Get().Namespace)
	if err != nil {
		return "", err
	}

	privateKeyPath := util.PrivateKeyPath(name)
	generator, err := util.Generate(privateKeyPath)
	if err != nil {
		return "", err
	}
	configMap, err2 := k.CreateConfigMapWithSshKey(map[string]string{}, name, opt.Get().Namespace, generator)

	if err2 != nil {
		return "", errors.New("Found shadow pod but no configMap. Please delete the pod " + pod.Name)
	}

	err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[common.SshAuthPrivateKey]))

	privateKey := base64.StdEncoding.EncodeToString([]byte(configMap.Data[common.SshAuthPrivateKey]))

	ec := coreV1.EphemeralContainer{
		EphemeralContainerCommon: coreV1.EphemeralContainerCommon{
			Name:  containerName,
			Image: opt.Get().Image,
			Env: []coreV1.EnvVar{
				{Name: common.SshAuthPrivateKey, Value: privateKey},
			},
			SecurityContext: &coreV1.SecurityContext{
				Capabilities: &coreV1.Capabilities{Add: []coreV1.Capability{"NET_ADMIN"}},
			},
		},
	}

	for k, v := range envs {
		ec.Env = append(ec.Env, coreV1.EnvVar{Name: k, Value: v})
	}

	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, ec)

	pod, err = k.Clientset.CoreV1().Pods(pod.Namespace).UpdateEphemeralContainers(context.TODO(), pod.Name, pod, metav1.UpdateOptions{})
	return privateKeyPath, err
}

// RemoveEphemeralContainer remove ephemeral container from specified pod
func (k *Kubernetes) RemoveEphemeralContainer(_, podName string, namespace string) (err error) {
	// TODO: implement container removal
	return k.RemovePod(podName, namespace)
}

// IncreaseRef increase pod ref count by 1
func (k *Kubernetes) IncreaseRef(name string, namespace string) error {
	pod, err := k.Clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	annotations := pod.ObjectMeta.Annotations
	count, err := strconv.Atoi(annotations[common.KtRefCount])
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse annotations[%s] of pod %s with value %s",
			common.KtRefCount, name, annotations[common.KtRefCount])
		return err
	}

	pod.Annotations[common.KtRefCount] = strconv.Itoa(count + 1)

	_, err = k.Clientset.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}

// DecreaseRef decrease pod ref count by 1
func (k *Kubernetes) DecreaseRef(name string, namespace string) (cleanup bool, err error) {
	pod, err := k.GetPod(name, namespace)
	if err != nil {
		return
	}
	refCount := pod.Annotations[common.KtRefCount]
	if refCount == "1" {
		cleanup = true
		log.Info().Msgf("Pod %s has only one ref, gonna remove", name)
		err = k.RemovePod(pod.Name, pod.Namespace)
	} else {
		err = k.decreasePodRefByOne(refCount, pod)
	}
	return
}

func (k *Kubernetes) waitPodReady(name, namespace string, timeoutSec int, times int) (*coreV1.Pod, error) {
	const interval = 6
	if times > timeoutSec / interval {
		return nil, fmt.Errorf("pod %s failed to start", name)
	}
	pod, err := k.GetPod(name, namespace)
	if err != nil {
		return nil, err
	}
	if pod.Status.Phase != coreV1.PodRunning {
		log.Info().Msgf("Waiting for pod %s ...", name)
		time.Sleep(interval * time.Second)
		return k.waitPodReady(name, namespace, timeoutSec, times + 1)
	}
	log.Info().Msgf("Pod %s is ready", pod.Name)
	return pod, err
}

func (k *Kubernetes) waitPodTerminate(name, namespace string, times int) (*coreV1.Pod, error) {
	const interval = 6
	if times > 10 {
		return nil, fmt.Errorf("pod '%s' still terminating, please try again later", name)
	}
	log.Info().Msgf("Pod '%s' not finished yet, waiting ...", name)
	time.Sleep(interval * time.Second)
	routerPod, err := k.GetPod(name, namespace)
	if err != nil {
		// Note: will return a Not Found error when pod finally terminated
		return nil, err
	} else if routerPod.DeletionTimestamp != nil {
		return k.waitPodTerminate(name, namespace, times+1)
	} else {
		return routerPod, nil
	}
}

func (k *Kubernetes) decreasePodRefByOne(refCount string, pod *coreV1.Pod) (err error) {
	count, err := decreaseRef(refCount)
	if err != nil {
		return
	}
	log.Info().Msgf("Pod %s has %s refs, decrease to %s", pod.Name, refCount, count)
	util.MapPut(pod.Annotations, common.KtRefCount, count)
	_, err = k.UpdatePod(pod)
	return
}