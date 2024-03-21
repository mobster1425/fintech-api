package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	// "fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	//mockdb "github.com/techschool/simplebank/db/mock"
	mockdb "feyin/digital-fintech-api/db/mock"
	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"
)


func TestPeerToPeerAPI(t *testing.T) {


	

	user2, _ := randomUser(t)
	user3, _ := randomUser(t)
	user1, _ := randomUser(t)
	wallet1 := randomWallet(user1.Username)
	 wallet1.Balance=10000
	 wallet2 := randomWallet(user2.Username)
	 wallet3 := randomWallet(user3.Username)
	

//	voucher := randomVoucher(user2.Username)
ChargeForSender:=true
	// amount := util.RandomMoney()
	amount := int64(100)
	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), ChargeForSender)
	note := "Peer to peer payment"
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{

		{
			name: "OK",
			body: gin.H{
				"receiver_username": user2.Username,
				"amount":     amount,
				"note":      note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
					//Times(1).
					Return(wallet1, nil).AnyTimes()

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user2.Username)).
				//	Times(1).
					Return(wallet2, nil).AnyTimes()







					arg:=db.TransactionParams{
						SenderWalletID:   wallet1.ID,
						ReceiverWalletID: pgtype.Int8{Int64: wallet2.ID, Valid: true},
						Amount:           pgtype.Int8{Int64:amount, Valid: true},
						Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
						Type: db.NullTransactionType{
							TransactionType: db.TransactionTypeTRANSFER,
							Valid:           true,
						},
						Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
						Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
						Note:          pgtype.Text{
							String: note,
							Valid: true,
						},
						Status: db.NullTransactionStatus{
							TransactionStatus: db.TransactionStatusPROCESSING,
							Valid:             true,
						},
					//	VoucherID:      voucher.ID,
					//	UsedByUsername: []string{user1.Username},
					}
		
					store.EXPECT().
							TransferTx(gomock.Any(), gomock.Eq(arg)).AnyTimes()
						//	Times(1)
				
						store.EXPECT().
						UpdateTransactionStatus(gomock.Any(), gomock.Eq(db.UpdateTransactionStatusParams{
							ID:     0,
							Status: db.NullTransactionStatus{TransactionStatus: db.TransactionStatusSUCCESS, Valid: true},
						})).AnyTimes()

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			//	fmt.Printf("recorder body = %v",recorder.Body)
			//	fmt.Printf("recorder code = %v",recorder.Code)
				require.Equal(t, http.StatusOK, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		/*
		{
			name: "Invalid Request - Missing Required Field",
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No store expectations as the request is invalid
				store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
			//	Times(0)
				

			store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
			//	Times(0)
				


				store.EXPECT().
							TransferTx(gomock.Any(), gomock.Any()).AnyTimes()
						
						//	Times(0)

						store.EXPECT().
						UpdateTransactionStatus(gomock.Any(), gomock.Eq(db.UpdateTransactionStatusParams{
							ID:     0,
							Status: db.NullTransactionStatus{TransactionStatus: db.TransactionStatusSUCCESS, Valid: true},
						})).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
*/
		{
			name: "Invalid Request - Sender Wallet Not Found",
			body: gin.H{
				"receiver_username": user2.Username,
				"amount":     amount,
				"note":      note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
				//	Times(1).
					Return(db.Wallet{}, db.ErrRecordNotFound).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Invalid Request - Receiver Wallet Not Found",
			body: gin.H{
				"receiver_username": user2.Username,
				"amount":     amount,
				"note":       note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
				//	Times(1).
					Return(wallet1, nil).AnyTimes()
		
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user2.Username)).
				//	Times(1).
					Return(db.Wallet{}, db.ErrRecordNotFound).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Invalid Request - Sender Wallet Owner Mismatch",
			body: gin.H{
				"receiver_username": user2.Username,
				"amount":     amount,
				"note":       note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
			//	fmt.Printf("user 3 username = %v",user3.Username)
			//	fmt.Printf("user 1 username = %v",user1.Username)
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
				//	Times(1).
					Return(wallet3, nil).AnyTimes()
		
					/*
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user2.Username)).
				//	Times(1).
					Return(wallet2, nil).AnyTimes()*/
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			//	fmt.Printf("recorder code for mismatch = %v",recorder.Code)
			//	fmt.Printf("recorder body for mismatch = %v",recorder.Body)
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Invalid Request - Negative Amount",
			body: gin.H{
				"receiver_username": user2.Username,
				"amount":     -amount,
				"note":       note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No store expectations as the request is invalid
				store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
			//	Times(0)
				

			store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
			//	Times(0)
				


				store.EXPECT().
							TransferTx(gomock.Any(), gomock.Any()).AnyTimes()
						//	Times(0)

						store.EXPECT().
						UpdateTransactionStatus(gomock.Any(), gomock.Any()).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "InternalError",
			body: gin.H{
				"receiver_username": user2.Username,
				"amount":     amount,
				"note":       note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Mocking an internal error during voucher creation
				
				
				store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
			//	Times(1).
				Return(db.Wallet{},sql.ErrConnDone).AnyTimes()

			store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Eq(user2.Username)).
			//	Times(1).
				Return(db.Wallet{}, sql.ErrConnDone).AnyTimes()


				
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"receiver_username": user2.Username,
				"amount":     amount,
				"note":       note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Omit authorization setup to simulate no authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//Times(0)
				

			store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
			//	Times(0)
				


				store.EXPECT().
							TransferTx(gomock.Any(), gomock.Any()).AnyTimes()
						//	Times(0)


						store.EXPECT().
						UpdateTransactionStatus(gomock.Any(),gomock.Any()).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "UnauthorizedUser",
			body: gin.H{
				"receiver_username": user2.Username,
				"amount":     amount,
				"note":       note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Set up an unauthorized user
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", util.RandomRole(), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {


				store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Eq("unauthorized_user")).
			//	Times(1).
				Return(db.Wallet{},nil).AnyTimes()

			store.EXPECT().
				GetWalletbyOwner(gomock.Any(), gomock.Eq(user2.Username)).
			//	Times(1).
				Return(db.Wallet{}, nil).AnyTimes()

				store.EXPECT().
				TransferTx(gomock.Any(), gomock.Any()).AnyTimes()
			//	Times(0)


			store.EXPECT().
			UpdateTransactionStatus(gomock.Any(),gomock.Any()).AnyTimes()
			
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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

		// Marshal body data to JSON
		data, err := json.Marshal(tc.body)
		require.NoError(t, err)

		url := "/p2p-payment/"
		request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		require.NoError(t, err)

		tc.setupAuth(t, request, server.tokenMaker)
		server.router.ServeHTTP(recorder, request)
		tc.checkResponse(t,recorder)
	})
}



}