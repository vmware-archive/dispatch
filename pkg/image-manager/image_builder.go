///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package imagemanager

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watchv1 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
)

type BaseImageBuilder struct {
	baseImageChannel chan BaseImage
	done             chan bool
	es               entitystore.EntityStore
	clientset        *kubernetes.Clientset
	namespace        string
	orgID            string
}

type imageStatusResult struct {
	Result int `json:"result"`
}

func NewBaseImageBuilder(es entitystore.EntityStore) (*BaseImageBuilder, error) {
	var err error
	var config *rest.Config
	if ImageManagerFlags.K8sConfig == "" {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", ImageManagerFlags.K8sConfig)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubernetes config")
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating kubernetes client")
	}

	return &BaseImageBuilder{
		baseImageChannel: make(chan BaseImage),
		done:             make(chan bool),
		es:               es,
		clientset:        clientset,
		namespace:        ImageManagerFlags.K8sNamespace,
		orgID:            ImageManagerFlags.OrgID,
	}, nil
}

func (b *BaseImageBuilder) createPullJob(baseImage *BaseImage) (*batchv1.Job, error) {
	name := fmt.Sprintf("create-base-image-%v", baseImage.ID)
	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"role":        "create-base-image",
				"baseImageID": baseImage.ID,
			},
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds: swag.Int64(60),
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{apiv1.Container{
						Name:  name,
						Image: "berndtj/image-status:latest",
						Command: []string{
							"/image_status.sh",
							baseImage.DockerURL,
						},
						VolumeMounts: []apiv1.VolumeMount{apiv1.VolumeMount{
							Name:      "docker",
							MountPath: "/var/run/docker.sock",
						}},
					}},
					RestartPolicy: apiv1.RestartPolicyNever,
					Volumes: []apiv1.Volume{apiv1.Volume{
						Name: "docker",
						VolumeSource: apiv1.VolumeSource{
							HostPath: &apiv1.HostPathVolumeSource{
								Path: "/var/run/docker.sock",
							},
						},
					}},
				},
			},
		},
	}
	job, err := b.clientset.BatchV1().Jobs(b.namespace).Create(jobSpec)
	return job, err
}

func (b *BaseImageBuilder) poll() error {
	var baseImages []BaseImage
	err := b.es.List(b.orgID, nil, &baseImages)
	if err != nil {
		return errors.Wrap(err, "Failed to get base images from entity store")
	}
	for _, bi := range baseImages {
		_, err := b.createPullJob(&bi)
		if err != nil {
			log.Printf("Error creating job for base image: %s [%v]: %v", bi.Name, bi.ID, err)
		} else {
			log.Printf("Created pull image job for base image: %s [%v]", bi.Name, bi.ID)
		}
	}
	return nil
}

func (b *BaseImageBuilder) deleteJob(jobName string) error {
	background := metav1.DeletePropagationBackground
	err := b.clientset.BatchV1().Jobs(b.namespace).Delete(jobName, &metav1.DeleteOptions{PropagationPolicy: &background})
	if err != nil {
		return errors.Wrap(err, "Failed to delete job")
	}
	return nil
}

func (b *BaseImageBuilder) jobResult(bi *BaseImage) (bool, error) {
	name := fmt.Sprintf("create-base-image-%v", bi.ID)
	job, err := b.clientset.BatchV1().Jobs(b.namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return false, errors.Wrap(err, "Failed to get job")
	}
	podList, err := b.clientset.CoreV1().Pods(b.namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s", job.Name)})
	if err != nil {
		return false, errors.Wrap(err, "Failed to list pods associated with job")
	}
	if len(podList.Items) > 0 {
		pod := podList.Items[len(podList.Items)-1]
		req := b.clientset.CoreV1().Pods("default").GetLogs(pod.Name, &apiv1.PodLogOptions{})
		bytes, err := req.DoRaw()
		if err != nil {
			return false, errors.Wrap(err, "Request for logs (result) failed")
		}
		var result imageStatusResult
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return false, errors.Wrap(err, "Failed to parse logs into result")
		}
		return result.Result == 0, nil
	}
	return false, nil
}

func (b *BaseImageBuilder) jobLog(bi *BaseImage) error {
	name := fmt.Sprintf("create-base-image-%v", bi.ID)
	job, err := b.clientset.BatchV1().Jobs(b.namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error getting job %s", job.Name)
	}
	podList, err := b.clientset.CoreV1().Pods(b.namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s", job.Name)})
	if len(podList.Items) > 0 {
		pod := podList.Items[len(podList.Items)-1]
		req := b.clientset.CoreV1().Pods("default").GetLogs(pod.Name, &apiv1.PodLogOptions{})
		resp := req.Do()
		bytes, err := resp.Raw()
		if err != nil {
			return errors.Wrapf(err, "Error getting logs for job %s", job.Name)
		}
		bi.Reason = []string{string(bytes)}
		log.Printf("Logs for job %s: %v\n", job.Name, string(bytes))
	}
	return nil
}

func (b *BaseImageBuilder) watch() error {
	watch, err := b.clientset.BatchV1().Jobs(b.namespace).Watch(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to create watch")
	}
	for {
		select {
		case w := <-watch.ResultChan():
			retJob, ok := w.Object.(*batchv1.Job)
			if !ok {
				log.Printf("Wrong kind returned from base image builder watch: %v", w.Object.GetObjectKind())
				continue
			}
			if w.Type == watchv1.Deleted {
				continue
			}
			var baseImage BaseImage
			err := b.es.GetById(b.orgID, retJob.Labels["baseImageID"], &baseImage)
			if err != nil {
				log.Printf("Error fetching image %s from entity store: %v", retJob.Labels["baseImageID"], err)
				continue
			}
			if retJob.Status.Active > 0 {
				baseImage.Status = StatusCREATING
			} else if retJob.Status.Failed > 0 {
				baseImage.Status = StatusERROR
				err = b.jobLog(&baseImage)
				if err != nil {
					log.Printf("Error getting logs for failed job: %v", err)
					continue
				}
				err = b.deleteJob(retJob.Name)
				if err != nil {
					log.Printf("Error deleting failed job: %v", err)
				}
			} else if retJob.Status.Succeeded > 0 {
				success, err := b.jobResult(&baseImage)
				if success {
					baseImage.Status = StatusREADY
				} else {
					baseImage.Status = StatusERROR
				}
				err = b.deleteJob(retJob.Name)
				if err != nil {
					log.Printf("Error deleting successful job: %v", err)
				}
			} else {
				continue
			}
			_, err = b.es.Update(baseImage.Revision, &baseImage)
			if err != nil {
				log.Printf("Error updating image %s to entity store: %v", retJob.Labels["baseImageID"], err)
				continue
			}
		case bi := <-b.baseImageChannel:
			log.Printf("Received base image update %s", bi.Name)
			_, err := b.createPullJob(&bi)
			if err != nil {
				log.Printf("Error creating job for base image: %s [%v]: %v", bi.Name, bi.ID, err)
				continue
			}
		case <-time.After(120 * time.Second):
			b.poll()
		}
	}
}

func (b *BaseImageBuilder) Run() {
	go b.watch()
	<-b.done
}

func (b *BaseImageBuilder) Shutdown() {
	b.done <- true
}
