package client

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	URLTargetForClientBundle           = "api/clientbundle"
	URLTargetForClientBundleQueryLabel = "label"

	filenameCAPem             = "ca.pem"
	filenameCertPem           = "cert.pem"
	filenamePrivKeyPem        = "key.pem"
	filenamePubKeyPem         = "cert.pub"
	filenameKubeconfig        = "kube.yml"
	filenameDockerBundleZip   = "ucp-docker-bundle.zip" // there is a zip file in the zip file.
	filenameDockerBundlerMeta = "meta.json"
)

var (
	ErrFailedToRetrieveClientBundle         = errors.New("failed to retrieve the client bundle from MKE")
	ErrFailedToFindClientBundleMKEPublicKey = errors.New("no MKE Public key was found that matches the client bundle")
)

// ApiClientBundle retrieve a client bundle.
func (c *Client) ApiClientBundleCreate(ctx context.Context, label string) (ClientBundle, error) {
	var cb ClientBundle

	req, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodPost, URLTargetForClientBundle, []byte{})
	if err != nil {
		return cb, err
	}

	// @SEE https://github.com/Mirantis/orca/blob/master/controller/api/bundle.go#L35 for source.
	q := req.URL.Query()
	q.Add(URLTargetForClientBundleQueryLabel, label)
	req.URL.RawQuery = q.Encode()

	resp, err := c.doAuthorizedRequest(req)
	if err != nil {
		return cb, err
	}
	defer resp.Body.Close()

	zb, err := io.ReadAll(resp.Body)
	if err != nil {
		return cb, err
	}

	zr, err := zip.NewReader(bytes.NewReader(zb), resp.ContentLength)
	if err != nil {
		return cb, err
	}

	cb.ID = zr.Comment

	errs := []error{}

	for _, f := range zr.File {
		switch f.Name {
		case filenameCAPem:
			fr, _ := f.Open()
			capem, err := ClientBundleRetrieveValue(fr)
			fr.Close()

			if err != nil {
				errs = append(errs, err)
			} else {
				cb.CACert = capem
			}
		case filenameCertPem:
			fr, _ := f.Open()
			cert, err := ClientBundleRetrieveValue(fr)
			fr.Close()

			if err != nil {
				errs = append(errs, err)
			} else {
				cb.Cert = cert
			}
		case filenamePrivKeyPem:
			fr, _ := f.Open()
			prkpem, err := ClientBundleRetrieveValue(fr)
			fr.Close()

			if err != nil {
				errs = append(errs, err)
			} else {
				cb.PrivateKey = prkpem
			}
		case filenamePubKeyPem:
			fr, _ := f.Open()
			pbkpem, err := ClientBundleRetrieveValue(fr)
			fr.Close()

			if err != nil {
				errs = append(errs, err)
			} else {
				cb.PublicKey = pbkpem
			}
		case filenameKubeconfig:
			fr, _ := f.Open()
			kube, err := NewClientBundleKubeFromKubeYml(fr)
			fr.Close()

			if err != nil {
				errs = append(errs, err)
			} else {
				cb.Kube = &kube
			}
		case filenameDockerBundleZip:
			fr, _ := f.Open()
			dbzb, err := io.ReadAll(fr)
			fr.Close()

			if err != nil {
				errs = append(errs, err)
				continue
			}

			dbzr, err := zip.NewReader(bytes.NewReader(dbzb), resp.ContentLength)
			if err != nil {
				return cb, err
			}

			if err != nil {
				errs = append(errs, err)
				continue
			}

			for _, f := range dbzr.File {
				switch f.Name {
				case filenameDockerBundlerMeta:
					mfr, _ := f.Open()
					cbm, err := NewClientBundleMetaFromReader(mfr)
					mfr.Close()

					if err != nil {
						errs = append(errs, err)
						continue
					}

					if cbm.Name != "" {
						cb.ID = cbm.Name
					}
					cb.Meta = cbm
				}
			}
		}
	}

	if len(errs) > 0 {
		errString := ""

		for _, err := range errs {
			errString = fmt.Sprintf("%s, %s", errString, err)
		}

		return cb, fmt.Errorf("%w; %s", ErrFailedToRetrieveClientBundle, errString)
	}

	return cb, nil
}

// ApiClientBundleGetPublicKey retrieve a client bundle by finding the matching public key.
// There isn't really a great way of doing this.
func (c *Client) ApiClientBundleGetPublicKey(ctx context.Context, cb ClientBundle) (AccountPublicKey, error) {
	var k AccountPublicKey

	account := c.Username()

	keys, err := c.ApiPublicKeyList(ctx, account)
	if err != nil {
		return k, err
	}

	foundKeys := []string{}
	cbpk := strings.TrimSpace(cb.PublicKey)
	for _, key := range keys {
		pk := strings.TrimSpace(key.PublicKey)
		if pk == cbpk {
			return key, nil
		}
		foundKeys = append(foundKeys, pk)
	}

	return k, fmt.Errorf("%w; Could not match key: \n%s\n in \n%s", ErrFailedToFindClientBundleMKEPublicKey, cb.PublicKey, strings.Join(foundKeys, "\n"))
}

// ApiClientBundleDelete delete a client bundle by finding and deleting the matching public key.
// There isn't really a great way of doing this.
func (c *Client) ApiClientBundleDelete(ctx context.Context, cb ClientBundle) error {
	account := c.Username()

	key, err := c.ApiClientBundleGetPublicKey(ctx, cb)
	if err != nil {
		return err
	}

	return c.ApiPublicKeyDelete(ctx, account, key.ID)
}
