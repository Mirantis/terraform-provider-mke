package client

import (
	"context"
	"fmt"
	"net/http"
)

// CreateAccount struct.
type CreateAccount struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	Password   string `json:"password"`
	FullName   string `json:"fullName,omitempty"`
	IsActive   bool   `json:"isActive,omitempty"`
	IsAdmin    bool   `json:"isAdmin,omitempty"`
	IsOrg      bool   `json:"isOrg,omitempty"`
	SearchLDAP bool   `json:"searchLDAP,omitempty"`
}

// UpdateAccount struct.
type UpdateAccount struct {
	FullName string `json:"fullName,omitempty"`
	IsActive bool   `json:"isActive,omitempty"`
	IsAdmin  bool   `json:"isAdmin,omitempty"`
}

// ResponseAccount struct.
type ResponseAccount struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	FullName     string `json:"fullName,omitempty"`
	IsActive     bool   `json:"isActive"`
	IsAdmin      bool   `json:"isAdmin"`
	IsOrg        bool   `json:"isOrg"`
	IsImported   bool   `json:"isImported"`
	OnDemand     bool   `json:"onDemand"`
	OtpEnabled   bool   `json:"otpEnabled"`
	MembersCount int    `json:"membersCount"`
	TeamsCount   int    `json:"teamsCount"`
}

// ResponseAccounts struct.
type ResponseAccounts struct {
	UsersCount    int               `json:"usersCount"`
	OrgsCount     int               `json:"orgsCount"`
	ResourceCount int               `json:"resourceCount"`
	NextPageStart string            `json:"nextPageStart"`
	Accounts      []ResponseAccount `json:"accounts"`
}

// Account filters enum.
type AccountFilter string

const (
	AccountFilterUsers         AccountFilter = "user"
	AccountFilterOrgs          AccountFilter = "orgs"
	AccountFilterAdmins        AccountFilter = "admins"
	AccountFilterNonAdmins     AccountFilter = "non-admins"
	AccountFilterActiveUsers   AccountFilter = "active-users"
	AccountFilterInactiveUsers AccountFilter = "inactive-users"
	URLTargetForAccounts                     = "accounts"
)

// APIFormOfFilter is a string readable form of the AccountFilters enum.
func (accF AccountFilter) APIFormOfFilter() string {
	filters := [...]string{"users", "orgs", "admins", "non-admins", "active-users"}

	x := string(accF)
	for _, v := range filters {
		if v == x {
			return x
		}
	}

	return "all"
}

// CreateAccount method - checking the MKE health endpoint.
func (c *Client) ApiCreateAccount(ctx context.Context, acc CreateAccount) (ResponseAccount, error) {
	if (acc == CreateAccount{}) {
		return ResponseAccount{}, fmt.Errorf("creating account failed. %w: %+v", ErrEmptyStruct, acc)
	}

	req, err := c.RequestFromTargetAndJSONBody(ctx, http.MethodPost, URLTargetForAccounts, acc)
	if err != nil {
		return ResponseAccount{}, fmt.Errorf("creating account %s failed. %w: %s", acc.Name, ErrRequestCreation, err)
	}

	resp, err := c.doAuthorizedRequest(req)
	if err != nil {
		return ResponseAccount{}, fmt.Errorf("creating account %s failed. %w", acc.Name, err)
	}

	resAcc := ResponseAccount{}
	if err := resp.JSONMarshallBody(&resAcc); err != nil {
		return ResponseAccount{}, fmt.Errorf("creating account %s failed. %w: %s", acc.Name, ErrUnmarshaling, err)
	}

	return resAcc, nil
}

// DeleteAccount deletes a user from in Enzi.
func (c *Client) ApiDeleteAccount(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s", URLTargetForAccounts, id)
	req, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodDelete, url, []byte{})

	if err != nil {
		return fmt.Errorf("deleting account %s failed. %w: %s", id, ErrRequestCreation, err)
	}

	if _, err = c.doAuthorizedRequest(req); err != nil {
		return fmt.Errorf("deleting account %s failed. %w", id, err)
	}
	return nil
}

// ReadAccount method retrieves a user from the enzi endpoint.
func (c *Client) ApiReadAccount(ctx context.Context, id string) (ResponseAccount, error) {
	url := fmt.Sprintf("%s/%s", URLTargetForAccounts, id)
	req, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodGet, url, []byte{})
	if err != nil {
		return ResponseAccount{}, fmt.Errorf("reading account %s failed. %w: %s", id, ErrRequestCreation, err)
	}

	resp, err := c.doAuthorizedRequest(req)
	if err != nil {
		return ResponseAccount{}, fmt.Errorf("reading account %s failed. %w", id, err)
	}

	resAcc := ResponseAccount{}
	if err := resp.JSONMarshallBody(&resAcc); err != nil {
		return ResponseAccount{}, fmt.Errorf("reading account %s failed. %w: %s", id, ErrUnmarshaling, err)
	}
	return resAcc, nil
}

// UpdateAccount updates a user in the enzi endpoint.
func (c *Client) ApiUpdateAccount(ctx context.Context, id string, acc UpdateAccount) (ResponseAccount, error) {
	url := fmt.Sprintf("%s/%s", URLTargetForAccounts, id)

	req, err := c.RequestFromTargetAndJSONBody(ctx, http.MethodPatch, url, acc)

	if err != nil {
		return ResponseAccount{}, fmt.Errorf("updating account %s failed. %w: %s", id, ErrRequestCreation, err)
	}

	resp, err := c.doAuthorizedRequest(req)
	if err != nil {
		return ResponseAccount{}, fmt.Errorf("updating account %s failed. %w", id, err)
	}

	resAcc := ResponseAccount{}
	if err := resp.JSONMarshallBody(&resAcc); err != nil {
		return ResponseAccount{}, fmt.Errorf("updating account %s failed. %w: %s", id, ErrUnmarshaling, err)
	}
	return resAcc, nil
}

// ReadAccounts method retrieves all accounts depending on the filter passed from the enzi endpoint.
func (c *Client) ApiReadAccounts(ctx context.Context, accFilter AccountFilter) ([]ResponseAccount, error) {
	// req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.createEnziUrl("accounts"), nil)
	req, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodGet, URLTargetForAccounts, []byte{})
	if err != nil {
		return []ResponseAccount{}, fmt.Errorf("reading accounts in bulk '%s' failed. %w: %s",
			accFilter.APIFormOfFilter(), ErrRequestCreation, err)
	}

	q := req.URL.Query()
	q.Add("filter", accFilter.APIFormOfFilter())
	req.URL.RawQuery = q.Encode()

	resp, err := c.doAuthorizedRequest(req)
	if err != nil {
		return []ResponseAccount{}, fmt.Errorf("reading accounts in bulk '%s' failed. %w",
			accFilter.APIFormOfFilter(), err)
	}

	var accs ResponseAccounts
	if err := resp.JSONMarshallBody(&accs); err != nil {
		return []ResponseAccount{}, fmt.Errorf("reading accounts in bulk '%s' failed. %w: %s",
			accFilter.APIFormOfFilter(), ErrUnmarshaling, err)
	}

	return accs.Accounts, nil
}
