package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Worker struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace,omitempty"`
	Image       string `json:"image"`
	Port        int    `json:"port"`
	DatasetPath string `json:"datasetPath"`
	clientset   *kubernetes.Clientset
}

type WorkerApi struct {
	WorkersNo    int    `json:"workersNo"`
	DatasetPath  string `json:"datasetPath"`
	Namespace    string `json:"namespace,omitempty"`
	Image        string `json:"image,omitempty"`
	WorkerPort   int    `json:"workerPort,omitempty"`
	UploadFolder string `json:"uploadFolder,omitempty"`
}

type CmData struct {
	WorkerID     string `json:"workerID"`
	Port         int    `json:"port"`
	UploadFolder string `json:"uploadFolder"`
	DatasetPath  string `json:"datasetPath"`
}

func getWorkerName(count int) string {
	return fmt.Sprintf("worker-%d", count)
}

func (w *Worker) createDeployment() {
	// Create a Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-deployment", w.Name),
			Namespace: w.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": w.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": w.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "worker",
							Image:           w.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"python3", "api.py"},
							//  Args:            []string{fmt.Sprintf("/data/dataset/%s", w.TrainingDataDir)},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/etc/config",
									ReadOnly:  true,
								},
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("%s-configmap", w.Name),
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "config.json",
											Path: "config.json",
										},
									},
								},
							},
						},
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/data",
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := w.clientset.AppsV1().Deployments(w.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Errorf("Error creating deployment: %v\n", err)
	} else {
		fmt.Printf("Deployment created: %s\n", deployment.Name)
	}
}

func (w *Worker) createConfigMap(configMapData string) {
	// Create a ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-configmap", w.Name),
			Namespace: w.Namespace,
		},
		Data: map[string]string{
			"config.json": configMapData,
		},
	}
	_, err := w.clientset.CoreV1().ConfigMaps(w.Namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		fmt.Errorf("Error creating configmap: %v\n", err)
	} else {
		fmt.Printf("ConfigMap created: %s\n", configMap.Name)
	}
}

func (w *Worker) createService() {
	// Create a Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-service", w.Name),
			Namespace: w.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": w.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     int32(w.Port),
					TargetPort: intstr.IntOrString{
						IntVal: int32(w.Port),
					},
				},
			},
		},
	}
	_, err := w.clientset.CoreV1().Services(w.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		fmt.Errorf("Error creating service: %v\n", err)
	} else {
		fmt.Printf("Service created: %s\n", service.Name)
	}
}

func (w *Worker) createNamespace() {
	// Create a namespace
	namespace := "ofl-namespace"
	_, err := w.clientset.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		fmt.Errorf("Error creating namespace: %v\n", err)
	} else {
		fmt.Printf("Namespace created: %s\n", namespace)
	}
}

func int32Ptr(i int32) *int32 { return &i }

func getClusterClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return clientset, nil
}

func CreateKubeObjects(workersDetails *WorkerApi) error {

	clientset, err := getClusterClientSet()
	if err != nil {
		return fmt.Errorf("Error getting clientset: %v\n", err)
	}

	// err = checkWorkerDetailsValidity()
	// if err != nil {
	// 	return fmt.Errorf("Invalid Worker Details: %v", err.Error())
	// }

	if workersDetails.WorkersNo <= 0 {
		return fmt.Errorf("Invalid number of workers\n")
	}

	if workersDetails.DatasetPath != "" {
		return fmt.Errorf("DatasetPath not specified")
	}

	ns := "default"
	workerPort := 9000
	image := "ofl-worker:latest"
	uploadFolder := "/data"

	if workersDetails.Namespace != "" {
		ns = workersDetails.Namespace
	}
	if workersDetails.WorkerPort > 0 {
		workerPort = workersDetails.WorkerPort
	}
	if workersDetails.Image != "" {
		image = workersDetails.Image
	}

	if workersDetails.UploadFolder != "" {
		uploadFolder = workersDetails.UploadFolder
	}

	for i := 1; i <= workersDetails.WorkersNo; i++ {

		worker := &Worker{
			Name:        getWorkerName(i),
			Namespace:   ns,
			Image:       image,
			Port:        workerPort,
			DatasetPath: workersDetails.DatasetPath,
			clientset:   clientset,
		}

		configMapData := &CmData{
			WorkerID:     fmt.Sprintf("worker-%d", i),
			Port:         workerPort,
			UploadFolder: uploadFolder,
			DatasetPath:  workersDetails.DatasetPath,
		}

		cmData, err := json.Marshal(configMapData)
		if err != nil {
			return fmt.Errorf("Invalid ConfigMap data for worker: %v\n", err)
		}

		worker.createConfigMap(string(cmData))
		worker.createDeployment()
		worker.createService()
	}

	return nil

}
