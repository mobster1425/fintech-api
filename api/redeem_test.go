package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func TestCreateRedeemAPI(t *testing.T) {
	user1, _ := randomUser(t)
	wallet1 := randomWallet(user1.Username)
	wallet1.Balance=100000
	user1.Role = db.NullUserRole{
		UserRole:	db.UserRoleCustomer,
		}
	user2, _ := randomUser(t)
	user2.Role = db.NullUserRole{
		UserRole:	db.UserRoleCustomer,
		}
//	wallet2 := randomWallet(user2.Username)
//	amount := util.RandomMoney()
amount := int64(100)
	ChargeForSender := true
	note := "Redeem transaction"
	//fmt.Println(">> Amount:", amount)
	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), ChargeForSender)
	//	fmt.Println(">> SendAmount:", sendAmount)
	//	fmt.Println(">> ReceiveAmount:", receiveAmount)

	//	user2, _ := randomUser(t)
	//	wallet2 := randomWallet(user2.Username)

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
				"charge": ChargeForSender,
				"amount": amount,
				"note":   note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
					//Times(1).
					Return(wallet1, nil).AnyTimes()

				arg := db.TransactionParams{
					SenderWalletID: wallet1.ID,
					//	ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
					Amount: pgtype.Int8{Int64: amount, Valid: true},
					Charge: pgtype.Int8{Int64: int64(chargeFee), Valid: true},
					Type: db.NullTransactionType{
						TransactionType: db.TransactionTypeREDEEM,
						Valid:           true,
					},
					Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
					Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
					Note: pgtype.Text{
						String: note,
						Valid:  true,
					},
					Status: db.NullTransactionStatus{
						TransactionStatus: db.TransactionStatusPROCESSING,
						Valid:             true,
					},
				}

		//		fmt.Printf("arg is = %v", arg)

				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Eq(arg), gomock.Eq(true)).AnyTimes()
				//	Times(1)
				/*
						Return(db.RedeemTxResult{
							Transaction:  db.Transaction{ID: 1},
							SenderWallet: db.Wallet{ID: wallet1.ID, Balance: wallet1.Balance - 100},
							Redeem:       db.Redeem{ID: 1},
						}, nil)



					
				*/
				store.EXPECT().
				UpdateTransactionStatus(gomock.Any(), gomock.Eq(db.UpdateTransactionStatusParams{
					ID:     0,
					Status: db.NullTransactionStatus{TransactionStatus: db.TransactionStatusPENDING, Valid: true},
				})).AnyTimes()
			//	Times(1)
				//implement updatetransactionstatus
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			//	fmt.Printf("recorder body = %v", recorder.Body)
			//	fmt.Printf("recorder code = %v", recorder.Code)
				require.Equal(t, http.StatusOK, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Missing Required Field",
			body: gin.H{
				// Missing 'charge' field
				"amount": amount,
				"note":   note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No store expectations as the request is invalid
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//Times(0)
				//	Return(wallet1, nil)

				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Non-positive Amount",
			body: gin.H{
				"charge": ChargeForSender,
				// Non-positive amount
				"amount": -amount,
				"note":   note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(wallet1, nil)

				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Non-boolean ChargeForSender",
			body: gin.H{
				// Non-boolean charge
				"charge": "invalid",
				"amount": amount,
				"note":   note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(wallet1, nil)

				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				// No store expectations as the request is invalid
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Non-integer Amount",
			body: gin.H{
				"charge": ChargeForSender,
				// Non-integer amount
				"amount": 10.5,
				"note":   note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No store expectations as the request is invalid
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(wallet1, nil)

				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Sender Wallet Owner Mismatch",
			body: gin.H{
				"charge": ChargeForSender,
				"amount": amount,
				"note":   note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Authenticated user does not match sender wallet owner
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user2.Username, string(user2.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user2.Username)).
					//	Times(1).
					Return(wallet1, nil).AnyTimes()
/*
				arg := db.TransactionParams{
					SenderWalletID: wallet2.ID,
					//	ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
					Amount: pgtype.Int8{Int64: amount, Valid: true},
					Charge: pgtype.Int8{Int64: int64(chargeFee), Valid: true},
					Type: db.NullTransactionType{
						TransactionType: db.TransactionTypeREDEEM,
						Valid:           true,
					},
					Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
					Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
					Note: pgtype.Text{
						String: note,
						Valid:  true,
					},
					Status: db.NullTransactionStatus{
						TransactionStatus: db.TransactionStatusPROCESSING,
						Valid:             true,
					},
				}

				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Eq(arg), gomock.Eq(true)).AnyTimes()
				//	Times(1)


				store.EXPECT().
				UpdateTransactionStatus(gomock.Any(), gomock.Eq(db.UpdateTransactionStatusParams{
					ID:     0,
					Status: db.NullTransactionStatus{TransactionStatus: db.TransactionStatusPENDING, Valid: true},
				})).AnyTimes()*/
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		/*
		{
			name: "Valid Request - Status and Type Default to Appropriate Values",
			body: gin.H{
				"charge": ChargeForSender,
				"amount": amount,
				"note":   note,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
					//	Times(1).
					Return(wallet1, nil).AnyTimes()

				arg := db.TransactionParams{
					SenderWalletID: wallet1.ID,
					//	ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
					Amount: pgtype.Int8{Int64: amount, Valid: true},
					Charge: pgtype.Int8{Int64: int64(chargeFee), Valid: true},
					/*
						Type: db.NullTransactionType{
							TransactionType: db.TransactionTypeREDEEM,
							Valid:           true,
						}, 
					Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
					Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
					Note: pgtype.Text{
						String: note,
						Valid:  true,
					},
					
						Status: db.NullTransactionStatus{
							TransactionStatus: db.TransactionStatusPROCESSING,
							Valid:             true,
						}, 
				}

				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Eq(arg), gomock.Eq(true)).AnyTimes()
				//	Times(1)

				store.EXPECT().
				UpdateTransactionStatus(gomock.Any(), gomock.Eq(db.UpdateTransactionStatusParams{
					ID:     0,
					Status: db.NullTransactionStatus{TransactionStatus: db.TransactionStatusPENDING, Valid: true},
				})).AnyTimes()

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				// Additional checks for the response body if needed
			},
		},*/
		/*
			{
				name: "Valid Request - ChargeForSender False (No Charge)",
				body: gin.H{
					"charge": false,
					"amount": amount,
					"note":   note,
				},
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, string(user1.Role.UserRole), time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetWalletbyOwner(gomock.Any(), gomock.Eq(user1.Username)).
						Times(1).
						Return(wallet1, nil)

					arg := db.TransactionParams{
						SenderWalletID: wallet1.ID,
						//	ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
						Amount: pgtype.Int8{Int64: amount, Valid: true},
						Charge: pgtype.Int8{Int64: int64(chargeFee), Valid: true},

						Type: db.NullTransactionType{
							TransactionType: db.TransactionTypeREDEEM,
							Valid:           true,
						},
						Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
						Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
						Note: pgtype.Text{
							String: note,
							Valid:  true,
						},

						Status: db.NullTransactionStatus{
							TransactionStatus: db.TransactionStatusPROCESSING,
							Valid:             true,
						},
					}

					store.EXPECT().
						RedeemTx(gomock.Any(), gomock.Eq(arg), gomock.Eq(true)).
						Times(1)

				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					// Additional checks for the response body if needed
				},
			},
		*/

		// Add other test cases for different scenarios
		// ...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				store := mockdb.NewMockStore(ctrl)
				tc.buildStubs(store)

				server := newTestServer(t, store, nil)
				recorder := httptest.NewRecorder()

				url := "/redeem/" // Adjust the endpoint URL if needed
				requestBody, err := json.Marshal(tc.body)
				require.NoError(t, err)

				request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBody))
				require.NoError(t, err)

				tc.setupAuth(t, request, server.tokenMaker)
				server.router.ServeHTTP(recorder, request)
				tc.checkResponse(t, recorder)
			})
		})
	}
}

func randomRedeem() db.Redeem {

	return db.Redeem{
		ID:        util.RandomInt(1, 1000),
		Createdat: time.Now(),
		Updatedat: time.Now(),
		//	Sender:        pgtype.Text{String: senderUsername, Valid: true},
		//	Receiver:      pgtype.Text{String: receiverUsername, Valid: true},
		//		Amount:        util.RandomMoney(),
		Code:          util.GenerateRandomCode(),
		Transactionid: util.RandomInt(1, 1000),
		//	Status:        db.RedeemStatusPENDING,
	}
}

func TestUseRedeemAPI(t *testing.T) {
	user1, _ := randomUser(t)
	user1.Role = db.NullUserRole{
		UserRole:	db.UserRoleCustomer,
		}
	wallet1 := randomWallet(user1.Username)
	wallet1.Balance=10000

	user2, _ := randomUser(t)
	user2.Role = db.NullUserRole{
		UserRole:	db.UserRoleCustomer,
		}
	wallet2 := randomWallet(user2.Username)
// fmt.Printf("this is the user = %v",user2.Role)
	// redeem := randomRedeem(user1.Username, user2.Username)
	redeem := randomRedeem()

//	amount := util.RandomMoney()
amount := int64(100)
	ChargeForSender := true
	note := "Redeem transaction"
	//fmt.Println(">> Amount:", amount)
	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), ChargeForSender)

	testCases := []struct {
		name string
		//	body          gin.H
		code          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			code: redeem.Code,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user2.Username, string(user2.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user2.Username)).
				//	Times(1).
					Return(wallet2, nil).AnyTimes()

				store.EXPECT().
					GetRedeemWithCode(gomock.Any(), gomock.Eq(redeem.Code)).
				//	Times(1).
					Return(redeem, nil).AnyTimes()

				store.EXPECT().
					GetTransaction(gomock.Any(), gomock.Eq(redeem.Transactionid)).
					//Times(1).
					/*
					Return(db.Transaction{ID: redeem.Transactionid, SenderWalletID: wallet1.ID,
						ReceiverWalletID: pgtype.Int8{Int64: wallet2.ID, Valid: false}, Charge: pgtype.Int8{Int64: int64(chargeFee), Valid: true}, Sendamount: pgtype.Int8{Int64: int64(sendAmount), Valid: true},
						Amount: pgtype.Int8{Int64: amount, Valid: true}, Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true}, 
						Note: pgtype.Text{String: note, Valid: true},Type: db.NullTransactionType{TransactionType: db.TransactionTypeREDEEM,Valid: true}, Status:db.NullTransactionStatus{
							TransactionStatus: db.TransactionStatusPROCESSING,
							Valid:             true,
						} ,*/Return(
						db.Transaction{
							ID: redeem.Transactionid,
							SenderWalletID: wallet1.ID,
							//	ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
							Amount: pgtype.Int8{Int64: amount, Valid: true},
							Charge: pgtype.Int8{Int64: int64(chargeFee), Valid: true},
							Type: db.NullTransactionType{
								TransactionType: db.TransactionTypeREDEEM,
								Valid:           true,
							},
							Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
							Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
							Note: pgtype.Text{
								String: note,
								Valid:  true,
							},
							Status: db.NullTransactionStatus{
								TransactionStatus: db.TransactionStatusPROCESSING,
								Valid:             true,
							},
						}, nil).AnyTimes()
					
				/*
				Note: pgtype.Text{String: note, Valid: true}
					store.EXPECT().
						UpdateTransactionStatus(gomock.Any(), gomock.Any()).
						Times(1).
						Return(nil) */

				arg := db.TransactionParams{
					SenderWalletID: wallet1.ID,
						ReceiverWalletID: pgtype.Int8{Int64: wallet2.ID, Valid: true},
					Amount: pgtype.Int8{Int64: amount, Valid: true},
					Charge: pgtype.Int8{Int64: int64(chargeFee), Valid: true},

					Type: db.NullTransactionType{
						TransactionType: db.TransactionTypeREDEEM,
						Valid:           true,
					},
					Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
					Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
					Note: pgtype.Text{
						String: note,
						Valid:  true,
					},

					Status: db.NullTransactionStatus{
						TransactionStatus: db.TransactionStatusPROCESSING,
						Valid:             true,
					},
				}

			//	fmt.Printf("arg is = %v",arg)
// fmt.Print("i got here 1")
				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Eq(arg), gomock.Eq(false)).AnyTimes()
				//	Times(1)

			//	fmt.Print("i got here 2")
				
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
			},
		},
		{
			name: "Invalid Request - Missing Redemption Code",
			//body: gin.H{},
			//code: redeem.Code,
			code:"",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user2.Username, string(user2.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No store expectations as the request is invalid
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).
					Times(0)
				//	Return(wallet2, nil)

				store.EXPECT().
					GetRedeemWithCode(gomock.Any(), gomock.Any()).
					Times(0)
				//	Return(redeem, nil)

				store.EXPECT().
					GetTransaction(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					RedeemTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		// Add other test cases for different scenarios
		// ...

		{
			name: "Invalid Request - Invalid Redemption Code",
			code: "invalid_code",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user2.Username, string(user2.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {


				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user2.Username)).
				//	Times(1).
					Return(wallet2, nil).AnyTimes()

					
				store.EXPECT().
					GetRedeemWithCode(gomock.Any(), gomock.Eq("invalid_code")).
				//	Times(1).
					Return(db.Redeem{}, db.ErrRecordNotFound).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				store := mockdb.NewMockStore(ctrl)
				tc.buildStubs(store)

				server := newTestServer(t, store, nil)
				recorder := httptest.NewRecorder()

				//	url := "/use-redeem/%code" // Adjust the endpoint URL if needed
				url := fmt.Sprintf("/redeem/%s", tc.code)

				//	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBody))
				request, err := http.NewRequest(http.MethodPost, url, nil)
				require.NoError(t, err)

				tc.setupAuth(t, request, server.tokenMaker)
				server.router.ServeHTTP(recorder, request)
				tc.checkResponse(t, recorder)
			})
		})
	}
}
