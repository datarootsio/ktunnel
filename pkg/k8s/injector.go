package k8s

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"net/url"
	"strings"
	"time"
)

func injectToDeployment(o *appsv1.Deployment, c *apiv1.Container, image string, readyChan chan<- bool) (bool, error) {
	if hasSidecar(o.Spec.Template.Spec, image) {
		log.Warn(fmt.Sprintf("%s already injected to the deplyoment", image))
		readyChan <- true
		return true, nil
	}
	o.Spec.Template.Spec.Containers = append(o.Spec.Template.Spec.Containers, *c)
	_, updateErr := deploymentsClient.Update(o)
	if updateErr != nil {
		return false, updateErr
	}
	return true, nil
}

func InjectSidecar(namespace, objectName *string, port *int, image string, readyChan chan<- bool) (bool, error) {
	log.Infof("Injecting tunnel sidecar to %s/%s", *namespace, *objectName)
	getClients(namespace)
	co := newContainer(*port, image)
	creationTime := time.Now().Add(-1 * time.Second)
	obj, err := deploymentsClient.Get(*objectName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if *obj.Spec.Replicas > int32(1) {
		return false, errors.New("sidecar injection only support deployments with one replica")
	}
	_, err = injectToDeployment(obj, co, image, readyChan)
	if err != nil {
		return false, err
	}

	waitForReady(objectName, creationTime, *obj.Spec.Replicas, readyChan)
	return true, nil
}

func removeFromSpec(s *apiv1.PodSpec, image string) (bool, error) {
	if !hasSidecar(*s, image) {
		return true, errors.New(fmt.Sprintf("%s is not present on spec", image))
	}
	cIndex := -1
	for i, c := range s.Containers {
		if c.Image == image {
			cIndex = i
			break
		}
	}

	if cIndex != -1 {
		containers := s.Containers
		s.Containers = append(containers[:cIndex], containers[cIndex+1:]...)
		return true, nil
	} else {
		return false, errors.New("container not found on spec")
	}
}

func RemoveSidecar(namespace, objectName *string, image string, readyChan chan<- bool) (bool, error) {
	log.Infof("Removing tunnel sidecar from %s/%s", *namespace, *objectName)
	getClients(namespace)
	obj, err := deploymentsClient.Get(*objectName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	deletionTime := time.Now().Add(-1 * time.Second)
	_, err = removeFromSpec(&obj.Spec.Template.Spec, image)
	if err != nil {
		return false, err
	}
	_, updateErr := deploymentsClient.Update(obj)
	if updateErr != nil {
		return false, updateErr
	}
	waitForReady(objectName, deletionTime, *obj.Spec.Replicas, readyChan)
	return true, nil
}

func getPortForwardUrl(config *rest.Config, namespace string, podName string) *url.URL {
	host := strings.TrimPrefix(config.Host, "https://")
	trailingHostPath := strings.Split(host, "/")
	hostIp := trailingHostPath[0]
	trailingPath := ""
	if len(trailingHostPath) > 1 {
		trailingPath = fmt.Sprintf("/%s", strings.Join(trailingHostPath[1:], "/"))
	}
	path := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", trailingPath, namespace, podName)
	return &url.URL{
		Scheme: "https",
		Path:   path,
		Host:   hostIp,
	}
}
