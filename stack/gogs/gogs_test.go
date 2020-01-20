package gogs

import (
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func TestDeploymentConfig(t *testing.T) {
	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	err := e.Encode(Deployment("image", "bedrock"), os.Stdout)
	if err != nil {
		t.Fatal("Error generating YAML", err)
	}
}

func TestServiceConfig(t *testing.T) {
	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	err := e.Encode(Service("bedrock"), os.Stdout)
	if err != nil {
		t.Fatal("Error generating YAML", err)
	}
}
