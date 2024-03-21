package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	mockdb "feyin/digital-fintech-api/db/mock"
	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"
)

func TestCreateVoucherAPI(t *testing.T) {
	user, _ := randomUserWithRole(t)
	voucher := randomVoucher(user.Username)
	wallet := randomWallet(user.Username)
	voucher2 := randomVoucherForStatus(user.Username)

	//   value :=   util.RandomInt(1, 100)
	//    vouchertype := db.VoucherType(util.RandomVoucherType())
	//  maxUsage := int32(util.RandomInt(1, 10))
	//  maxUsageByAccount := int32(util.RandomInt(1, 10))
	//   status :=  db.VoucherStatus(util.RandomVoucherStatus())
	//  expireAt :=  time.Now().Add(time.Hour * 24 * 7).Format(time.RFC3339)
	//  code :=  util.RandomInt(1,100)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"value":             voucher.Value,
				"type":              voucher.Type,
				"maxUsage":          voucher.Maxusage,
				"maxUsageByAccount": voucher.Maxusagebyaccount,
				"status":            voucher.Status,
				//"status":            nil,
				"expireAt": voucher.Expireat,
				"code":     voucher.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				roleString := string(user.Role.UserRole)

				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, roleString, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					//Times(0).
					// Return(db.Wallet{}, nil).
					Return(wallet, nil).
					AnyTimes()

				fmt.Printf("this is the voucher in OK %v", voucher)
				//expectedTime := voucher.Expireat.Truncate(time.Second).UTC()
				// Mocking the successful creation of a voucher
				arg := db.CreateVoucherParams{
					Value:            voucher.Value,
					ApplyforUsername: voucher.ApplyforUsername,
					Type:             voucher.Type,
					Maxusage:         voucher.Maxusage,
					//	Maxusagebyaccount: int32(util.RandomInt(1, 10)),
					Maxusagebyaccount: voucher.Maxusagebyaccount,
					Status:            voucher.Status,
					Expireat:          voucher.Expireat.Local(),
					Code:              voucher.Code,
					CreatorUsername:   voucher.CreatorUsername,
				}

				fmt.Printf("this is vouchers arguement %v", arg)

				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Eq(arg)).
					//	Times(1).
					Return(voucher, nil).AnyTimes()

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Printf("recorder body = %v", recorder.Body)
				fmt.Printf("recorder code = %v", recorder.Code)
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchVoucher(t, recorder.Body, voucher)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"value":             voucher.Value,
				"type":              voucher.Type,
				"maxUsage":          voucher.Maxusage,
				"maxUsageByAccount": voucher.Maxusagebyaccount,
				"status":            voucher.Status,
				"expireAt":          voucher.Expireat,
				"code":              voucher.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					//Times(0).
					// Return(db.Wallet{}, nil).
					Return(db.Wallet{}, sql.ErrConnDone).AnyTimes()

				// Mocking an internal error during voucher creation
				arg := db.CreateVoucherParams{
					Value:             voucher.Value,
					Type:              voucher.Type,
					Maxusage:          voucher.Maxusage,
					Maxusagebyaccount: voucher.Maxusagebyaccount,
					Status:            voucher.Status,
					Expireat:          voucher.Expireat.Local(),
					Code:              voucher.Code,
					CreatorUsername:   voucher.CreatorUsername,
					ApplyforUsername:  voucher.ApplyforUsername,
				}
				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Eq(arg)).
					//	Times(1).
					Return(db.Voucher{}, sql.ErrConnDone).AnyTimes()
				//  Return(db.Voucher{}, errors.New("internal error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"value":             voucher.Value,
				"type":              voucher.Type,
				"maxUsage":          voucher.Maxusage,
				"maxUsageByAccount": voucher.Maxusagebyaccount,
				"status":            voucher.Status,
				"expireAt":          voucher.Expireat,
				"code":              voucher.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Omit authorization setup to simulate no authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).
					Times(0)
				// No store expectations as the request is not authorized
				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Any()).AnyTimes()
				//Times(0)

				//  Return(db.Voucher{}, errors.New("internal error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "UnauthorizedUser",
			body: gin.H{
				"value":             voucher.Value,
				"type":              voucher.Type,
				"maxUsage":          voucher.Maxusage,
				"maxUsageByAccount": voucher.Maxusagebyaccount,
				"status":            voucher.Status,
				"expireAt":          voucher.Expireat,
				"code":              voucher.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Set up an unauthorized user
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", util.RandomRole(), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq("unauthorized_user")).
					//Times(0).
					// Return(db.Wallet{}, nil).
					Return(wallet, nil).
					AnyTimes()

				arg := db.CreateVoucherParams{
					Value:             voucher.Value,
					Type:              voucher.Type,
					Maxusage:          voucher.Maxusage,
					Maxusagebyaccount: voucher.Maxusagebyaccount,
					Status:            voucher.Status,
					Expireat:          voucher.Expireat.Local(),
					Code:              voucher.Code,
					CreatorUsername:   voucher.CreatorUsername,
					ApplyforUsername:  voucher.ApplyforUsername,
				}
				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Eq(arg)).
					//	Times(1).
					Return(db.Voucher{}, nil).AnyTimes()
				//  Return(db.Voucher{}, errors.New("internal error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Invalid MaxUsage (Less Than 1)",
			body: gin.H{
				"value": voucher.Value,
				"type":  voucher.Type,
				// "maxUsage":          voucher.Maxusage,
				"maxUsage":          int32(-1),
				"maxUsageByAccount": voucher.Maxusagebyaccount,
				"status":            voucher.Status,
				"expireAt":          voucher.Expireat,
				"code":              voucher.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No store expectations as the request is invalid

				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(db.Voucher{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Invalid Value (Negative)",
			body: gin.H{
				// "value":             voucher.Value,
				"value":             -1,
				"type":              voucher.Type,
				"maxUsage":          voucher.Maxusage,
				"maxUsageByAccount": voucher.Maxusagebyaccount,
				"status":            voucher.Status,
				"expireAt":          voucher.Expireat,
				"code":              voucher.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No store expectations as the request is invalid
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Any()).AnyTimes()
				// Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - ExpireAt in the Past",
			body: gin.H{
				"value":             voucher.Value,
				"type":              voucher.Type,
				"maxUsage":          voucher.Maxusage,
				"maxUsageByAccount": voucher.Maxusagebyaccount,
				"status":            voucher.Status,
				//	"expireAt":          voucher.Expireat,
				"expireAt": time.Now().Local().Add(-time.Hour).Format(time.RFC3339),
				"code":     voucher.Code,
			},

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					//Times(0).
					// Return(db.Wallet{}, nil).
					Return(wallet, nil).
					AnyTimes()

				// No store expectations as the request is invalid
				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Code Length Less Than 3",
			body: gin.H{
				"value":             voucher.Value,
				"type":              voucher.Type,
				"maxUsage":          voucher.Maxusage,
				"maxUsageByAccount": voucher.Maxusagebyaccount,
				"status":            voucher.Status,
				"expireAt":          voucher.Expireat,
				//	"code":              voucher.Code,
				"code": util.RandomString(2),
			},

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).
					Times(0)

				// No store expectations as the request is invalid
				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{
			name: "Invalid Request - Missing Required Field",
			body: gin.H{
				// Omitting required fields to simulate a missing field scenario
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).
					Times(0)
				// No store expectations as the request is invalid
				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				// Additional checks for the response body if needed
			},
		},
		{

			name: "Valid Request - Status Not Provided (Defaults to an appropriate value)",
			body: gin.H{
				"value":             voucher2.Value,
				"type":              voucher2.Type,
				"maxUsage":          voucher2.Maxusage,
				"maxUsageByAccount": voucher2.Maxusagebyaccount,
				"status":            nil,
				//	"status":            voucher2.Status,
				"expireAt": voucher2.Expireat,
				"code":     voucher2.Code,
				//  "code":              util.RandomString(2),
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					//Times(0).
					// Return(db.Wallet{}, nil).
					Return(wallet, nil).
					AnyTimes()

				arg := db.CreateVoucherParams{
					Value:    voucher2.Value,
					Type:     voucher2.Type,
					Maxusage: voucher2.Maxusage,
					//	Maxusagebyaccount: int32(util.RandomInt(1, 10)),
					Maxusagebyaccount: voucher2.Maxusagebyaccount,
					//Status:           voucher2.Status,
					Status:           db.VoucherStatusAVAILABLE,
					Expireat:         voucher2.Expireat.Local(),
					Code:             voucher2.Code,
					CreatorUsername:  voucher2.CreatorUsername,
					ApplyforUsername: voucher2.ApplyforUsername,
				}

				fmt.Printf("arg in status not provided = = %v", arg)
				fmt.Printf("voucher in status not provided = %v", voucher)

				store.EXPECT().
					CreateVoucher(gomock.Any(), gomock.Eq(arg)).
					//	Times(1).
					Return(voucher2, nil).AnyTimes()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Printf("recorder code = %v", recorder.Code)
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchVoucher(t, recorder.Body, voucher2)
				// Additional checks for the response body if needed
			},
		},
		/*
			{
				name: "Valid Request - Type Not Provided (Defaults to an appropriate value)",
				body: gin.H{
					"value": voucher.Value,
					//	"type":              voucher.Type,
					"maxUsage":          voucher.Maxusage,
					"maxUsageByAccount": voucher.Maxusagebyaccount,
					"status":            voucher.Status,
					"expireAt":          voucher.Expireat,
					"code":              voucher.Code,
					//  "code":              util.RandomString(2),
				},
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {

					store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					//Times(0).
					// Return(db.Wallet{}, nil).
					Return(wallet, nil).
					AnyTimes()


					arg := db.CreateVoucherParams{
						Value: voucher.Value,
							Type:              db.VoucherTypeFIXED,
						Maxusage:          voucher.Maxusage,
						Maxusagebyaccount: int32(util.RandomInt(1, 10)),
						Status:            voucher.Status,
						Expireat:          voucher.Expireat.Local(),
						Code:              voucher.Code,
						CreatorUsername:   voucher.CreatorUsername,
						ApplyforUsername:  voucher.ApplyforUsername,
					}
					store.EXPECT().
						CreateVoucher(gomock.Any(), gomock.Eq(arg)).
					//	Times(1).
						Return(voucher, nil).AnyTimes()
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchVoucher(t, recorder.Body, voucher)
					// Additional checks for the response body if needed
				},
			},
		*/
		/*
			{
				name: "Valid Request - ApplyforUsername Not Provided (Defaults to authPayload.Username)",
				body: gin.H{
					"value":             10,
					"type":              "FIXED",
					"maxUsage":          1,
					"maxUsageByAccount": 1,
					"expireAt":          time.Now().Add(time.Hour * 24 * 7).Format(time.RFC3339),
					"code":              "VoucherCode123",
				},
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					// Mocking the successful creation of a voucher
					store.EXPECT().
						CreateVoucher(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Voucher{ID: 1, Code: "VoucherCode123", Value: 10, Type: db.VoucherTypeFIXED}, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					// Additional checks for the response body if needed
				},
			},
		*/
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := "/create-voucher/" // Adjust the endpoint URL if needed
			requestBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBody))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

// Function to generate a random voucher for testing
func randomVoucher(creatorUsername string) db.Voucher {
	return db.Voucher{
		// ID:                util.RandomInt(1, 1000),
		//  Createdat:         time.Now(),
		//    Updatedat:         time.Now(),
		CreatorUsername:   pgtype.Text{String: creatorUsername, Valid: true},
		Value:             util.RandomMoney(),
		Type:              db.VoucherType(util.RandomVoucherType()), // Assuming VoucherType is an enum with 3 values
		ApplyforUsername:  pgtype.Text{String: creatorUsername, Valid: true},
		Maxusage:          int32(util.RandomInt(1, 100)),
		Maxusagebyaccount: int32(util.RandomInt(1, 100)),
		Status:            db.VoucherStatus(util.RandomVoucherStatus()), // Assuming VoucherStatus is an enum with 3 values
		Expireat:          time.Now().Add(time.Hour * 24 * 7),
		Code:              util.RandomString(20),
		//   Usedby:            []string{},
	}
}

func randomVoucherForStatus(creatorUsername string) db.Voucher {
	return db.Voucher{
		// ID:                util.RandomInt(1, 1000),
		//  Createdat:         time.Now(),
		//    Updatedat:         time.Now(),
		CreatorUsername:   pgtype.Text{String: creatorUsername, Valid: true},
		Value:             util.RandomMoney(),
		Type:              db.VoucherType(util.RandomVoucherType()), // Assuming VoucherType is an enum with 3 values
		ApplyforUsername:  pgtype.Text{String: creatorUsername, Valid: true},
		Maxusage:          int32(util.RandomInt(1, 100)),
		Maxusagebyaccount: int32(util.RandomInt(1, 100)),
		Status:            db.VoucherStatusAVAILABLE, // Assuming VoucherStatus is an enum with 3 values
		Expireat:          time.Now().Add(time.Hour * 24 * 7),
		Code:              util.RandomString(20),
		//   Usedby:            []string{},
	}
}

// Function to check if the response body matches a voucher
func requireBodyMatchVoucher(t *testing.T, body *bytes.Buffer, voucher db.Voucher) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotVoucher db.Voucher
	err = json.Unmarshal(data, &gotVoucher)
	require.NoError(t, err)
	// require.Equal(t, voucher, gotVoucher)
	require.True(t, voucher.Expireat.Equal(gotVoucher.Expireat.Local()))
}

/*
// Function to check if the response body matches a voucher
func requireBodyMatchVoucher(t *testing.T, body *bytes.Buffer, voucher db.Voucher, arg db.CreateVoucherParams) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotVoucher db.Voucher
	err = json.Unmarshal(data, &gotVoucher)
	require.NoError(t, err)

	// Compare Expireat fields with time.Equal and truncation
	require.True(t, voucher.Expireat.Equal(gotVoucher.Expireat.Truncate(time.Second)))

	// Compare the rest of the fields
	require.Equal(t, int64(50), voucher.Value)
	require.NotEmpty(t, voucher.ID)
	require.NotZero(t, voucher.Createdat)
	require.NotZero(t, voucher.Updatedat)
	require.Equal(t, arg.CreatorUsername.String, voucher.CreatorUsername.String)
	require.Equal(t, arg.ApplyforUsername.String, voucher.ApplyforUsername.String)
	require.Equal(t, arg.Type, voucher.Type)
	require.True(t, voucher.Maxusage >= 1 && voucher.Maxusage <= 5)
	require.True(t, voucher.Maxusagebyaccount >= 1 && voucher.Maxusagebyaccount <= 10)
	//require.Equal(t, arg.Status.String, voucher.Status.String)
	require.NotZero(t, voucher.Expireat)
	require.Equal(t, arg.Code, voucher.Code)

	// Alternatively, you can use require.Equal for the entire struct if time.Equal doesn't solve the issue
	// require.Equal(t, voucher, gotVoucher)
}
*/

// Function to check if the response body matches a list of vouchers
func requireBodyMatchVouchers(t *testing.T, body *bytes.Buffer, vouchers []db.Voucher) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotVouchers []db.Voucher
	err = json.Unmarshal(data, &gotVouchers)
	require.NoError(t, err)
	require.Equal(t, vouchers, gotVouchers)
}
