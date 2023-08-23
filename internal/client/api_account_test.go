package client_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestCreateValidAccount(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	accCreate := client.CreateAccount{
		Name:     "testuser",
		FullName: "testname",
	}
	accResp := client.ResponseAccount{
		ID:       "fake-test-id",
		Name:     accCreate.Name,
		FullName: accCreate.FullName,
	}

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodPost, client.URLTargetForAccounts, MockServerHandlerGeneratorReturnJson(accResp))
	defer s.Close()

	c, _ := s.Client()

	resp, err := c.ApiCreateAccount(ctx, accCreate)
	if err != nil {
		t.Fatalf("Account create failed: %s", err)
	}

	if !reflect.DeepEqual(accResp, resp) {
		t.Errorf("expected (%v), got (%v)", accResp, resp)
	}
}

func TestCreateAccountServerRejectsRequest(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	accCreate := client.CreateAccount{
		Name: "testuser",
	}
	expectedError := client.ErrResponseError

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodPost, client.URLTargetForAccounts, MockServerHandlerGeneratorReturnResponseStatus(http.StatusBadRequest))
	defer s.Close()

	c, _ := s.Client()

	_, err := c.ApiCreateAccount(ctx, accCreate)

	if !errors.Is(err, expectedError) {
		t.Errorf("expected error: (%v),\n got (%v)", expectedError, err)
	}
}

func TestCreateAccountResponseUnmarshalErr(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	accCreate := client.CreateAccount{
		Name: "testuser",
	}
	expectedError := client.ErrUnmarshaling

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodPost, client.URLTargetForAccounts, MockServerHandlerGeneratorReturnBytes([]byte("I am not json")))
	defer s.Close()

	c, _ := s.Client()

	_, err := c.ApiCreateAccount(ctx, accCreate)

	if !errors.Is(err, expectedError) {
		t.Errorf("expected error: (%v),\n got (%v)", expectedError, err)
	}
}

func TestDeleteAccountSuccess(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	acc := "testuser"

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodDelete, fmt.Sprintf("%s/%s", client.URLTargetForAccounts, acc), MockServerHandlerGeneratorReturnResponseStatus(http.StatusOK))
	defer s.Close()

	c, err := s.Client()
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err = c.ApiDeleteAccount(ctx, acc); err != nil {
		t.Fatalf("delete account api call failed: %s", err)
	}
}

func TestReadAccountSuccess(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	acc := "testuser"
	expectedAcc := client.ResponseAccount{
		Name: "fakeacc",
	}

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodGet, fmt.Sprintf("%s/%s", client.URLTargetForAccounts, acc), MockServerHandlerGeneratorReturnJson(expectedAcc))

	c, _ := s.Client()

	if resp, err := c.ApiReadAccount(ctx, acc); err != nil {
		t.Errorf("unexpected error reading account: %s", err.Error())
	} else if !reflect.DeepEqual(expectedAcc, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", expectedAcc, resp)
	}
}

func TestUpdateAccountSuccess(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	acc := "testuser"
	expectedAcc := client.ResponseAccount{
		Name: "fakeacc",
	}

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodPatch, fmt.Sprintf("%s/%s", client.URLTargetForAccounts, acc), MockServerHandlerGeneratorReturnJson(expectedAcc))
	defer s.Close()

	c, _ := s.Client()

	if resp, err := c.ApiUpdateAccount(ctx, acc, client.UpdateAccount{FullName: "mock"}); err != nil {
		t.Errorf("unexpected error updating account: %s", err.Error())
	} else if !reflect.DeepEqual(expectedAcc, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", expectedAcc, resp)
	}
}

func TestReadAccountsSuccess(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	readFilter := client.AccountFilterUsers
	expectedAccs := []client.ResponseAccount{
		{Name: "mock1"},
		{Name: "mock2"},
	}
	responseAccs := client.ResponseAccounts{
		UsersCount:    2,
		OrgsCount:     0,
		ResourceCount: 0,
		NextPageStart: "",

		Accounts: expectedAccs,
	}

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodGet, client.URLTargetForAccounts, MockServerHandlerGeneratorReturnJson(responseAccs))
	defer s.Close()

	c, _ := s.Client()

	if resp, err := c.ApiReadAccounts(ctx, readFilter); err != nil {
		t.Errorf("unexpected error reading accounts: %s", err.Error())
	} else if !reflect.DeepEqual(expectedAccs, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", expectedAccs, resp)
	}
}
