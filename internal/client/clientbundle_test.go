package client_test

import (
	"bytes"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

var (
	// a good kube yaml file
	// @note some values are base64 encoded
	GoodKubeYml = `
apiVersion: v1
kind: Config
preferences: {}

clusters:
- name: 6443_admin_cluster
  cluster:
    certificate-authority-data: RUZHSElK
    server: localhost:6443

contexts:
- name: 6443_admin
  context:
    cluster: 6443_admin_cluster
    user: 6443_admin_user

current-context: 6443_admin
users:
- name: 6443_admin_user
  user:
    client-certificate-data: QUJDREU=
    client-key-data: QkNERUZH
`
)

func TestClientBundleFromKubeYmlGood(t *testing.T) {
	buf := bytes.NewBuffer([]byte(GoodKubeYml))
	cbk, err := client.NewClientBundleKubeFromKubeYml(buf)
	if err != nil {
		t.Fatalf("Error converting Kube CB from yaml: %s", err)
	}

	if cbk.Host != "localhost:6443" {
		t.Errorf("CBK from yaml got the wrong host: %+v", cbk)
	}
	if cbk.ClientKey != "BCDEFG" {
		t.Errorf("CBK from yaml got the wrong ClientKey: %+v", cbk)
	}
	if cbk.ClientCertificate != "ABCDE" {
		t.Errorf("CBK from yaml got the wrong ClientCertificate: %+v", cbk)
	}
	if cbk.CACertificate != "EFGHIJ" {
		t.Errorf("CBK from yaml got the wrong ClientCertificate: %+v", cbk)
	}
}
