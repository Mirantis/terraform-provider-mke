package client

/**
Public Key abstractions

@NOTE this should be implemented using a direct import of enzi client code
	@see https://github.com/Mirantis/orca/blob/master/enzi/schema/account_public_keys.go#L18
	@see https://github.com/Mirantis/orca/blob/master/enzi/api/client/client.go
*/

// AccountPublicKey api interpretation of a public key
// @see https://github.com/Mirantis/orca/blob/c999ce63c591beba851926500c3d66f9af3cb244/enzi/api/responses/responses.go#L1058
type AccountPublicKey struct {
	ID           string        `json:"id"        description:"the hash of the public key's DER bytes"`
	AccountID    string        `json:"accountID" description:"the ID of the account"`
	PublicKey    string        `json:"publicKey" description:"the encoded PEM of the public key"`
	Label        string        `json:"label"     description:"the label or description for the key"`
	Certificates []Certificate `json:"certificates,omitempty" description:"certificates for the public key"`
}

// Certificate is a sub-form for the account certificate.
// @see https://github.com/Mirantis/orca/blob/master/enzi/api/responses/responses.go#L1103
type Certificate struct {
	Label string `json:"label" description:"Label for the certificate"`
	Cert  string `json:"cert"  description:"Encoded PEM for the cert"`
}
