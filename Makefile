.PHONY: run minikube

run :
	go test -cover github.com/xumak-grid/bedrock
	go test -cover github.com/xumak-grid/bedrock/cmd/api
	go test -cover github.com/xumak-grid/bedrock/http
	go test -cover github.com/xumak-grid/bedrock/k8s
	go test -cover github.com/xumak-grid/bedrock/stack/drone
	go test -cover github.com/xumak-grid/bedrock/stack/gogs
	go test -cover github.com/xumak-grid/bedrock/stack/nexus
	DEVELOPMENT=false KUBECONFIG=$(HOME)/.kube/config gin -a 8000 -b bin/api -i --build cmd/api

minikube:
	minikube start --vm-driver xhyve --insecure-registry your.registry:5000

default : run
