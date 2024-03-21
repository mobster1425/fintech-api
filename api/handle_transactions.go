package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	db "feyin/digital-fintech-api/db/sqlc"
)

type TransferType string

const (
	TransferTypeTransfer       TransferType = "TRANSFER"
	TransferTypeRequest        TransferType = "REQUEST"
	TransferTypeRedeem         TransferType = "REDEEM"
	TransferTypePayment        TransferType = "PAYMENT"
	TransferTypePaymentVoucher TransferType = "PAYMENT_VOUCHER"
	TransferTypeWithdraw       TransferType = "WITHDRAW"
	TransferTypeDeposit        TransferType = "DEPOSIT"
)

type TransferAction string

const (
	CREATE_CODE TransferAction = "create_code"
	USE_CODE    TransferAction = "use_code"
)

type TransferTxParams struct {
	SenderWalletID   int64                    `json:"sender_wallet_id"`
	ReceiverWalletID pgtype.Int8              `json:"receiver_wallet_id"`
	Charge           pgtype.Int8              `json:"charge"`
	Amount           pgtype.Int8              `json:"amount"`
	Sendamount       pgtype.Int8              `json:"sendamount"`
	Receiveamount    pgtype.Int8              `json:"receiveamount"`
	Note             pgtype.Text              `json:"note"`
	Type             db.NullTransactionType   `json:"type"`
	Status           db.NullTransactionStatus `json:"status"`
	UsedByUsername   []string                 `json:"UsedByUsername"`
	VoucherID        int64                    `json:"id"`
}

type TransferTxResult struct {
	Transaction    db.Transaction
	SenderWallet   db.Wallet
	Receiverwallet db.Wallet
}

//under handletransactions

func (server *Server) TransferTxApi(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {

	var res TransferTxResult

	result, err := server.store.TransferTx(ctx, db.TransactionParams{
		SenderWalletID:   arg.SenderWalletID,
		ReceiverWalletID: arg.ReceiverWalletID,
		Charge:           arg.Charge,
		Amount:           arg.Amount,
		Sendamount:       arg.Sendamount,
		Receiveamount:    arg.Receiveamount,
		Note:             arg.Note,
		Type:             arg.Type,
		Status:           arg.Status,
	})

	if err != nil {
		return TransferTxResult{}, err
	}

	res = TransferTxResult{
		Transaction:    result.Transaction,
		SenderWallet:   result.SenderWallet,
		Receiverwallet: result.Receiverwallet,
	}

	var TransactionStatusSUCCESS = db.NullTransactionStatus{
		TransactionStatus: db.TransactionStatusSUCCESS,
		Valid:             true,
	}

	server.store.UpdateTransactionStatus(ctx, db.UpdateTransactionStatusParams{
		ID:     result.Transaction.ID,
		Status: TransactionStatusSUCCESS,
	})

	return res, err

}

func (server *Server) TransferTxVoucherApi(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {

	var res TransferTxResult

	result, err := server.store.TransferTxForVoucher(ctx, db.TransactionParams{
		SenderWalletID:   arg.SenderWalletID,
		ReceiverWalletID: arg.ReceiverWalletID,
		Charge:           arg.Charge,
		Amount:           arg.Amount,
		Sendamount:       arg.Sendamount,
		Receiveamount:    arg.Receiveamount,
		Note:             arg.Note,
		Type:             arg.Type,
		Status:           arg.Status,
		VoucherID:        arg.VoucherID,
		UsedByUsername:   arg.UsedByUsername,
	})

	if err != nil {
		return TransferTxResult{}, err
	}

	res = TransferTxResult{
		Transaction:    result.Transaction,
		SenderWallet:   result.SenderWallet,
		Receiverwallet: result.Receiverwallet,
	}

	var TransactionStatusSUCCESS = db.NullTransactionStatus{
		TransactionStatus: db.TransactionStatusSUCCESS,
		Valid:             true,
	}

	server.store.UpdateTransactionStatus(ctx, db.UpdateTransactionStatusParams{
		ID:     result.Transaction.ID,
		Status: TransactionStatusSUCCESS,
	})

	return res, err

}

/*
type CreateRedeemTransactionParams struct{

}

type CreateRedeemTransactionResult struct{

}
*/

type RedeemTxResult struct {
	Transaction  db.Transaction
	SenderWallet db.Wallet
	Redeem       db.Redeem
}

// under handletransactions

func (server *Server) CreateRedeemTx(ctx context.Context, arg TransferTxParams, IsSend bool, actions ...TransferAction) (RedeemTxResult, error) {

	var res RedeemTxResult

	for _, action := range actions {
		switch action {
		case CREATE_CODE:

			result, err := server.store.RedeemTx(ctx, db.TransactionParams{
				SenderWalletID: arg.SenderWalletID,
				// ReceiverWalletID: arg.ReceiverWalletID,
				Charge:        arg.Charge,
				Amount:        arg.Amount,
				Sendamount:    arg.Sendamount,
				Receiveamount: arg.Receiveamount,
				Note:          arg.Note,
				Type:          arg.Type,
				Status:        arg.Status,
			}, IsSend)

			if err != nil {
				return res, err
			}

			var TransactionStatusPENDING = db.NullTransactionStatus{
				TransactionStatus: db.TransactionStatusPENDING,
				Valid:             true,
			}

			server.store.UpdateTransactionStatus(ctx, db.UpdateTransactionStatusParams{
				ID:     result.Transaction.ID,
				Status: TransactionStatusPENDING,
			})
			fmt.Printf("result in the handle transaction = %v",result)

			res = RedeemTxResult{
				Transaction:  result.Transaction,
				SenderWallet: result.SenderWallet,
				//Receiverwallet: result.Receiverwallet,
				Redeem: result.Redeem,
			}

			return res, err

		case USE_CODE:

			result, err := server.store.RedeemTx(ctx, db.TransactionParams{
				SenderWalletID:   arg.SenderWalletID,
				ReceiverWalletID: arg.ReceiverWalletID,
				Charge:           arg.Charge,
				Amount:           arg.Amount,
				Sendamount:       arg.Sendamount,
				Receiveamount:    arg.Receiveamount,
				Note:             arg.Note,
				Type:             arg.Type,
				Status:           arg.Status,
			}, IsSend)

			if err != nil {
				return res, err
			}

			res = RedeemTxResult{
				Transaction:  result.Transaction,
				SenderWallet: result.SenderWallet,
				//Receiverwallet: result.Receiverwallet,
				Redeem: result.Redeem,
			}

			var TransactionStatusSUCCESS = db.NullTransactionStatus{
				TransactionStatus: db.TransactionStatusSUCCESS,
				Valid:             true,
			}

			server.store.UpdateTransactionStatus(ctx, db.UpdateTransactionStatusParams{
				ID:     result.Transaction.ID,
				Status: TransactionStatusSUCCESS,
			})

			return res, err
		}
	}

	// Handle the case when no action is provided
	return res, fmt.Errorf("no valid action provided")

}

/*

type TransferTxHandlerResult interface {
	IsTransferTxHandlerResult()
}

func (TransferTxResult) IsTransferTxHandlerResult() {}

func (RedeemTxResult) IsTransferTxHandlerResult() {}

type TransferTypeHandler func(ctx context.Context, arg TransferTxParams, actions ...TransferAction) (TransferTxHandlerResult, error)

func (server *Server) HandleTransaction(ctx context.Context, params TransferTxParams, actions ...TransferAction) (TransferTxHandlerResult, error) {
	typeHandlers := map[TransferType]TransferTypeHandler{
		TransferTypeTransfer:       server.TransferTxApi,
		TransferTypeRequest:        server.TransferTxApi,
		TransferTypeRedeem:         server.CreateRedeemTx,
		TransferTypePayment:        server.TransferTxApi,
		TransferTypePaymentVoucher: server.TransferTxApi,
		TransferTypeWithdraw:       server.TransferTxApi,
		TransferTypeDeposit:        server.TransferTxApi,
	}

	handler, ok := typeHandlers[TransferType(params.Type.TransactionType)]
	if !ok {
		return TransferTxResult{}, fmt.Errorf("unsupported transaction type: %s", params.Type)
	}
	return handler(ctx, params, actions...)

}

*/

func (server *Server) Handletransactions(ctx context.Context, arg TransferTxParams, IsSend bool, actions ...TransferAction) (interface{}, error) {

	switch arg.Type.TransactionType {
	case db.TransactionTypeTRANSFER, db.TransactionTypeREQUEST, db.TransactionTypePAYMENT, db.TransactionTypeWITHDRAW, db.TransactionTypeDEPOSIT:
		// Call TransferTxApi for specified types
		return server.TransferTxApi(ctx, arg)
	case db.TransactionTypeREDEEM:
		// Call CreateRedeemTx for redeem type
		return server.CreateRedeemTx(ctx, arg, IsSend, actions...)
	case db.TransactionTypePAYMENTVOUCHER:
		return server.TransferTxVoucherApi(ctx, arg)
	default:
		return nil, errors.New("unsupported transfer type")
	}
}
