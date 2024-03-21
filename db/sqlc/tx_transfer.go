package db

import (
	"context"
	//"strings"
	
	"feyin/digital-fintech-api/util"

	"github.com/jackc/pgx/v5/pgtype"
)

type TransferTxResult struct {
	Transaction    Transaction
	SenderWallet   Wallet
	Receiverwallet Wallet
}

type TransactionParams struct {
	SenderWalletID   int64                 `json:"sender_wallet_id"`
	ReceiverWalletID pgtype.Int8           `json:"receiver_wallet_id"`
	Charge           pgtype.Int8           `json:"charge"`
	Amount           pgtype.Int8           `json:"amount"`
	Sendamount       pgtype.Int8           `json:"sendamount"`
	Receiveamount    pgtype.Int8           `json:"receiveamount"`
	Note             pgtype.Text           `json:"note"`
	Type             NullTransactionType   `json:"type"`
	Status           NullTransactionStatus `json:"status"`
	UsedByUsername          []string `json:"UsedByUsername"`
	VoucherID                int64         `json:"id"`
}

func (store *SQLStore) TransferTx(ctx context.Context, arg TransactionParams) (TransferTxResult, error) {

	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transaction, err = q.CreateTransaction(ctx, CreateTransactionParams{
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
			return err
		}

		if arg.SenderWalletID < arg.ReceiverWalletID.Int64 {
			result.SenderWallet, result.Receiverwallet, err = addMoney(ctx, q, arg.SenderWalletID, -arg.Sendamount.Int64, arg.ReceiverWalletID.Int64, arg.Receiveamount.Int64)
		} else {
			//	result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			result.SenderWallet, result.Receiverwallet, err = addMoney(ctx, q, arg.ReceiverWalletID.Int64, arg.Receiveamount.Int64, arg.SenderWalletID, -arg.Sendamount.Int64)
		}

		return err

	})

	return result, err

}

func addMoney(
	ctx context.Context,
	q *Queries,
	walletID1 int64,
	sendAmount int64,
	walletID2 int64,
	ReceiveAmount int64,
) (SenderWallet Wallet, ReceiverWallet Wallet, err error) {
	SenderWallet, err = q.AddWalletBalance(ctx, AddWalletBalanceParams{
		ID:     walletID1,
		Amount: sendAmount,
	})
	if err != nil {
		return
	}

	ReceiverWallet, err = q.AddWalletBalance(ctx, AddWalletBalanceParams{
		ID:     walletID2,
		Amount: ReceiveAmount,
	})
	return
}

type RedeemTxResult struct {
	Transaction  Transaction
	SenderWallet Wallet
	Redeem       Redeem
}

func (store *SQLStore) RedeemTx(ctx context.Context, arg TransactionParams, IsSend bool) (RedeemTxResult, error) {
	var result RedeemTxResult
	code := util.GenerateRandomCode()

	if IsSend {
		err := store.execTx(ctx, func(q *Queries) error {
			var err error

			result.Transaction, err = q.CreateTransaction(ctx, CreateTransactionParams{
				SenderWalletID: arg.SenderWalletID,
				// ReceiverWalletID: arg.ReceiverWalletID,
				Charge:        arg.Charge,
				Amount:        arg.Amount,
				Sendamount:    arg.Sendamount,
				Receiveamount: arg.Receiveamount,
				Note:          arg.Note,
				Type:          arg.Type,
				Status:        arg.Status,
			})
			if err != nil {
				return err
			}

			result.Redeem, err = q.CreateRedeem(ctx, CreateRedeemParams{
				Transactionid: result.Transaction.ID,
				Code:          code,
			})

			if err != nil {
				return err
			}

			result.SenderWallet, err = q.AddWalletBalance(ctx, AddWalletBalanceParams{
				ID: arg.SenderWalletID,
				//Amount: -arg.Sendamount,
				Amount: -arg.Sendamount.Int64,
			})
			if err != nil {
				return err
			}

			return err
		})

		return result, err
	} else {
		err := store.execTx(ctx, func(q *Queries) error {
			var err error

      //  result.Redeem,err= q.GetRedeemWithCode(ctx,arg.)


			_, err = q.AddWalletBalance(ctx, AddWalletBalanceParams{
				ID:     arg.ReceiverWalletID.Int64,
				Amount: arg.Receiveamount.Int64,
			})
			if err != nil {
				return err
			}

			return err
		}) //end of 2nd transaction

		return RedeemTxResult{}, err
	} //end of else
}




func (store *SQLStore) TransferTxForVoucher(ctx context.Context, arg TransactionParams) (TransferTxResult, error) {

	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transaction, err = q.CreateTransaction(ctx, CreateTransactionParams{
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
			return err
		}



	  err= q.UpdateVoucherUsedBy(ctx,UpdateVoucherUsedByParams{
			ID: arg.VoucherID,
			Column2: arg.UsedByUsername,
		})
		if err != nil {
			return err
		}
		
		if arg.SenderWalletID < arg.ReceiverWalletID.Int64 {
			result.SenderWallet, result.Receiverwallet, err = addMoney(ctx, q, arg.SenderWalletID, -arg.Sendamount.Int64, arg.ReceiverWalletID.Int64, arg.Receiveamount.Int64)
		} else {
			//	result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			result.SenderWallet, result.Receiverwallet, err = addMoney(ctx, q, arg.ReceiverWalletID.Int64, arg.Receiveamount.Int64, arg.SenderWalletID, -arg.Sendamount.Int64)
		}

		return err

	})

	return result, err

}
