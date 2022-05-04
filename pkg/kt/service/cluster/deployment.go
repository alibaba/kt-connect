package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)

// GetDeployment ...
func (k *Kubernetes) GetDeployment(name string, namespace string) (*appV1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetDeploymentsByLabel get deployments by label
func (k *Kubernetes) GetDeploymentsByLabel(labels map[string]string, namespace string) (pods *appV1.DeploymentList, err error) {
	return k.Clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// GetAllDeploymentInNamespace get all deployment in specified namespace
func (k *Kubernetes) GetAllDeploymentInNamespace(namespace string) (*appV1.DeploymentList, error) {
	return k.Clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
}

// UpdateDeployment ...
func (k *Kubernetes) UpdateDeployment(deployment *appV1.Deployment) (*appV1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(deployment.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
}

// RemoveDeployment remove deployment instances
func (k *Kubernetes) RemoveDeployment(name, namespace string) (err error) {
	deletePolicy := metav1.DeletePropagationBackground
	return k.Clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// IncreaseDeploymentRef increase deployment ref count by 1
func (k *Kubernetes) IncreaseDeploymentRef(name string, namespace string) error {
	app, err := k.GetDeployment(name, namespace)
	if err != nil {
		return err
	}
	annotations := app.ObjectMeta.Annotations
	count, err := strconv.Atoi(annotations[util.KtRefCount])
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse annotations[%s] of deployment %s with value %s",
			util.KtRefCount, name, annotations[util.KtRefCount])
		return err
	}

	app.Annotations[util.KtRefCount] = strconv.Itoa(count + 1)
	_, err = k.UpdateDeployment(app)
	return err
}

// DecreaseDeploymentRef decrease deployment ref count by 1
func (k *Kubernetes) DecreaseDeploymentRef(name string, namespace string) (cleanup bool, err error) {
	app, err := k.GetDeployment(name, namespace)
	if err != nil {
		return
	}
	refCount := app.Annotations[util.KtRefCount]
	if refCount == "1" {
		log.Info().Msgf("Deployment %s has only one ref, gonna remove", name)
		return true, nil
	} else {
		count, err2 := decreaseRef(refCount)
		if err2 != nil {
			return
		}
		log.Info().Msgf("Deployment %s has %s refs, decrease to %s", app.Name, refCount, count)
		app.Annotations = util.MapPut(app.Annotations, util.KtRefCount, count)
		_, err = k.UpdateDeployment(app)
		return
	}
}

func (k *Kubernetes) UpdateDeploymentHeartBeat(name, namespace string) {
	if _, err := k.Clientset.AppsV1().Deployments(namespace).
		Patch(context.TODO(), name, types.JSONPatchType, []byte(resourceHeartbeatPatch()), metav1.PatchOptions{}); err != nil {
		if healthy, exists := LastHeartBeatStatus["deployment_" + name]; healthy || !exists {
			log.Warn().Err(err).Msgf("Failed to update heart beat of deployment %s", name)
		} else {
			log.Debug().Err(err).Msgf("Deployment %s heart beat interrupted", name)
		}
		LastHeartBeatStatus["deployment_" + name] = false
	} else {
		log.Debug().Msgf("Heartbeat deployment %s ticked at %s", name, util.FormattedTime())
		LastHeartBeatStatus["deployment_" + name] = true
	}
}

// ScaleTo scale deployment to
func (k *Kubernetes) ScaleTo(name, namespace string, replicas *int32) (err error) {
	deployment, err := k.GetDeployment(name, namespace)
	if err != nil {
		return
	}

	// replicas field is refer type, must compare with its real value
	if *deployment.Spec.Replicas == *replicas {
		log.Warn().Msgf("Deployment %s already having %d replicas, not need to scale", name, *replicas)
		return nil
	}

	log.Info().Msgf("Scaling deployment %s from %d to %d", deployment.Name, *deployment.Spec.Replicas, *replicas)
	deployment.Spec.Replicas = replicas

	if _, err = k.UpdateDeployment(deployment); err != nil {
		log.Error().Err(err).Msgf("Failed to scale deployment %s", deployment.Name)
		return
	}
	log.Info().Msgf("Deployment %s successfully scaled to %d replicas", name, *replicas)
	return
}
