package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

type testAccountStruct struct {
	server           *httptest.Server
	expectedResponse client.ResponseAccount
	expectedErr      error
}

func TestCreateValidAccount(t *testing.T) {
	testResAcc := client.ResponseAccount{
		ID:       "fake-test-id",
		Name:     "testuser",
		FullName: "Test Name",
	}
	mAccount, err := json.Marshal(testResAcc)
	if err != nil {
		t.Error(err)
	}
	tc := testAccountStruct{
		server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(mAccount); err != nil {
				t.Error(err)
				return
			}
		})),
		expectedResponse: testResAcc,
		expectedErr:      nil,
	}
	defer tc.server.Close()
	testClient, err := client.NewUnsafeSSLClient(tc.server.URL, "fakeuser", "fakepass")

	if err != nil {
		t.Error("couldn't create test client")
	}
	ctx := context.Background()
	resp, err := testClient.ApiCreateAccount(ctx, client.CreateAccount{Name: testResAcc.Name})
	if !reflect.DeepEqual(tc.expectedResponse, resp) {
		t.Errorf("expected (%v), got (%v)", tc.expectedResponse, resp)
	}
	if !errors.Is(err, tc.expectedErr) {
		t.Errorf("expected (%v), got (%v)", tc.expectedErr, err)
	}
}

func TestCreateInvalidAccount(t *testing.T) {
	testResAcc := client.ResponseAccount{}
	mAccount, err := json.Marshal(testResAcc)
	tUser := client.CreateAccount{Name: "testuser"}
	if err != nil {
		t.Fatal(err)
	}
	tc := testAccountStruct{
		server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write(mAccount); err != nil {
				t.Error(err)
				return
			}
		})),
		expectedResponse: testResAcc,
		expectedErr:      client.ErrResponseError,
	}
	defer tc.server.Close()
	testClient, err := client.NewUnsafeSSLClient(tc.server.URL, "fakeuser", "fakepass")
	if err != nil {
		t.Error("couldn't create test client")
	}
	ctx := context.Background()

	resp, err := testClient.ApiCreateAccount(ctx, tUser)

	if !reflect.DeepEqual(tc.expectedResponse, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", tc.expectedResponse, resp)
	}
	if !errors.Is(err, tc.expectedErr) {
		t.Errorf("expected error: (%v),\n got (%v)", tc.expectedErr, err)
	}
}

func TestCreateAccountEmpty(t *testing.T) {
	tc := testAccountStruct{
		server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write(nil); err != nil {
				t.Error(err)
				return
			}
		})),
		expectedResponse: client.ResponseAccount{},
		expectedErr:      client.ErrEmptyStruct,
	}
	defer tc.server.Close()
	testClient, err := client.NewUnsafeSSLClient(tc.server.URL, "fakeuser", "fakepass")
	if err != nil {
		t.Error("couldn't create test client")
	}
	ctx := context.Background()

	resp, err := testClient.ApiCreateAccount(ctx, client.CreateAccount{})

	if !reflect.DeepEqual(tc.expectedResponse, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", tc.expectedResponse, resp)
	}
	if !errors.Is(err, tc.expectedErr) {
		t.Errorf("expected error: (%v),\n got (%v)", tc.expectedErr, err)
	}
}

func TestCreateAccountUnmarshalErr(t *testing.T) {
	tc := testAccountStruct{
		server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write(nil); err != nil {
				t.Error(err)
				return
			}
		})),
		expectedResponse: client.ResponseAccount{},
		expectedErr:      client.ErrResponseError,
	}
	defer tc.server.Close()
	testClient, err := client.NewUnsafeSSLClient(tc.server.URL, "fakeuser", "fakepass")
	if err != nil {
		t.Error("couldn't create test client")
	}
	ctx := context.Background()

	resp, err := testClient.ApiCreateAccount(ctx, client.CreateAccount{Name: "fake"})

	if !reflect.DeepEqual(tc.expectedResponse, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", tc.expectedResponse, resp)
	}
	if !errors.Is(err, tc.expectedErr) {
		t.Errorf("expected error: (%v),\n got (%v)", tc.expectedErr, err)
	}
}

func TestDeleteAccountSuccess(t *testing.T) {
	ctx := context.Background()
	auth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	testAcc := "fakeid"
	mockRequest := MockHandlerKey{
		Path:   fmt.Sprintf("%s/%s", client.URLTargetForAccounts, testAcc),
		Method: http.MethodDelete,
	}
	keysResp := errors.New("")

	svr := MockTestServer(&auth, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnJson(keysResp),
	})

	u, _ := url.Parse(svr.URL)
	c, err := client.NewClient(u, &auth, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	err = c.ApiDeleteAccount(ctx, testAcc)
	if err != nil {
		t.Fatalf("delete account api call failed: %s", err)
	}
}

func TestReadAccountSuccess(t *testing.T) {
	resAcc := client.ResponseAccount{
		Name: "fakeacc",
	}
	mResAcc, err := json.Marshal(resAcc)
	if err != nil {
		t.Fatal(err)
	}
	tc := testAccountStruct{
		server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(mResAcc); err != nil {
				t.Error(err)
				return
			}
		})),
		expectedResponse: resAcc,
		expectedErr:      nil,
	}
	defer tc.server.Close()
	testClient, err := client.NewUnsafeSSLClient(tc.server.URL, "fakeuser", "fakepass")
	if err != nil {
		t.Error("couldn't create test client")
	}
	ctx := context.Background()
	resp, err := testClient.ApiReadAccount(ctx, "fakeid")

	if !reflect.DeepEqual(tc.expectedResponse, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", tc.expectedResponse, resp)
	}
	if !errors.Is(err, tc.expectedErr) {
		t.Errorf("expected error: (%v),\n got (%v)", tc.expectedErr, err)
	}
}

func TestUpdateAccountSuccess(t *testing.T) {
	uAcc := client.ResponseAccount{
		Name: "fakeacc",
	}
	mUAcc, err := json.Marshal(uAcc)
	if err != nil {
		t.Fatal(err)
	}
	tc := testAccountStruct{
		server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(mUAcc); err != nil {
				t.Error(err)
				return
			}
		})),
		expectedResponse: uAcc,
		expectedErr:      nil,
	}
	defer tc.server.Close()
	testClient, err := client.NewUnsafeSSLClient(tc.server.URL, "fakeuser", "fakepass")
	if err != nil {
		t.Error("couldn't create test client")
	}
	ctx := context.Background()
	resp, err := testClient.ApiUpdateAccount(ctx, "fakeid", client.UpdateAccount{FullName: "mock"})

	if !reflect.DeepEqual(tc.expectedResponse, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", tc.expectedResponse, resp)
	}
	if !errors.Is(err, tc.expectedErr) {
		t.Errorf("expected error: (%v),\n got (%v)", tc.expectedErr, err)
	}
}

func TestReadAccountsSuccess(t *testing.T) {
	resAccs := []client.ResponseAccount{}
	resAccs = append(resAccs,
		client.ResponseAccount{Name: "mock1"}, client.ResponseAccount{Name: "mock2"})

	accs := struct {
		UsersCount    int    `json:"usersCount"`
		OrgsCount     int    `json:"orgsCount"`
		ResourceCount int    `json:"resourceCount"`
		NextPageStart string `json:"nextPageStart"`

		Accounts []client.ResponseAccount `json:"accounts"`
	}{Accounts: resAccs}
	mResAccs, err := json.Marshal(accs)
	if err != nil {
		t.Fatal(err)
	}
	tc := testAccountStruct{
		server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(mResAccs); err != nil {
				t.Error(err)
				return
			}
		})),
		expectedErr: client.ErrUnmarshaling,
	}
	defer tc.server.Close()
	testClient, err := client.NewUnsafeSSLClient(tc.server.URL, "fakeuser", "fakepass")
	if err != nil {
		t.Error("couldn't create test client")
	}
	ctx := context.Background()
	resp, err := testClient.ApiReadAccounts(ctx, client.Users)

	if !reflect.DeepEqual(resAccs, resp) {
		t.Errorf("expected resp: (%+v),\n got (%+v)", tc.expectedResponse, resp)
	}
	if errors.Is(err, tc.expectedErr) {
		t.Errorf("expected error: (%v),\n got (%v)", tc.expectedErr, err)
	}
}
