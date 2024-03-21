package api

import (
	"errors"
	"fmt"
	//	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"

	"feyin/digital-fintech-api/worker"

	"github.com/hibiken/asynq"
)

type createUserRequest struct {
	Username string          `json:"username,omitempty" binding:"required,alphanum"`
	Password string          `json:"password,omitempty" binding:"required,min=6"`
	Email    string          `json:"email,omitempty" binding:"required,email"`
	Role     db.UserRole `json:"role,omitempty" binding:"required"`
// Role     db.NullUserRole `json:"role" binding:"omitempty"`
	//Status          NullUserStatus `json:"status" binding:"required"`
	//IsEmailVerified bool           `json:"is_email_verified" binding:"required"`

}

type userResponse struct {
	Username        string            `json:"username,omitempty"`
	Email           string            `json:"email,omitempty"`
	Createdat       time.Time         `json:"createdat,omitempty"`
	Updatedat       time.Time         `json:"updatedat,omitempty"`
	Role            db.UserRole   `json:"role,omitempty" `
	Status          db.UserStatus `json:"status,omitempty" `
	IsEmailVerified bool              `json:"is_email_verified,omitempty"`
	Owner           string            `json:"owner,omitempty"`
	Balance         int64             `json:"balance,omitempty"`
}

func newUserResponse(user db.User, wallet db.Wallet) *userResponse {
	return &userResponse{
		Username:        user.Username,
		Createdat:       user.Createdat,
		Updatedat:       user.Updatedat,
		Role:            user.Role.UserRole,
		Status:          user.Status.UserStatus,
		IsEmailVerified: user.IsEmailVerified,
		Email:           user.Email,
		Owner:           wallet.Owner,
		Balance:         wallet.Balance,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
//	fmt.Printf("hey")
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}


// fmt.Printf("request from the test in the api is %v", req)

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg2 := db.CreateWalletTxParams{
		CreateWalletParams: db.CreateWalletParams{
			Owner:   req.Username,
			Balance: 0,
		},
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username: req.Username,
			Password: hashedPassword,
			Email:    req.Email,
			Role:     db.NullUserRole{
				UserRole: req.Role,
				Valid: true,
			},
		
			//IsEmailVerified: req.,
		},

		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	userResult, err := server.store.CreateUserTx(ctx, arg, arg2)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Extracting db.User from db.CreateUserTxResult
	user := userResult.User
	wallet := userResult.Wallet
	rsp := newUserResponse(user, wallet)

	ctx.JSON(http.StatusOK, rsp)
}

//UPDATE USER

type UpdateUserRequest struct {
	Username string `json:"username,omitempty" binding:"required,alphanum"`
	Password string `json:"password,omitempty" binding:"omitempty,min=6"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	// Role     db.NullUserRole `json:"role" binding:"required"`
	//Status          NullUserStatus `json:"status" binding:"required"`
	//IsEmailVerified bool           `json:"is_email_verified" binding:"required"`

}

type UpdateuserResponse struct {
	User userResponse `json:"user,omitempty"`
}

func (server *Server) UpdateUser(ctx *gin.Context) {

	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if req.Username != authPayload.Username {
		err := errors.New("Username doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}



	
	_, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	pswrd := req.Password
fmt.Printf("pswrd = %v",pswrd)
	arg := db.UpdateUserParams{
		Username: req.Username,
		Email: pgtype.Text{
			String: req.Email,
			Valid:  req.Email != "",
		},
	}

	if req.Password != "" {
		hashedPassword, err := util.HashPassword(pswrd)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		arg.Password = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}

		arg.Updatedat = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}
fmt.Printf("arg username is %v",arg.Username)
fmt.Printf("arg email is %v",arg.Email)
fmt.Printf("arg Password is %v",arg.Password)
	userResult, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return

	}

	walletResult, err := server.store.GetWalletbyOwner(ctx, authPayload.Username)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return

	}

	rsp := newUserResponse(userResult, walletResult)

	ctx.JSON(http.StatusOK, rsp)
}

// LOGIN USER

type loginUserRequest struct {
	Username string `json:"username,omitempty" binding:"required,alphanum"`
	Password string `json:"password,omitempty" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID     `json:"session_id,omitempty"`
	AccessToken           string        `json:"access_token,omitempty"`
	AccessTokenExpiresAt  time.Time     `json:"access_token_expires_at,omitempty"`
	RefreshToken          string        `json:"refresh_token,omitempty"`
	RefreshTokenExpiresAt time.Time     `json:"refresh_token_expires_at,omitempty"`
	User                  *userResponse `json:"user,omitempty"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fmt.Printf("login request = %v",req)

	_, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := db.UpdateUserStatusParams{
		Username: req.Username,
		Status: db.NullUserStatus{
			UserStatus: db.UserStatusActive,
			Valid:      true,
		},
	}

	fmt.Printf("hey i reached here 1")
	user, err := server.store.UpdateUserStatus(ctx, arg)
	fmt.Printf("hey i reached here 12")
	fmt.Printf("this is the user =              %v",user)
fmt.Printf("user password = %v",user.Password)

	err = util.CheckPassword(req.Password, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	fmt.Printf("hey i passed here")

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		string(user.Role.UserRole),
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		string(user.Role.UserRole),
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	walletResult, err := server.store.GetWalletbyOwner(ctx, user.Username)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return

	}

	rsp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user, walletResult),
	}
	ctx.JSON(http.StatusOK, rsp)
}

/*

Access tokens, refresh tokens, and sessions are components used in authentication systems to manage user identity and authorization. Let's break down the differences and provide a real-life example of a user logging into a web app:

1. **Access Token:**
   - **Purpose:** An access token is a short-lived token that grants access to specific resources on behalf of a user.
   - **Usage:** It is sent with each request to the server to access protected resources or perform specific actions.
   - **Lifespan:** Typically, it has a short lifespan to minimize the risk if it gets compromised.
   - **Example:** A user logs in, and upon successful authentication, the server issues an access token. The user includes this token in the header of subsequent requests to access their account information, make API calls, or perform other authorized actions.

2. **Refresh Token:**
   - **Purpose:** A refresh token is a long-lived token used to obtain a new access token after the original access token expires.
   - **Usage:** It is securely stored and sent to the server when the access token expires to request a new one.
   - **Lifespan:** Longer lifespan compared to access tokens.
   - **Example:** After receiving the access token, the user's client-side application stores the refresh token securely. When the access token expires, the application sends the refresh token to the server, which issues a new access token without requiring the user to log in again.

3. **Session:**
   - **Purpose:** A session is a server-side mechanism to maintain state and user identity across multiple requests.
   - **Usage:** It often involves creating a session identifier that is sent to the client and stored as a cookie. The server uses this identifier to associate requests with a specific user's session.
   - **Lifespan:** It can last for the duration of a user's visit or persist longer based on server configurations.
   - **Example:** When a user logs in, the server creates a session, assigns a session ID, and sends it to the client as a cookie. Subsequent requests from the client include this session ID, allowing the server to identify the user and maintain their authenticated state.

In your provided code snippet, after a user logs in:
- An access token is issued for short-lived authorization.
- A refresh token is issued for obtaining a new access token when needed.
- A session is created in the server store, associating the user's session with the refresh token, user agent, client IP, etc. This server-side session helps maintain the user's authenticated state.



*/


	type VerifyEmailRequest struct {
		EmailId    int64  `form:"email_id" binding:"required"`
		SecretCode string `form:"secret_code" binding:"required"`
	}
	

type VerifyEmailResponse struct {
	IsVerified bool `json:"is_verified,omitempty"`
}

func (server *Server) VerifyEmail(ctx *gin.Context) {

	var req VerifyEmailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.EmailId,
		SecretCode: req.SecretCode,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}
	ctx.JSON(http.StatusOK, rsp)
}
