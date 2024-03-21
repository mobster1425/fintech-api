package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
//	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	
	mockdb "feyin/digital-fintech-api/db/mock"
	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"
)

func TestGetWalletAPI(t *testing.T) {
	user, _ := randomUser(t)
	wallet := randomWallet(user.Username)
	//fmt.Print(wallet)
	// fmt.Printf("GetWallet before called with ID: %v\n", wallet.ID)
	// fmt.Printf("GetWallet before called with Owner: %v\n", wallet.Owner)
// 	fmt.Printf("GetWallet before called with balance: %v\n", wallet.Balance)
	testCases := []struct {
		name          string
		ID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			ID: wallet.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				 addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			//	 fmt.Print("after authorization called")
			},
			buildStubs: func(store *mockdb.MockStore) {
			//	fmt.Printf("GetWallet called with ID: %v\n", wallet.ID)
				store.EXPECT().
					GetWallet(gomock.Any(),gomock.Eq(wallet.ID)).
					Times(1).
					Return(wallet, nil)
					//.AnyTimes()

				//	fmt.Print("get wallet called")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			//	fmt.Printf("recorder response: %v\n", recorder.Code)
			//	fmt.Printf("recorder response body: %v\n", recorder.Body)
				 require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWallet(t, recorder.Body, wallet)
			},
		},
		{
			name:      "UnauthorizedUser",
			ID: wallet.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", util.RandomRole(), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWallet(gomock.Any(), gomock.Eq(wallet.ID)).
					Times(1).
					Return(wallet, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			ID: wallet.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWallet(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			ID: wallet.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWallet(gomock.Any(), gomock.Eq(wallet.ID)).
					Times(1).
					Return(db.Wallet{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			ID: wallet.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWallet(gomock.Any(), gomock.Eq(wallet.ID)).
					Times(1).
					Return(db.Wallet{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			ID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWallet(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}



	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/wallet/%d", tc.ID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
		//	fmt.Print("auth setup finish")
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}


type eqAddWalletBalanceParamsMatcher struct {
	arg      db.AddWalletBalanceParams
//	password string
}



func (e eqAddWalletBalanceParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.( db.AddWalletBalanceParams)
	if !ok {
		return false
	}

	

	
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqAddWalletBalanceParamsMatcher) String() string {
	return fmt.Sprintf("matches Add balance params %v", e.arg)
}

func EqDepositMoneyParams(arg db.AddWalletBalanceParams) gomock.Matcher {
	return eqAddWalletBalanceParamsMatcher{arg}
}

func TestDepositIntoWalletAPI(t *testing.T) {
	user, _ := randomUser(t)
	wallet := randomWallet(user.Username)
amount := util.RandomMoney()
// fmt.Printf("user role %v",user.Role)

	testCases := []struct {
		name          string
	//	request        addWalletBalanceRequest
	body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"Amount": amount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
				// fmt.Print("this is called")
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AddWalletBalanceParams{
					Amount: amount,
					ID: wallet.ID,
				}
			//	fmt.Printf("arg for ok %v",arg)
				store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					//Times(1).
					Return(wallet, nil).AnyTimes()
				//	fmt.Print("this is called 2")

				store.EXPECT().
			     	AddWalletBalance(gomock.Any(), EqDepositMoneyParams(arg)).
			//	AddWalletBalance(gomock.Any(), gomock.Eq(&arg)).
					Times(1).
				Return(db.Wallet{ID: wallet.ID, Balance: wallet.Balance + amount}, nil)
				//.AnyTimes()
			//	Return(wallet, nil)
		//	fmt.Print("this is called 3")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			//	fmt.Printf("recorder body= %v",recorder.Body)
			//	fmt.Printf("recorder body= %v",recorder.Code)
				require.Equal(t, http.StatusOK, recorder.Code)
			//	requireBodyMatchWallet(t, recorder.Body, wallet)
			requireBodyMatchWallet(t, recorder.Body, db.Wallet{ID: wallet.ID, Balance: wallet.Balance + amount})
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request",
			body: gin.H{
				"Amount": -amount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No store expectations as the request is invalid
				store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Any()).
				Times(0)
			//	Return(db.Wallet{}, nil)

			store.EXPECT().
				AddWalletBalance(gomock.Any(),gomock.Any()).
				Times(0)
			//	Return(db.Wallet{ID: wallet.ID, Balance: wallet.Balance + 100}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Wallet Not Found",
			body: gin.H{
				"Amount": amount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AddWalletBalanceParams{
					Amount: amount,
					ID: wallet.ID,
				}
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.Wallet{}, db.ErrRecordNotFound)

					store.EXPECT().
				AddWalletBalance(gomock.Any(),arg).
				Times(0)
			//	Return(db.Wallet{ID: wallet.ID, Balance: wallet.Balance + 100}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"Amount": amount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No authorization setup
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
				AddWalletBalance(gomock.Any(),gomock.Any()).
				Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "UnauthorizedUser",
			body: gin.H{
				"Amount": amount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", util.RandomRole(), time.Minute)
				// fmt.Print("this authrization s called")
			},
			buildStubs: func(store *mockdb.MockStore) {
				
     
               store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Eq("unauthorized_user")).
				//Times(0).
				// Return(db.Wallet{}, nil).
				Return(wallet, nil).
				AnyTimes()
// fmt.Print("this is called 1")

//a:=EqDepositMoneyParams(arg)
//fmt.Printf("a = %v",a)
arg := db.AddWalletBalanceParams{
	//	Amount: amount,
		ID: wallet.ID,
		Amount: amount,
	}
//	fmt.Printf("arg = %v",arg)
				store.EXPECT().
					AddWalletBalance(gomock.Any(), EqDepositMoneyParams(arg)).
				//	Times(1).
					Return(db.Wallet{ID: wallet.ID, Balance: wallet.Balance + amount}, nil).AnyTimes()
				
					//Times(0)
				//	Return(db.Wallet{}, nil).AnyTimes()
			//			fmt.Print("this is called 2")

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			//	fmt.Printf("recorder code and body = %v   code %v",recorder.Body,recorder.Code)
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"Amount": amount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					// Times(1).
					Return(db.Wallet{}, sql.ErrConnDone).AnyTimes()



					arg := db.AddWalletBalanceParams{
						//	Amount: amount,
							ID: wallet.ID,
							Amount: amount,
						}

					store.EXPECT().
					AddWalletBalance(gomock.Any(), EqDepositMoneyParams(arg)).
				//	Times(1).
					Return(db.Wallet{},sql.ErrConnDone).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := "/wallet/add-money"
			requestBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}



}






func randomWallet(owner string) db.Wallet {
	return db.Wallet{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
	//	Currency: util.RandomCurrency(),
	}
}




func requireBodyMatchWallet(t *testing.T, body *bytes.Buffer, wallet db.Wallet) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotWallet db.Wallet
	err = json.Unmarshal(data, &gotWallet)
	require.NoError(t, err)
	require.Equal(t, wallet, gotWallet)
}

func requireBodyMatchWallets(t *testing.T, body *bytes.Buffer, wallets []db.Wallet) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotWallets []db.Wallet
	err = json.Unmarshal(data, &gotWallets)
	require.NoError(t, err)
	require.Equal(t, wallets, gotWallets)
}