package main

import (
	"fmt"
	gateways "kube-workers-manager/gateways"
	usecases "kube-workers-manager/usecases"
)

func main() {

	config, err := usecases.ReadConfig()
	if err != nil {
		fmt.Errorf("Error while trying to read service config: %v\n", err.Error())
	}

	gateways.StartServer(config)

}
