# kube-workers-manager
This service creates kubernetes resources based on requests from the api server through api calls

- These are parameters that can be set for the workers through the manager's API:
```
	WorkersNo       int    `json:"workersNo"`
	TrainingDataDir string `json:"trainingDataDir`
	Namespace       string `json:"namespace,omitempty"`
	Image           string `json:"image,omitempty"`
	BasePort        int    `json:"basePort,omitempty"`
	UploadFolder    string `json:"uploadFolder,omitempty"`
```
- These are the defaults for the fields that can be empty:

```
namespace: default
Image: ofl-worker:latest
BasePort: 9000
UploadFolder: /data
```

## Examples of API calls:

- the most basic(and probably the most used)
```
POST /worker-manager/createWorkers
{ 
    "workersNo": 4,
    "trainingDataDir": "trainingDataDir1"
}
```
```
curl -X POST http://localhost:80/worker-manager/createWorkers -H "Content-Type: application/json" -d '{"workersNo": 4, "trainingDataDir": "trainingDataDir1"}'
curl -X POST http://worker-manager-service.svc.cluster.local/worker-manager/createWorkers -H "Content-Type: application/json" -d '{"workersNo": 4, "trainingDataDir": "trainingDataDir1"}'
```


- other example:
POST /worker-manager/createWorkers
```
{ 
    "workersNo": 4,
    "trainingDataDir": "trainingDataDir1",
    "namespace": "kube-ofl",
    "image": "ofl-worker:v1",
    "basePort": "6000",
    "uploadFolder": "/data5"
}
```