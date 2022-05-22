package cluster

import (
	"bytes"
	"context"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PodMetaAndSpec ...
type PodMetaAndSpec struct {
	Meta  *ResourceMeta
	Image string
	Envs  map[string]string
	Ports map[string]int
	IsLeaf bool
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

// UpdatePod ...
func (k *Kubernetes) UpdatePod(pod *coreV1.Pod) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
}

// RemovePod remove pod instances
func (k *Kubernetes) RemovePod(name, namespace string) (err error) {
	deletePolicy := metav1.DeletePropagationBackground
	return k.Clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func (k *Kubernetes) WaitPodsReady(labels map[string]string, namespace string, timeoutSec int) ([]coreV1.Pod, error) {
	return k.waitPodsReady(labels, namespace, timeoutSec, 0)
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
	if _, err := k.Clientset.CoreV1().Pods(namespace).
		Patch(context.TODO(), name, types.JSONPatchType, []byte(resourceHeartbeatPatch()), metav1.PatchOptions{}); err != nil {
		if healthy, exists := LastHeartBeatStatus["pod_" + name]; healthy || !exists {
			log.Warn().Err(err).Msgf("Failed to update heart beat of pod %s", name)
		} else {
			log.Debug().Err(err).Msgf("Pod %s heart beat interrupted", name)
		}
		LastHeartBeatStatus["pod_" + name] = false
	} else {
		log.Debug().Msgf("Heartbeat pod %s ticked at %s", name, util.FormattedTime())
		LastHeartBeatStatus["pod_" + name] = true
	}
}

// WatchPod ...
func (k *Kubernetes) WatchPod(name, namespace string, fAdd, fDel, fMod func(*coreV1.Pod)) {
	k.watchResource(name, namespace, string(coreV1.ResourcePods), &coreV1.Pod{},
		func(obj any) {
			handlePodEvent(obj, "added", fAdd)
		},
		func(obj any) {
			handlePodEvent(obj, "deleted", fDel)
		},
		func(obj any) {
			handlePodEvent(obj, "modified", fMod)
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
	err := execute("POST", req.URL(), opt.Store.RestConfig, nil, &stdout, &stderr, false)
	stdoutMsg := util.RemoveColor(strings.TrimSpace(stdout.String()))
	stderrMsg := util.RemoveColor(strings.TrimSpace(stderr.String()))
	rawErrMsg := util.ExtractErrorMessage(stderrMsg)
	if err == nil && rawErrMsg != "" {
		err = fmt.Errorf(rawErrMsg)
	}
	return stdoutMsg, stderrMsg, err
}

// IncreasePodRef increase pod ref count by 1
func (k *Kubernetes) IncreasePodRef(name string, namespace string) error {
	pod, err := k.GetPod(name, namespace)
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
	_, err = k.UpdatePod(pod)
	return err
}

// DecreasePodRef decrease pod ref count by 1
func (k *Kubernetes) DecreasePodRef(name string, namespace string) (bool, error) {
	pod, err := k.GetPod(name, namespace)
	if err != nil {
		return false, err
	}
	refCount := pod.Annotations[util.KtRefCount]
	if refCount == "1" {
		log.Info().Msgf("Pod %s has only one ref, gonna remove", name)
		return true, nil
	} else {
		count, err2 := decreaseRef(refCount)
		if err2 != nil {
			return false, err2
		}
		log.Info().Msgf("Pod %s has %s refs, decrease to %s", pod.Name, refCount, count)
		pod.Annotations = util.MapPut(pod.Annotations, util.KtRefCount, count)
		_, err = k.UpdatePod(pod)
		return false, nil
	}
}

func handlePodEvent(obj any, status string, f func(*coreV1.Pod)) {
	switch obj.(type) {
	case *coreV1.Pod:
		if f != nil {
			log.Debug().Msgf("Pod %s %s", obj.(*coreV1.Pod).Name, status)
			f(obj.(*coreV1.Pod))
		}
	default:
		// ignore
	}
}

func (k *Kubernetes) waitPodsReady(labels map[string]string, namespace string, timeoutSec int, times int) ([]coreV1.Pod, error) {
	pods, err := k.GetPodsByLabel(labels, namespace)
	if err != nil {
		return nil, err
	}
	const interval = 6
	if times > timeoutSec / interval {
		if len(pods.Items) < 1 {
			return nil, fmt.Errorf("pod with label %v not found", labels)
		} else {
			return nil, fmt.Errorf("pod %s failed to start", pods.Items[0].Name)
		}
	}
	runningPods := filterRunningPods(pods.Items)
	if len(runningPods) > 0 {
		log.Info().Msgf("Pod %s is ready", runningPods[0].Name)
		return runningPods, nil
	}
	log.Info().Msgf("Waiting for shadow pod ...")
	time.Sleep(1 * time.Second)
	return k.waitPodsReady(labels, namespace, timeoutSec, times + 1)
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
		if strings.HasPrefix(name, util.RectifierPodPrefix) {
			log.Info().Msgf("Fetching cluster time ...")
		} else {
			log.Info().Msgf("Waiting for pod %s ...", name)
		}
		time.Sleep(interval * time.Second)
		return k.waitPodReady(name, namespace, timeoutSec, times + 1)
	}
	if !strings.HasPrefix(name, util.RectifierPodPrefix) {
		log.Info().Msgf("Pod %s is ready", pod.Name)
	}
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

func addImagePullSecret(pod *coreV1.Pod, imagePullSecret string) {
	pod.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{
		{
			Name: imagePullSecret,
		},
	}
}

func addResourceLimit(container *coreV1.Container, quotaText string) {
	for _, quota := range strings.Split(quotaText, ",") {
		if ok, err := regexp.MatchString("^[0-9.]+[Cc]$", quota); ok && err == nil {
			limit := quota[0: len(quota) - 1]
			container.Resources.Limits[coreV1.ResourceCPU] = resource.MustParse(limit)
			container.Resources.Requests[coreV1.ResourceCPU] = resource.MustParse(limit)
		} else if ok2, err2 := regexp.MatchString("^[0-9.]+[KkMmGg]$", quota); ok2 && err2 == nil {
			limit := strings.ToUpper(quota) + "i"
			container.Resources.Limits[coreV1.ResourceMemory] = resource.MustParse(limit)
			container.Resources.Requests[coreV1.ResourceMemory] = resource.MustParse(limit)
		} else {
			log.Warn().Msgf("Pod quote \"%s\" is invalid, ignoring", quota)
		}
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
