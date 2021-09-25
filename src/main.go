package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	httpProxy "gopkg.in/elazarl/goproxy.v1"
)

func main() {

	log.Print("Reading startup information")

	daprdArgs := os.Args[1:]
	appRedirectAddress := getAppRedirectAddress()
	appListeningAddress := getAppListeningAddress(daprdArgs)
	consulRedirectAddress := getConsulRedirectAddress()
	consulListeningAddress := getConsulListeningAddress()
	isDebug := getIsDebugMode()
	debugCommand := getDebugCommand()
	appProxyStrategy := getAppProxyStrategy()
	consulProxyStrategy := getConsulProxyStrategy()
	appConsulAddress := getAppConsulAddress()

	log.Printf("daprdArgs: %s", daprdArgs)
	log.Printf("appRedirectAddress: %s", appRedirectAddress)
	log.Printf("appListeningAddress: %s", appListeningAddress)
	log.Printf("consulRedirectAddress: %s", consulRedirectAddress)
	log.Printf("consulListeningAddress: %s", consulListeningAddress)
	log.Printf("appConsulAddress: %s", appConsulAddress)
	log.Printf("isDebug: %t", isDebug)
	log.Printf("debugCommand: %s", debugCommand)
	log.Printf("appProxyStrategy: %s", appProxyStrategy)
	log.Printf("consulProxyStrategy: %s", consulProxyStrategy)

	log.Print("Start")

	var wg sync.WaitGroup

	goLaunch(&wg, func() { launchDaprd(isDebug, debugCommand, daprdArgs) })
	goLaunch(&wg, func() { createAppProxy(appProxyStrategy, appListeningAddress, appRedirectAddress) })
	goLaunch(&wg, func() {
		createConsulProxy(consulProxyStrategy, consulListeningAddress, consulRedirectAddress, appConsulAddress)
	})

	log.Print("Waiting for termination")
	wg.Wait()
}

func createProxy(proxyStrategy string, listeningAddress string, redirectAddress string) *httpProxy.ProxyHttpServer {
	switch proxyStrategy {
	case "HTTP":
		return createHttpProxy(listeningAddress, redirectAddress, "http")
	case "HTTPS":
		return createHttpProxy(listeningAddress, redirectAddress, "https")
	default:
		log.Fatal("Invalid Proxy Strategy")
		return nil
	}
}

func createHttpProxy(listeningAddress string, redirectAddress string, schema string) *httpProxy.ProxyHttpServer {
	log.Printf("Start HTTP Proxy: Redirect  %s to %s", listeningAddress, redirectAddress)

	proxy := httpProxy.NewProxyHttpServer()
	proxy.Verbose = true

	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Host = redirectAddress
		req.URL.Scheme = schema

		req.Host = redirectAddress
		log.Print("CCCCCC")
		proxy.ServeHTTP(w, req)
	})
	proxy.OnRequest().HandleConnect(httpProxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *httpProxy.ProxyCtx) (*http.Request, *http.Response) {
		req.URL.Host = redirectAddress
		req.URL.Scheme = schema
		req.Host = redirectAddress
		log.Print("BBBBB")
		return req, nil
	})
	return proxy
	//log.Fatal(http.ListenAndServe(listeningAddress, proxy))
}

func launchDaprd(isDebug bool, debugCommand string, daprdArgs []string) {
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
	daprd.Wait()
	log.Print("Daprd Close")
}

func createAppProxy(proxyStrategy string, listeningAddress string, redirectAddress string) {

	if listeningAddress != "" {
		proxy := createProxy(proxyStrategy, listeningAddress, redirectAddress)
		log.Fatal(http.ListenAndServe(listeningAddress, proxy))
		log.Print("App Proxy Closed")
	}
}

func createConsulProxy(proxyStrategy string, listeningAddress string, redirectAddress string, serviceAddress string) {
	if redirectAddress != "" {
		proxy := createProxy(proxyStrategy, listeningAddress, redirectAddress)
		proxy.OnRequest().DoFunc(func(req *http.Request, ctx *httpProxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Print("AAAAA")

			if req.URL.Path == "/v1/agent/service/register" {
				byteData := readReadCloser(req.Body)
				jsonData := extractJson(byteData)
				jsonData = changeServiceRegistrationJson(jsonData, serviceAddress)
				byteData = tryConvertJson(jsonData, byteData)

				var byteDataLen int64 = 0
				var newReadCloser io.ReadCloser = nil
				if byteData != nil {
					newReadCloser = io.NopCloser(bytes.NewReader(byteData))
					byteDataLen = int64(len(byteData))
				}
				req.Body = newReadCloser
				req.ContentLength = byteDataLen

			}

			return req, nil
		})

		log.Fatal(http.ListenAndServe(listeningAddress, proxy))
		log.Print("Consul Proxy Closed")
	}
}

func changeServiceRegistrationJson(json map[string]interface{}, serviceAddress string) map[string]interface{} {
	if json != nil {
		log.Printf("Original service address: %s", json["Address"])
		json["Address"] = serviceAddress
		log.Printf("New service address: %s", json["Address"])
		return json
	}
	return nil
}

func tryConvertJson(jsonData map[string]interface{}, onErrorData []byte) []byte {
	newBody, err := json.Marshal(jsonData)
	if newBody != nil && err == nil {
		return newBody
	}
	return onErrorData
}
