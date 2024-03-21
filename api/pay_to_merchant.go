package api

import (
	//"context"
	"errors"
	// "fmt"
	"time"
	//"fmt"
	db "feyin/digital-fintech-api/db/sqlc"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"feyin/digital-fintech-api/token"
	"feyin/digital-fintech-api/util"
)

type PayRequest struct {
	ReceiverUsername pgtype.Text `json:"ReceiverUsername" binding:"required"`
	Amount           pgtype.Int8 `json:"amount" binding:"required,min=0.01"`
	Note             pgtype.Text `json:"note,omitempty" binding:"omitempty"`
// Note             pgtype.Text `json:"note" binding:"required"`
	VoucherCode      string      `json:"voucher,omitempty" binding:"omitempty"`
// VoucherCode      string      `json:"voucher" binding:"required"`
}

func (server *Server) payMerchant(ctx *gin.Context) {

	var req PayRequest
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

	receiverName, err := server.store.GetUser(ctx, req.ReceiverUsername.String)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if receiverName.Role.UserRole != db.UserRoleMerchant {
		err := errors.New("The User is not a merchant")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

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

	// amount := req.Amount
	var resultVoucher db.Voucher

	var amount int64 = req.Amount.Int64

	if req.VoucherCode != "" {
		var err error
		arg1 := db.GetVoucherWithCodeParams{
			CreatorUsername: req.ReceiverUsername,
			Code:            req.VoucherCode,
		}
		resultVoucher, err = server.store.GetVoucherWithCode(ctx, arg1)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, errorResponse(err))
				return
			}
			err := errors.New("The voucher does not exist")
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		/*
			       if resultVoucher.Status == db.VoucherStatusUNAVAILABLE {
					err := errors.New("The voucher is unavailable")
					ctx.JSON(http.StatusUnauthorized, errorResponse(err))
					return
				   }
		*/
		// T compare expiresAt with current time, even if status changed
		if resultVoucher.Status == db.VoucherStatusUNAVAILABLE && resultVoucher.Expireat.After(time.Now()) {
			err := errors.New("The voucher is unavailable")
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		//  check max usage
		if resultVoucher.Maxusage > 0 && resultVoucher.Maxusagebyaccount >= resultVoucher.Maxusage {
			err := errors.New("The voucher has reached its maximum usage limit")
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		arg := db.CreateVoucherParams{
			Value:             resultVoucher.Value,
			ApplyforUsername:  resultVoucher.ApplyforUsername,
			Type:              resultVoucher.Type,
			Maxusage:          resultVoucher.Maxusage,
			Maxusagebyaccount: resultVoucher.Maxusagebyaccount,
			Status:            resultVoucher.Status,
			Expireat:          resultVoucher.Expireat,
			Code:              resultVoucher.Code,
			CreatorUsername:   resultVoucher.CreatorUsername,
		}

		// change amount by voucher value
		amount = CalculateAmountWithVoucher(amount, arg)
	}

	// calculate after charge
	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), false)

	if senderWallet.Balance < int64(sendAmount) {
		err := errors.New("The balance is not enough to make the transaction")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := TransferTxParams{
		SenderWalletID:   senderWallet.ID,
		ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
		Amount:           pgtype.Int8{Int64: req.Amount.Int64, Valid: true},
		Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
		Type: db.NullTransactionType{
			TransactionType: db.TransactionTypePAYMENTVOUCHER,
			Valid:           true,
		},
		Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
		Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
		Note:          req.Note,
		Status: db.NullTransactionStatus{
			TransactionStatus: db.TransactionStatusPROCESSING,
			Valid:             true,
		},
		VoucherID:      resultVoucher.ID,
		UsedByUsername: []string{authPayload.Username},
	}

	actions := []TransferAction{}
	result, err := server.Handletransactions(ctx, arg, false, actions...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)

}
