package api

import (
	"fmt"

	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"
	"feyin/digital-fintech-api/worker"

	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	router          *gin.Engine
	taskDistributor worker.TaskDistributor
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	/*
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			v.RegisterValidation("currency", validCurrency)
		}
	*/

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	
		router.POST("/users", server.createUser)
		router.POST("/users/login", server.loginUser)
		router.POST("/tokens/renew_access", server.renewAccessToken)
		router.POST("/verify_email/", server.VerifyEmail)


		authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker, []string{string(db.UserRoleMerchant), string(db.UserRoleCustomer)}))
		authRoutes.PUT("/users/update", server.UpdateUser)
        authRoutes.GET("/wallet/:id", server.getWallet)
        authRoutes.POST("/wallet/add-money",server.AddMoneyWallet)
        authRoutes.POST("/p2p-payment/",server.p2p)
		


		voucherAuthRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker, []string{string(db.UserRoleMerchant)}))
        voucherAuthRoutes.POST("/create-voucher/",server.createVoucher)
	
        redeemAuthRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker, []string{string(db.UserRoleCustomer)}))
         redeemAuthRoutes.POST("/redeem/",server.createRedeem)
		 redeemAuthRoutes.POST("/redeem/:code",server.useRedeem)
        
         payAuthRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker, []string{string(db.UserRoleCustomer)}))
        payAuthRoutes.POST("/pay-to-merchant/",server.payMerchant)


	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
