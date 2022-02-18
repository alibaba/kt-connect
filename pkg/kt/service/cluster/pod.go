package cluster

import (
	"bytes"
	"context"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PodMetaAndSpec ...
type PodMetaAndSpec struct {
	Meta  *ResourceMeta
	Image string
	Envs  map[string]string
	Ports []int
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

// WatchPod ...
func (k *Kubernetes) WatchPod(name, namespace string, fAdd, fDel, fMod func(*coreV1.Pod)) {
	k.watchResource(name, namespace, string(coreV1.ResourcePods), &coreV1.Pod{},
		func(obj interface{}) {
			if fAdd != nil {
				log.Debug().Msgf("Pod %s added", obj.(*coreV1.Pod).Name)
				fAdd(obj.(*coreV1.Pod))
			}
		},
		func(obj interface{}) {
			if fDel != nil {
				log.Debug().Msgf("Pod %s deleted", obj.(*coreV1.Pod).Name)
				fDel(obj.(*coreV1.Pod))
			}
		},
		func(obj interface{}) {
			if fMod != nil {
				log.Debug().Msgf("Pod %s modified", obj.(*coreV1.Pod).Name)
				fMod(obj.(*coreV1.Pod))
			}
		},
	)
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

// IncreaseRef increase pod ref count by 1
func (k *Kubernetes) IncreaseRef(name string, namespace string) error {
	pod, err := k.Clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	annotations := pod.ObjectMeta.Annotations
	count, err := strconv.Atoi(annotations[util.KtRefCount])
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse annotations[%s] of pod %s with value %s",
			util.KtRefCount, name, annotations[util.KtRefCount])
		return err
	}

	pod.Annotations[util.KtRefCount] = strconv.Itoa(count + 1)

	_, err = k.Clientset.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}

// DecreaseRef decrease pod ref count by 1
func (k *Kubernetes) DecreaseRef(name string, namespace string) (cleanup bool, err error) {
	pod, err := k.GetPod(name, namespace)
	if err != nil {
		return
	}
	refCount := pod.Annotations[util.KtRefCount]
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
	util.MapPut(pod.Annotations, util.KtRefCount, count)
	_, err = k.UpdatePod(pod)
	return
}

func addImagePullSecret(pod *coreV1.Pod, imagePullSecret string) {
	pod.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{
		{
			Name: imagePullSecret,
		},
	}
}

func execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}

func decreaseRef(refCount string) (count string, err error) {
	currentCount, err := strconv.Atoi(refCount)
	if err != nil {
		return
	}
	count = strconv.Itoa(currentCount - 1)
	return
}
