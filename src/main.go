package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	tcpproxy "github.com/inetaf/tcpproxy"

	httpProxy "gopkg.in/elazarl/goproxy.v1"
)

func main() {

	log.Print("Reading startup information")

	daprdArgs := os.Args[1:]
	redirectAddress := getRedirectAddress()
	listeningAddress := getListeningAddress(daprdArgs)
	isDebug := getIsDebugMode()
	debugCommand := getDebugCommand()
	proxyStrategy := getProxyStrategy()

	log.Printf("daprdArgs: %s", daprdArgs)
	log.Printf("redirectAddress: %s", redirectAddress)
	log.Printf("listeningAddress: %s", listeningAddress)
	log.Printf("isDebug: %t", isDebug)
	log.Printf("debugCommand: %s", debugCommand)
	log.Printf("proxyStrategy: %s", proxyStrategy)

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
		daprd.Stderr = os.Stderr
		err := daprd.Start()
		if err != nil {
			log.Fatal(err)
		}
	}

	if listeningAddress != "" {
		createProxy(listeningAddress, redirectAddress)
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
		return fmt.Sprintf("0.0.0.0:%s", appPort)
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

func getProxyStrategy() string {
	value := os.Getenv("PROXY_STRATEGY")
	if value == "" {
		value = "HTTP"
	}
	return value
}

func createProxy(listeningAddress string, redirectAddress string) {
	switch getProxyStrategy() {
	case "HTTP":
		createHttpProxy(listeningAddress, redirectAddress, "http")
	case "HTTPS":
		createHttpProxy(listeningAddress, redirectAddress, "https")
	case "TCP":
		createTcpProxy(listeningAddress, redirectAddress)
	default:
		log.Fatal("Invalid Proxy Strategy")
	}
}

func createTcpProxy(listeningAddress string, redirectAddress string) {
	log.Printf("Start Tcp Proxy: Redirect  %s to %s", listeningAddress, redirectAddress)

	var proxy tcpproxy.Proxy
	proxy.AddRoute(listeningAddress, tcpproxy.To(redirectAddress)) // fallback
	defer proxy.Close()
	log.Fatal(proxy.Run())
}

func createHttpProxy(listeningAddress string, redirectAddress string, schema string) {
	log.Printf("Start HTTP Proxy: Redirect  %s to %s", listeningAddress, redirectAddress)

	proxy := httpProxy.NewProxyHttpServer()
	proxy.Verbose = true

	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Host = redirectAddress
		req.URL.Scheme = schema

		req.Host = redirectAddress
		req.Header.Set("X-Host", redirectAddress)
		proxy.ServeHTTP(w, req)
	})
	proxy.OnRequest().HandleConnect(httpProxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *httpProxy.ProxyCtx) (*http.Request, *http.Response) {
		req.URL.Host = redirectAddress
		req.URL.Scheme = schema
		req.Host = redirectAddress
		req.Header.Set("X-Host", redirectAddress)
		return req, nil
	})
	log.Fatal(http.ListenAndServe(listeningAddress, proxy))
}
