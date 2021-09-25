package main

import (
	"fmt"
	"os"
	"strings"
)

func getAppListeningAddress(daprdArgs []string) string {
	argsLen := len(daprdArgs)
	indexOfAppPort := indexOf(daprdArgs, "--app-port")
	if indexOfAppPort >= 0 && indexOfAppPort < argsLen-1 {
		appPort := daprdArgs[indexOfAppPort+1]
		return fmt.Sprintf("127.0.0.1:%s", appPort)
	}
	return ""
}

func getAppRedirectAddress() string {
	return os.Getenv("REMOTE_DAPRD_ADDRESS")
}

func getConsulListeningAddress() string {
	return "127.0.0.1:8500"
}

func getConsulRedirectAddress() string {
	return os.Getenv("REMOTE_CONSUL_ADDRESS")
}

func getIsDebugMode() bool {
	return os.Getenv("IS_DEBUG") == "1"
}

func getDebugCommand() string {
	return os.Getenv("GET_DEBUG_COMMAND")
}

func getAppProxyStrategy() string {
	value := os.Getenv("APP_PROXY_STRATEGY")
	if value == "" {
		value = "HTTP"
	}
	return value
}

func getConsulProxyStrategy() string {
	value := os.Getenv("CONSUL_PROXY_STRATEGY")
	if value == "" {
		value = "HTTP"
	}
	return value
}

func getAppConsulAddress() string {
	appAddress := getAppRedirectAddress()
	colonIndex := strings.LastIndex(appAddress, ":")
	if colonIndex < 0 {
		return appAddress
	}
	return appAddress[:colonIndex]
}
