package api

import (
	"bytes"
	"encoding/json"
	//"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	//"golang.org/x/tools/go/analysis/passes/printf"

	//mockdb "github.com/techschool/simplebank/db/mock"
	mockdb "feyin/digital-fintech-api/db/mock"
	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"
)


func TestPayToMerchantAPI(t *testing.T) {



	user1, _ := randomUser(t)
	user1.Role = db.NullUserRole{
		UserRole:	db.UserRoleCustomer,
		}

	user2,_:=randomUser(t)
	user2.Role = db.NullUserRole{
	UserRole:	db.UserRoleCustomer,
	}
	//user3, _ := randomUser(t)
//	user3.Role.UserRole=db.UserRoleCustomer
	wallet1 := randomWallet(user1.Username)
	wallet1.Balance=100000
	//user2, _ := randomUser(t)
	

	 merchantUser, _ := randomUserWithRole(t)
//	 wallet2 := randomWallet(merchantUser.Username)
	merchantWallet := randomWallet(merchantUser.Username)

	voucher := randomVoucher(merchantUser.Username)
ChargeForSender:=false
//	amount := util.RandomMoney()
amount := int64(100)
	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), ChargeForSender)
	note := "Pay to merchant"
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
				"ReceiverUsername": merchantUser.Username,
			// "ReceiverUsername": pgtype.Text{String: merchantUser.Username, Valid: true},
				"amount":     amount,
				"note":       note,
				"VoucherCode": voucher.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
		//		fmt.Printf("merchant user = %v",merchantUser.Username)
			//	us:=pgtype.Text{String: merchantUser.Username, Valid: true}
				store.EXPECT().
				//	GetUser(gomock.Any(), gomock.Eq(string(merchantUser.Username))).
			GetUser(gomock.Any(), gomock.Eq(merchantUser.Username)).
				//	Times(1).
					Return(merchantUser, nil).AnyTimes()

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
				//	Times(1).
					Return(wallet1, nil).AnyTimes()

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(merchantUser.Username)).
				//	Times(1).
					Return(merchantWallet, nil).AnyTimes()

				store.EXPECT().
					GetVoucherWithCode(gomock.Any(), gomock.Eq(voucher.Code)).
			//		Times(1).
					Return(voucher,nil).AnyTimes()


				



           
            arg:=db.TransactionParams{
				SenderWalletID:   wallet1.ID,
				ReceiverWalletID: pgtype.Int8{Int64: merchantWallet.ID, Valid: true},
				Amount:           pgtype.Int8{Int64:amount, Valid: true},
				Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
				Type: db.NullTransactionType{
					TransactionType: db.TransactionTypePAYMENTVOUCHER,
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
				VoucherID:      voucher.ID,
				UsedByUsername: []string{user1.Username},
			}

			store.EXPECT().
				//	TransferTx(gomock.Any(), gomock.Eq(arg)).AnyTimes()
				TransferTxForVoucher(gomock.Any(), gomock.Eq(arg)).AnyTimes()
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
					GetUser(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
					

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
					

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
					

				store.EXPECT().
					GetVoucherWithCode(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
					

					store.EXPECT().
					TransferTxForVoucher(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)



			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			//	fmt.Printf("recorder body = %v",recorder.Body)
			//	fmt.Printf("recorder code = %v",recorder.Code)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		*/

		{
			name: "Invalid Request - Receiver Not a Merchant",
			body: gin.H{
			//	"ReceiverUsername": merchantUser.Username,
			"ReceiverUsername": user2.Username,
				"amount":     amount,
				"note":       note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
			//	fmt.Printf("merchant username = %v",merchantUser.Username)
			//	fmt.Printf("user2 username = %v",user2.Username)
				store.EXPECT().
			//		GetUser(gomock.Any(), gomock.Eq("unohk")).
			GetUser(gomock.Any(), gomock.Eq(user2.Username)).
				//	Times(1).
					Return(user2, nil).AnyTimes()

/*
					store.EXPECT().
				//	GetWalletbyOwner(gomock.Any(), gomock.Eq("unohk")).
				GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
				//	Times(1).
					Return(wallet1, nil).AnyTimes()

					
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(merchantUser.Username)).
				//	Times(1).
					Return(merchantWallet, nil).AnyTimes()

				store.EXPECT().
					GetVoucherWithCode(gomock.Any(), gomock.Eq(voucher.Code)).
			//		Times(1).
					Return(voucher,nil).AnyTimes()


				



           
            arg:=db.TransactionParams{
				SenderWalletID:   wallet1.ID,
				ReceiverWalletID: pgtype.Int8{Int64: merchantWallet.ID, Valid: true},
				Amount:           pgtype.Int8{Int64:amount, Valid: true},
				Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
				Type: db.NullTransactionType{
					TransactionType: db.TransactionTypePAYMENTVOUCHER,
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
				VoucherID:      voucher.ID,
				UsedByUsername: []string{user1.Username},
			}

			store.EXPECT().
				//	TransferTx(gomock.Any(), gomock.Eq(arg)).AnyTimes()
				TransferTxForVoucher(gomock.Any(), gomock.Eq(arg)).AnyTimes()
				//	Times(1)

				store.EXPECT().
				UpdateTransactionStatus(gomock.Any(), gomock.Eq(db.UpdateTransactionStatusParams{
					ID:     0,
					Status: db.NullTransactionStatus{TransactionStatus: db.TransactionStatusSUCCESS, Valid: true},
				})).AnyTimes()
				*/
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "Invalid Request - Receiver Wallet Not Found",
			body: gin.H{
				"ReceiverUsername": merchantUser.Username,
				"amount":     amount,
				"note":       "Payment to merchant",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(merchantUser.Username)).
				//	Times(1).
					Return(merchantUser, nil).AnyTimes()

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
				//	Times(1).
					Return(wallet1, nil).AnyTimes()

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(merchantUser.Username)).
				//	Times(1).
					Return(db.Wallet{}, db.ErrRecordNotFound).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		//		fmt.Printf("recorder body = %v",recorder.Body)
		//		fmt.Printf("recorder code = %v",recorder.Code)
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		{
			name: "Invalid Request - Sender Wallet Not Found",
			body: gin.H{
				"ReceiverUsername": merchantUser.Username,
				"amount":     amount,
				"note":       "Payment to merchant",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(merchantUser.Username)).
				//	Times(1).
					Return(merchantUser, nil).AnyTimes()

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
					// Times(1).
					Return(db.Wallet{}, db.ErrRecordNotFound).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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

			url := "/pay-to-merchant/"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t,recorder)
		})
	}




}