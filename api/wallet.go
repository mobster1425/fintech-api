package api

import (
	//"context"
	"errors"
	//"fmt"
	"net/http"
	//	"github.com/jackc/pgx/v5/pgtype"
	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"

	"github.com/gin-gonic/gin"
)

type createWallettRequest struct {
	// Currency string `json:"currency" binding:"required,currency"`
}

type getWalletRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getWallet(ctx *gin.Context) {
	var req getWalletRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	wallet, err := server.store.GetWallet(ctx, req.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if wallet.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, wallet)
}

// deposit money

type addWalletBalanceRequest struct {
	Amount int64 `json:"amount" binding:"required"`
}

func (server *Server) AddMoneyWallet(ctx *gin.Context) {
	var req addWalletBalanceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// Check if req.Amount is negative
	if req.Amount < 0 {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("amount must be non-negative")))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	wallet, err := server.store.GetWalletbyOwner(ctx, authPayload.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if wallet.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	arg := db.AddWalletBalanceParams{
		ID:     wallet.ID,
		Amount: req.Amount,
	}

	result, err := server.store.AddWalletBalance(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)

}
