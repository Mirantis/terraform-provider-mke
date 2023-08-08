package client

import (
	"encoding/base64"
	"encoding/json"
	"io"

	"gopkg.in/yaml.v2"
)

const (
	ClientBundleMetaEndpointDocker     = "docker"
	ClientBundleMetaEndpointKubernetes = "kubernetes"
)

// ClientBundle interpretation of the ClientBundle data in memory.
type ClientBundle struct {
	ID         string            `json:"id"`
	PrivateKey string            `json:"private_key"`
	PublicKey  string            `json:"public_key"`
	Cert       string            `json:"cert"`
	CACert     string            `json:"ca_cert"`
	Kube       *ClientBundleKube `json:"kube"` // There is not always a kube config
	Meta       ClientBundleMeta  `json:"meta"`
}

// ToJSON stringify client bundle as json for debugging.
func (cb ClientBundle) ToJSON() []byte {
	b, _ := json.Marshal(cb)
	return b
}

// ClientBundleKube Kubernetes parts of the client bundle.
// primarily we are focused on satisfying requirements for a kubernetes provider.
// such as https://github.com/hashicorp/terraform-provider-kubernetes/blob/main/kubernetes/provider.go
type ClientBundleKube struct {
	Config            string `json:"config"`
	Host              string `json:"host"`
	ClientKey         string `json:"client_key"`
	ClientCertificate string `json:"client_certificate"`
	CACertificate     string `json:"cluster_ca_certificate"`
	Insecure          string `json:"insecure"`
}

// ClientBundleMeta in the client bundle is a flattenned meta.json file.  It is buried kind of deep.
type ClientBundleMeta struct {
	Name              string `json:"Name"`
	Description       string `json:"Description"`
	StackOrchestrator string `json:"StackOrchestrator"`

	DockerHost              string `json:"DockerHost"`
	DockerSkipVerifyTLS     bool   `json:"DockerSkipVerifyTLS"`
	KubernetesHost          string `json:"KubernetesHost"`
	KubernetesSkipVerifyTLS bool   `json:"KubernetesSkipVerifyTLS"`
}

// ClientBundleRetrieveValue read a value.
func ClientBundleRetrieveValue(val io.Reader) (string, error) {
	allBytes, err := io.ReadAll(val)
	if err != nil {
		return "", err
	}
	return string(allBytes), nil
}

// ClientBundleDecodeBase64Value read a value and base64 decode it.
func ClientBundleDecodeBase64Value(val io.Reader) (string, error) {
	allBytes, err := io.ReadAll(val)
	if err != nil {
		return "", err
	}
	return helperStringBase64Decode(string(allBytes)), nil
}

// NewClientBundleKubeFromKubeYml ClientBundleKube constructor from byte list of a kubeconfig file.
func NewClientBundleKubeFromKubeYml(val io.Reader) (ClientBundleKube, error) {
	k8bytes, _ := io.ReadAll(val)

	var cbk ClientBundleKube

	// Struct representation of a kube config file.
	// see https://zhwt.github.io/yaml-to-go/
	var cbkHolder struct {
		APIVersion  string            `yaml:"apiVersion"`
		Kind        string            `yaml:"kind"`
		Preferences map[string]string `yaml:"preferences"`
		Clusters    []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				CertificateAuthorityData string `yaml:"certificate-authority-data"`
				Server                   string `yaml:"server"`
			} `yaml:"cluster"`
		} `yaml:"clusters"`
		Contexts []struct {
			Name    string `yaml:"name"`
			Context struct {
				Cluster string `yaml:"cluster"`
				User    string `yaml:"user"`
			} `yaml:"context"`
		} `yaml:"contexts"`
		CurrentContext string `yaml:"current-context"`
		Users          []struct {
			Name string `yaml:"name"`
			User struct {
				ClientCertificateData string `yaml:"client-certificate-data"`
				ClientKeyData         string `yaml:"client-key-data"`
			} `yaml:"user"`
		} `yaml:"users"`
	}

	cbk.Config = string(k8bytes)

	if err := yaml.UnmarshalStrict(k8bytes, &cbkHolder); err != nil {
		return cbk, err
	}

	var contextName, clusterName, userName string

	contextName = cbkHolder.CurrentContext

	for _, context := range cbkHolder.Contexts {
		if context.Name == contextName {
			clusterName = context.Context.Cluster
			userName = context.Context.User
			break
		}
	}

	for _, cluster := range cbkHolder.Clusters {
		if cluster.Name == clusterName {
			cbk.Host = cluster.Cluster.Server
			cbk.CACertificate = helperStringBase64Decode(cluster.Cluster.CertificateAuthorityData)
			break
		}
	}

	for _, user := range cbkHolder.Users {
		if user.Name == userName {
			cbk.ClientKey = helperStringBase64Decode(user.User.ClientKeyData)
			cbk.ClientCertificate = helperStringBase64Decode(user.User.ClientCertificateData)
			break
		}
	}

	return cbk, nil
}

// NewClientBundleMetaFromReader interpret the meta.json file reader as a Meta struct
func NewClientBundleMetaFromReader(val io.Reader) (ClientBundleMeta, error) {
	var cbm ClientBundleMeta
	// Struct representation of the meta.json file
	var cbmHolder struct {
		Name     string `json:"Name"`
		Metadata struct {
			Description       string `json:"Description"`
			StackOrchestrator string `json:"StackOrchestrator"`
		} `json:"Metadata"`
		Endpoints map[string]struct {
			Host          string `json:"Host"`
			SkipTLSVerify bool   `json:"SkipTLSVerify"`
		} `json:"Endpoints"`
	}

	b, _ := io.ReadAll(val)

	if err := json.Unmarshal(b, &cbmHolder); err != nil {
		return cbm, err
	}

	cbm.Name = cbmHolder.Name
	cbm.Description = cbmHolder.Metadata.Description
	cbm.StackOrchestrator = cbmHolder.Metadata.StackOrchestrator

	if ke, ok := cbmHolder.Endpoints[ClientBundleMetaEndpointKubernetes]; ok {
		cbm.KubernetesHost = ke.Host
		cbm.KubernetesSkipVerifyTLS = ke.SkipTLSVerify
	}
	if de, ok := cbmHolder.Endpoints[ClientBundleMetaEndpointDocker]; ok {
		cbm.DockerHost = de.Host
		cbm.DockerSkipVerifyTLS = de.SkipTLSVerify
	}

	return cbm, nil
}

// this decodes some strings in the file that are base64 encoded.
func helperStringBase64Decode(val string) string {
	valDecodedBytes, _ := base64.StdEncoding.DecodeString(val)
	return string(valDecodedBytes)
}
