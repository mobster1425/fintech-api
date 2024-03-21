package api

import (
	//"context"
	"errors"
	// "fmt"
	//"time"
	//"fmt"
	db "feyin/digital-fintech-api/db/sqlc"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"
)



type P2PRequest struct {
	ReceiverUsername pgtype.Text `json:"receiver_username" binding:"required"`
	Amount           pgtype.Int8 `json:"amount" binding:"required,min=0.01"`
	Note             pgtype.Text `json:"note,omitempty" binding:"omitempty"`
//	VoucherCode      string      `json:"voucher,omitempty" binding:"omitempty"`
//	ChargeForSender bool        `json:"charge,omitempty" binding:"required"`
}


func (server *Server) p2p(ctx *gin.Context) {


	var req P2PRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// Check if req.Amount is negative
	if req.Amount.Int64 < 0 {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("amount must be non-negative")))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	
	senderWallet, err := server.store.GetWalletbyOwner(ctx, authPayload.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		err := errors.New("The receiver does not exist")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if senderWallet.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	receiverWallet, err := server.store.GetWalletbyOwner(ctx, req.ReceiverUsername.String)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		err := errors.New("The receiver does not exist")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ChargeForSender:=true

	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(req.Amount.Int64),ChargeForSender)



	if senderWallet.Balance < int64(sendAmount) {
		err := errors.New("The balance is not enough to make the transaction")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}


	arg := TransferTxParams{
		// SenderWalletID: senderWallet.ID,
		SenderWalletID:   senderWallet.ID,
		ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
		Amount:          pgtype.Int8{Int64: req.Amount.Int64, Valid: true},
		Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
		Type: db.NullTransactionType{
			TransactionType: db.TransactionTypeTRANSFER,
			Valid:           true,
		},
		Sendamount:      pgtype.Int8{Int64: int64(sendAmount), Valid: true},
		Receiveamount:    pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
		// Note:          req.Note,
		Note:          pgtype.Text{String: req.Note.String, Valid: true },
		Status: db.NullTransactionStatus{
			TransactionStatus: db.TransactionStatusPROCESSING,
			Valid:             true,
		},
	}


	actions := []TransferAction{}
	result, err := server.Handletransactions(ctx, arg, false, actions...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
	
	

}