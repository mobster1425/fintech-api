package api

import (
	//"context"
	"errors"
	//"fmt"
	"time"
	//"fmt"
	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/token"

	//	"feyin/digital-fintech-api/token"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	// "feyin/digital-fintech-api/token"
	//"feyin/digital-fintech-api/util"
)

type createVoucherRequest struct {
	Value float64          `json:"value" binding:"required,min=0.01"`
	Type  db.VoucherType `json:"type" binding:"omitempty"`
	//ApplyforUsername          string         `json:"applyFor_username" binding:"omitempty,uuid"`
	MaxUsage          int32            `json:"maxUsage" binding:"required,min=1"`
	MaxUsageByAccount int32            `json:"maxUsageByAccount" binding:"required,min=1"`
	Status            *db.VoucherStatus `json:"status" binding:"omitempty"`
	ExpireAt          time.Time        `json:"expireAt" binding:"required"`
	Code              string           `json:"code" binding:"required,min=3"`
}

func (server *Server) createVoucher(ctx *gin.Context) {
	var req createVoucherRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
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
/*
	status := req.Status
	if status == db.VoucherStatus(rune(0)) {
		status = db.VoucherStatusAVAILABLE
	} */
	// TODO
	// validate code
	if len(req.Code) < 3 && req.MaxUsage <1 && req.Value < 1{
		err := errors.New("Code must be at least 3 characters long")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// TODO: Validate expiresDate
	if req.ExpireAt.Before(time.Now()) {
		err := errors.New("Expiration date must be in the future")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateVoucherParams{
		Value:             int64(req.Value),
		ApplyforUsername:  pgtype.Text{String: authPayload.Username, Valid: true},
		Type:              req.Type,
		Maxusage:          req.MaxUsage,
		Maxusagebyaccount: req.MaxUsageByAccount,
	//	Status:            *req.Status,
	//Status:            status,
		Expireat:          req.ExpireAt,
		Code:              req.Code,
		CreatorUsername:   pgtype.Text{String: authPayload.Username, Valid: true}, 
	}


	// Setting default status if Status is nil
if req.Status == nil {
    defaultStatus := db.VoucherStatusAVAILABLE
    arg.Status = defaultStatus
} else {
    arg.Status = *req.Status
}

	result, err := server.store.CreateVoucher(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)

}

func CalculateAmountWithVoucher(amount int64, voucher db.CreateVoucherParams) int64 {
	var amountWithVoucher int64

	switch voucher.Type {
	case db.VoucherTypeFIXED:
		amountWithVoucher = amount - voucher.Value
		if amountWithVoucher < 0 {
			amountWithVoucher = 0
		}
	default:
		amountWithVoucher = amount - amount*voucher.Value
	}

	return amountWithVoucher
}
