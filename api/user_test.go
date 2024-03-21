package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"

	//	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	//mockdb "github.com/techschool/simplebank/db/mock"
	mockdb "feyin/digital-fintech-api/db/mock"
	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"
	"feyin/digital-fintech-api/worker"
	mockwk "feyin/digital-fintech-api/worker/mock"
)

type eqCreateUserTxParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
	user     db.User
}

func (e eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.Password)
	if err != nil {
		return false
	}

	e.arg.Password = arg.Password
	if !reflect.DeepEqual(e.arg.CreateUserParams, arg.CreateUserParams) {
		return false
	}

	err = arg.AfterCreate(e.user)
	return err == nil
}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}
func EqCreateUserTxParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, password, user}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)
	wallet1 := randomWallet(user.Username)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
				//	"full_name": user.FullName,
				"role":  user.Role.UserRole,
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						//	FullName: user.FullName,
						Email: user.Email,
						Role: db.NullUserRole{
							UserRole: user.Role.UserRole,
							Valid:    true,
						},
					},
				}
				fmt.Printf("arg =          %v", arg)

				arg1 := db.CreateWalletTxParams{
					CreateWalletParams: db.CreateWalletParams{
						//	Balance: wallet1.Balance,
						Balance: 0,
						Owner:   wallet1.Owner,
					},
				}
				fmt.Printf("arg1 =          %v", arg1)
				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, password, user), arg1).
					//	Times(1).
					Return(db.CreateUserTxResult{User: user}, nil).AnyTimes()

				taskPayload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}

				/*
					opts := []asynq.Option{
						asynq.MaxRetry(10),
						asynq.ProcessIn(10 * time.Second),
						asynq.Queue(worker.QueueCritical),
					}*/
				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).
					// Times(1).
					Return(nil).AnyTimes()
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				fmt.Printf("recorder body = %v", recorder.Body)
				fmt.Printf("recorder code = %v", recorder.Code)
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"role":     user.Role.UserRole,
				//	"full_name": user.FullName,
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username": user.Username,
				"password": password,
				//	"full_name": user.FullName,
				"role":  user.Role.UserRole,
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, db.ErrUniqueViolation)

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "invalid-user#1",
				"password": password,
				"role":     user.Role.UserRole,
				//	"full_name": user.FullName,
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any(), gomock.Any()).
					//	Times(1).
					Return(db.CreateUserTxResult{}, db.ErrUniqueViolation).AnyTimes()

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				// Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username": user.Username,
				"password": password,
				//"full_name": user.FullName,
				"role":  user.Role.UserRole,
				"email": "invalid-email",
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any(), gomock.Any()).
					//	Times(1).
					Return(db.CreateUserTxResult{}, db.ErrUniqueViolation).AnyTimes()

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username": user.Username,
				"password": "123",
				//	"full_name": user.FullName,
				"role":  user.Role.UserRole,
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any(), gomock.Any()).
					//	Times(1).
					Return(db.CreateUserTxResult{}, db.ErrUniqueViolation).AnyTimes()

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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
			//	tc.buildStubs(store)

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(store, taskDistributor)
			server := newTestServer(t, store, taskDistributor)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestLoginUserAPI(t *testing.T) {
	user, password := randomUser(t)
	// fmt.Printf("password = %v",password)
	wallet := randomWallet(user.Username)
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					//		Times(1).
					Return(user, nil).AnyTimes()

				arg := db.UpdateUserStatusParams{
					Username: user.Username,
					Status: db.NullUserStatus{
						UserStatus: db.UserStatusActive,
						Valid:      true,
					},
				}
				store.EXPECT().
					UpdateUserStatus(gomock.Any(), gomock.Eq(arg)).
					Return(user, nil).AnyTimes()

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(1)

				store.EXPECT().GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					Return(wallet, nil).AnyTimes()

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				fmt.Printf("recorder body = %v", recorder.Body)
				fmt.Printf("recorder code = %v", recorder.Code)
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UserNotFound",
			body: gin.H{
				"username": "NotFound",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					//	Times(1).
					Return(db.User{}, db.ErrRecordNotFound).AnyTimes()
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "IncorrectPassword",
			body: gin.H{
				"username": user.Username,
				"password": "incorrect",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					//	Times(1).
					Return(user, nil).AnyTimes()

				arg := db.UpdateUserStatusParams{
					Username: user.Username,
					Status: db.NullUserStatus{
						UserStatus: db.UserStatusActive,
						Valid:      true,
					},
				}
				store.EXPECT().
					UpdateUserStatus(gomock.Any(), gomock.Eq(arg)).AnyTimes()
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					//	Times(1).
					Return(db.User{}, sql.ErrConnDone).AnyTimes()
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "invalid-user#1",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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

			//	taskCtrl := gomock.NewController(t)
			//	defer taskCtrl.Finish()
			//	taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			//		server := newTestServer(t, store, taskDistributor)
			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users/login"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)

	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	rrole := util.RandomRole()
	user = db.User{
		Username: util.RandomOwner(),
		Password: hashedPassword,
		//FullName:       util.RandomOwner(),
		Email: util.RandomEmail(),

		Role: db.NullUserRole{
			UserRole: db.UserRole(rrole),
			Valid:    true,
		},
	}
	return
}

func randomUserWithRole(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	// rrole:=util.RandomRole()
	user = db.User{
		Username: util.RandomOwner(),
		Password: hashedPassword,
		//FullName:       util.RandomOwner(),
		Email: util.RandomEmail(),
		Role: db.NullUserRole{
			UserRole: db.UserRoleMerchant,
			Valid:    true,
		},
	}
	return
}

/*
func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	//	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
//	require.Empty(t, gotUser.Password)
}*/

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var userResponse struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		// Add other fields from db.User as needed
	}

	err = json.Unmarshal(data, &userResponse)
	require.NoError(t, err)

	// Convert the role string to db.NullUserRole
	var role db.NullUserRole
	switch userResponse.Role {
	case "customer":
		role = db.NullUserRole{UserRole: db.UserRoleCustomer, Valid: true}
	case "merchant":
		role = db.NullUserRole{UserRole: db.UserRoleMerchant, Valid: true}
	default:
		role = db.NullUserRole{}
	}

	// Construct the db.User object
	gotUser := db.User{
		Username: userResponse.Username,
		Email:    userResponse.Email,
		Role:     role,
		// Add other fields from db.User as needed
	}

	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.Role, gotUser.Role)
	// Add assertions for other fields as needed
}

func TestUpdateUserAPI(t *testing.T) {
	user, password := randomUser(t)
	wallet1 := randomWallet(user.Username)
	other, _ := randomUser(t)
	//	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	invalidEmail := "invalid-email"
	//  pswd :=util.RandomString(6)
	// hshpswd,_ := util.HashPassword(pswd)
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{

		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": "",
				// "email": user.Email,
				"email": newEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					//		Times(1).
					Return(user, nil).AnyTimes()

				arg := db.UpdateUserParams{
					Username: user.Username,
					/*
						Password: pgtype.Text{
							String: hshpswd,
							Valid:  true,
						},*/
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
				}
				updatedUser := db.User{
					Username: user.Username,
					Password: user.Password,
					//	FullName:          newName,
					Email:           newEmail,
					Updatedat:       user.Updatedat,
					Createdat:       user.Createdat,
					IsEmailVerified: user.IsEmailVerified,
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					//	Times(1).
					Return(updatedUser, nil).AnyTimes()

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Eq(user.Username)).
					//	Times(1).
					Return(wallet1, nil).AnyTimes()

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				fmt.Printf("recorder body = %v", recorder.Body)
				fmt.Printf("recorder code = %v", recorder.Code)
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},

		{
			name: "OtherUsersCannotUpdateThisUserInfo",
			body: gin.H{
				"username": user.Username,
				"password": password,
				//	"email": user.Email,
				"email": newEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, other.Username, string(other.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
                  store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).AnyTimes()
					//		Times(1)..AnyTimes()

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//Return(updatedUser, nil)

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(wallet1, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "UserNotFound",
			body: gin.H{
				"username": user.Username,
				"password": "",
				"email":    newEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
				GetUser(gomock.Any(), gomock.Any()).Return(db.User{}, db.ErrRecordNotFound).AnyTimes()
				//		Times(1)..AnyTimes()

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					//	Times(1).
					Return(db.User{}, db.ErrRecordNotFound).AnyTimes()

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).
					// Times(1).
					Return(db.Wallet{}, db.ErrRecordNotFound).AnyTimes()

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				fmt.Printf("recorder body =  %v",recorder.Body)
				fmt.Printf("recorder code =  %v",recorder.Code)
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		{
			name: "InvalidEmail",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"email":    invalidEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
				GetUser(gomock.Any(), gomock.Any()).Return(db.User{}, db.ErrRecordNotFound).AnyTimes()

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(db.User{}, db.ErrRecordNotFound)

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(db.Wallet{}, db.ErrRecordNotFound)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "ExpiredToken",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"email":    newEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), -time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
				GetUser(gomock.Any(), gomock.Any()).Return(db.User{}, db.ErrRecordNotFound).AnyTimes()

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(db.User{}, db.ErrRecordNotFound)

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				// Times(0)
				//	Return(db.Wallet{}, db.ErrRecordNotFound)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "NoAuthorization",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"email":    newEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//	addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, string(user.Role.UserRole), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
               store.EXPECT().
				GetUser(gomock.Any(), gomock.Any()).Return(db.User{}, db.ErrRecordNotFound).AnyTimes()

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(db.User{}, db.ErrRecordNotFound)

				store.EXPECT().
					GetWalletbyOwner(gomock.Any(), gomock.Any()).AnyTimes()
				//	Times(0)
				//	Return(db.Wallet{}, db.ErrRecordNotFound)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			//	tc.buildStubs(store)

			//	taskCtrl := gomock.NewController(t)
			//	defer taskCtrl.Finish()
			//	taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(store)
			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users/update"
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}

}
