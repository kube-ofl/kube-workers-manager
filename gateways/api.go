package gateways

import (
	"encoding/json"
	"fmt"
	"io"
	usecases "kube-workers-manager/usecases"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	backendBaseRoute = "/worker-manager"

	createWorkers = "/create-workers"
	deleteWorkers = "/delete-workers"
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
	fmt.Println(fmt.Sprintf("%s %s", r.Method, r.URL))

	// Read the body into a buffer
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}

	// Print the raw body as a string
	fmt.Println("Request Body:", string(bodyBytes))

	// err = json.NewDecoder(bodyBytes).Decode(workersDetails)
	err = json.Unmarshal(bodyBytes, workersDetails)
	if err != nil {
		http.Error(w, "Bad Request: cannot unmarshal the body "+err.Error(), http.StatusBadRequest)
		fmt.Println("Cannot unmarshal workersDetails struct", err.Error())
		return
	}

	fmt.Printf("Workers details: %v\n", workersDetails)

	err = usecases.CreateKubeObjects(workersDetails)
	if err != nil {
		http.Error(w, "Error creating workers: "+err.Error(), http.StatusBadRequest)
		// json.NewEncoder(w).Encode(fmt.Errorf("Error creating workers: %v\n", err.Error()))
		fmt.Println("Cannot create workers: ", err.Error())
		return
	}

	fmt.Println("Request handled successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fmt.Sprintf("Workers created successfully"))
}

func DeleteWorkers(w http.ResponseWriter, r *http.Request) {

}
