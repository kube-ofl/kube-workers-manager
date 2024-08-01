package main

import (
	"fmt"
	gateways "kube-workers-manager/gateways"
	usecases "kube-workers-manager/usecases"
)

func main() {

	config, err := usecases.ReadConfig()
	if err != nil {
		panic(fmt.Errorf("Error while trying to read service config: %v", err.Error()))
	}

	gateways.StartServer(config)

}
