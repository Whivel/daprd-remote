package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	tcpproxy "github.com/inetaf/tcpproxy"
)

func main() {

	log.Print("Reading startup information")

	daprdArgs := os.Args[1:]
	redirectAddress := getRedirectAddress()
	listeningAddress := getListeningAddress(daprdArgs)
	isDebug := getIsDebugMode()
	debugCommand := getDebugCommand()

	log.Printf("daprdArgs: %s", daprdArgs)
	log.Printf("redirectAddress: %s", redirectAddress)
	log.Printf("listeningAddress: %s", listeningAddress)
	log.Printf("isDebug: %t", isDebug)
	log.Printf("debugCommand: %s", debugCommand)

	log.Print("Start")

	var daprd *exec.Cmd = nil
	if isDebug {

		log.Printf("DEBUG: Skipped dapr spawn. Launching: %s", debugCommand)

		daprd = exec.Command(debugCommand)
		daprd.Start()
	} else {

		log.Print("Spawning daprd")
		daprd = exec.Command("/daprd", daprdArgs...)
		daprd.Stdout = os.Stdout
		err := daprd.Start()
		if err != nil {
			log.Fatal(err)
		}
	}

	if listeningAddress != "" {

		log.Printf("Start Proxy: Listening at %s", listeningAddress)

		var proxy tcpproxy.Proxy
		proxy.AddRoute(listeningAddress, tcpproxy.To(redirectAddress)) // fallback
		defer proxy.Close()
		log.Fatal(proxy.Run())
		log.Print("Closed")
	}

	if daprd != nil {
		log.Print("Waiting for daprd termination")
		daprd.Wait()
	}
}

func indexOf(arr []string, value string) int {
	for i, el := range arr {
		if el == value {
			return i
		}
	}
	return -1
}

func getListeningAddress(daprdArgs []string) string {
	argsLen := len(daprdArgs)
	indexOfAppPort := indexOf(daprdArgs, "--app-port")
	if indexOfAppPort >= 0 && indexOfAppPort < argsLen-1 {
		appPort := daprdArgs[indexOfAppPort+1]
		return fmt.Sprintf(":%s", appPort)
	}
	return ""
}

func getRedirectAddress() string {
	return os.Getenv("REMOTE_DAPRD_ADDRESS")
}

func getIsDebugMode() bool {
	return os.Getenv("IS_DEBUG") == "1"
}

func getDebugCommand() string {
	return os.Getenv("GET_DEBUG_COMMAND")
}
