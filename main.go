package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Worker struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace,omitempty"`
	Image           string `json:"image"`
	Port            int    `json:"port"`
	TrainingDataDir string `json:"trainingDir"`
	clientset       *kubernetes.Clientset
}

func (w *Worker) GenerateWorkerName(count int) {
	w.Name = fmt.Sprintf("ofl-worker-%d", count)
}

func (w *Worker) GenerateWorkerPort(count int) {
	w.Port = 9000 + count
}

func (w *Worker) CreateDeployment() {
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
							Args:            []string{fmt.Sprintf("/data/dataset/%s", w.TrainingDataDir)},
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

func (w *Worker) CreateConfigMap(configMapData string) {
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

func (w *Worker) CreateService() {
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
		fmt.Printf("Error creating service: %v\n", err)
	} else {
		fmt.Printf("Service created: %s\n", service.Name)
	}
}

func main() {
	// Load kubeconfig
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Build the clientset
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	worker := &Worker{
		Name:            "ofl-worker-1",
		Namespace:       "default",
		Image:           "ofl-worker:latest",
		Port:            9001,
		TrainingDataDir: "trainingSetWorker1",
		clientset:       clientset,
	}

	cmData := `{
      "port": 9001,
      "upload_folder": "/data"
    }`

	worker.CreateConfigMap(cmData)
	worker.CreateDeployment()
	worker.CreateService()

	// Create a namespace
	// namespace := "ofl-namespace"
	// _, err = clientset.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: namespace,
	// 	},
	// }, metav1.CreateOptions{})
	// if err != nil {
	// 	fmt.Errorf("Error creating namespace: %v\n", err)
	// } else {
	// 	fmt.Printf("Namespace created: %s\n", namespace)
	// }

}

func int32Ptr(i int32) *int32 { return &i }
