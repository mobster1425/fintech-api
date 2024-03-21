package api

import (
	//"context"
	"errors"
	"fmt"

	//"fmt"
	db "feyin/digital-fintech-api/db/sqlc"

	"feyin/digital-fintech-api/token"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"feyin/digital-fintech-api/util"
)

type createRedeemRequest struct {
	ChargeForSender bool        `json:"charge,omitempty" binding:"required"`
	Amount          pgtype.Int8 `json:"Amount,omitempty" binding:"required"`
	Note            pgtype.Text `json:"Note,omitempty" binding:"required"`
}

func (server *Server) createRedeem(ctx *gin.Context) {
	var req createRedeemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// Check if req.Amount is negative
	if req.Amount.Int64 < 0 {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("amount must be non-negative")))
		return
	}
	//code := util.GenerateRandomCode()

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	senderWallet, err := server.store.GetWalletbyOwner(ctx, authPayload.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if senderWallet.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(req.Amount.Int64), req.ChargeForSender)

	if senderWallet.Balance < int64(sendAmount) {
		err := errors.New("The balance is not enough to make the transaction")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := TransferTxParams{
		SenderWalletID: senderWallet.ID,
		Amount:         pgtype.Int8{Int64: req.Amount.Int64, Valid: true},
		Charge:         pgtype.Int8{Int64: int64(chargeFee), Valid: true},
		Type: db.NullTransactionType{
			TransactionType: db.TransactionTypeREDEEM,
			Valid:           true,
		},
		Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
		Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
		Note:          req.Note,
		Status: db.NullTransactionStatus{
			TransactionStatus: db.TransactionStatusPROCESSING,
			Valid:             true,
		},
	}

	/*
		newTransaction, err := server.store.CreateTransaction(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	*/

	actions := []TransferAction{CREATE_CODE}

	result, err := server.Handletransactions(ctx, arg, true, actions...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	fmt.Printf("result = %v", result)
	ctx.JSON(http.StatusOK, result)

}

// use code

type useRedeemmRequest struct {
	Code string `uri:"code" binding:"required"`
}

func (server *Server) useRedeem(ctx *gin.Context) {

	var req useRedeemmRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	receiverWallet, err := server.store.GetWalletbyOwner(ctx, authPayload.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if receiverWallet.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	resultRedeem, err := server.store.GetRedeemWithCode(ctx, req.Code)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
	}

	resultTransaction, err := server.store.GetTransaction(ctx, resultRedeem.Transactionid)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
	}

	fmt.Printf("result trnasaction = %v", resultTransaction)
	//resultTransaction.Status.Valid &&
	if resultTransaction.ReceiverWalletID.Valid == true || (resultTransaction.Status.Valid && resultTransaction.Status.TransactionStatus == db.TransactionStatusSUCCESS) {
		err := errors.New("code is used")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := TransferTxParams{
		// SenderWalletID: senderWallet.ID,
		SenderWalletID:   resultTransaction.SenderWalletID,
		ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
		Amount:           resultTransaction.Amount,
		Charge:           resultTransaction.Charge,
		Type:             resultTransaction.Type,
		Sendamount:       resultTransaction.Sendamount,
		Receiveamount:    resultTransaction.Receiveamount,
		Note:             resultTransaction.Note,
		Status:           resultTransaction.Status,
	}

	actions := []TransferAction{USE_CODE}

	result, err := server.Handletransactions(ctx, arg, false, actions...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	fmt.Print(result)

	ctx.JSON(http.StatusOK, nil)

}
