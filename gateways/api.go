package gateways

import (
	"encoding/json"
	"fmt"
	usecases "kube-workers-manager/usecases"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	backendBaseRoute = "/worker-manager/"

	createWorkers = "createWorkers"
	deleteWorkers = "deleteWorkers"
)

func NewAPI() *mux.Router {

	var router = mux.NewRouter()
	router.Use(commonMiddleware)
	router.Use(mux.CORSMethodMiddleware(router))

	router.HandleFunc(backendBaseRoute+createWorkers, CreateWorkers).Methods("POST")
	router.HandleFunc(backendBaseRoute+deleteWorkers, DeleteWorkers).Methods("DELETE")
	//router.HandleFunc(backendBaseRoute+createWorkers+"/{worker_id}", GetWorkerById).Methods("GET")

	return router
}

func StartServer(config *usecases.Config) {

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Origin", "application/json"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"POST", "DELETE"})

	router := NewAPI()

	fmt.Printf("HTTP Server is running on port %s\n", config.ManagerPort)
	log.Fatal(http.ListenAndServe(":"+config.ManagerPort, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func optionsHandler(_ http.ResponseWriter, _ *http.Request) {
	return
}

func CreateWorkers(w http.ResponseWriter, r *http.Request) {
	var workersDetails *usecases.WorkerApi = &usecases.WorkerApi{}

	err := json.NewDecoder(r.Body).Decode(workersDetails)
	if err != nil {
		http.Error(w, "Bad Request "+err.Error(), http.StatusBadRequest)
	}

	err = usecases.CreateKubeObjects(workersDetails)
	if err != nil {
		json.NewEncoder(w).Encode(fmt.Errorf("Error creating workers: %v\n", err.Error()))
	}

	json.NewEncoder(w).Encode(fmt.Sprintf("Workers created successfully\n"))

}

func DeleteWorkers(w http.ResponseWriter, r *http.Request) {

}
