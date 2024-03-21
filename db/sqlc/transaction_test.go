package db

import (
	"context"
	"feyin/digital-fintech-api/util"
	"testing"
	"time"

	 "github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomTransaction(t *testing.T, wallet1,wallet2 Wallet) Transaction {
//	senderWallet := createRandomWallet(t)
	// receiverWallet := createRandomWallet(t)

  sendamount1:= util.RandomMoney()
charge1:=50
	arg := CreateTransactionParams{
		SenderWalletID:   wallet1.ID,
		ReceiverWalletID: pgtype.Int8{Int64: wallet2.ID, Valid: true},
		Amount:           pgtype.Int8{Int64: sendamount1, Valid: true}, 
		Charge:           pgtype.Int8{Int64: int64(charge1), Valid: true},   // Set the charge as needed
		// Type:             util.RandomTransactionType(),   
		Type: NullTransactionType{
			TransactionType: TransactionType(util.RandomTransactionType()),
			Valid:           true,
		},
		Note: pgtype.Text{
			String: "not now",
       Valid: true,
		},
		Status: NullTransactionStatus{
			TransactionStatus: TransactionStatus(util.RandomTransactionStatus()),
			Valid: true,
		},
										 // Set the type as needed
		Sendamount:       pgtype.Int8{Int64: sendamount1-int64(charge1), Valid: true},  // Set the send amount as needed
		Receiveamount:    pgtype.Int8{Int64: sendamount1-int64(charge1), Valid: true},  // Set the receive amount as needed
	}

	transaction, err := testStore.CreateTransaction(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transaction)

	require.Equal(t, wallet1.ID, transaction.SenderWalletID)
	require.Equal(t, wallet2.ID, transaction.ReceiverWalletID.Int64)
	require.Equal(t, arg.Amount.Int64, transaction.Amount.Int64)
	require.Equal(t, arg.Charge.Int64, transaction.Charge.Int64)
	require.Equal(t, arg.Type, transaction.Type)
	require.Equal(t, arg.Sendamount.Int64, transaction.Sendamount.Int64)
	require.Equal(t, arg.Receiveamount.Int64, transaction.Receiveamount.Int64)
	require.Equal(t, arg.Note.String, transaction.Note.String)
	require.Equal(t, arg.Note.Valid, transaction.Note.Valid)
//	require.Empty(t,transaction.Note)

require.True(t, transaction.Status.Valid)
	// Add more assertions based on your requirements

	return transaction
}

func TestCreateTransaction(t *testing.T) {
	wallet1:=createRandomWallet(t)
	wallet2:=createRandomWallet(t)
	createRandomTransaction(t,wallet1,wallet2)
}



func TestGetTransaction(t *testing.T) {
	wallet1:=createRandomWallet(t)
	wallet2:=createRandomWallet(t)

	transaction1 := createRandomTransaction(t,wallet1,wallet2)
	transaction2, err := testStore.GetTransaction(context.Background(), transaction1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transaction2)

	// require.Equal(t, transaction1.SenderWalletID, transaction2.SenderWalletID)
	// Add more assertions based on your requirements
	require.Equal(t, transaction1.ID, transaction2.ID)
	require.WithinDuration(t, transaction1.Createdat, transaction2.Createdat, time.Second)
	require.WithinDuration(t, transaction1.Updatedat, transaction2.Updatedat, time.Second)
	require.Equal(t, transaction1.SenderWalletID, transaction2.SenderWalletID)
	require.Equal(t, transaction1.ReceiverWalletID.Int64, transaction2.ReceiverWalletID.Int64)
	require.Equal(t, transaction1.Charge.Int64, transaction2.Charge.Int64)
	require.Equal(t, transaction1.Amount.Int64, transaction2.Amount.Int64)
	require.Equal(t, transaction1.Sendamount.Int64, transaction2.Sendamount.Int64)
	require.Equal(t, transaction1.Receiveamount.Int64, transaction2.Receiveamount.Int64)
		// require.Equal(t, transaction1.Note, transaction2.Note)
	 require.Equal(t, transaction1.Note.Valid, transaction2.Note.Valid)
	// require.False(t, transaction1.Note.Valid)
   //	require.False(t,transaction2.Note.Valid)
	require.Equal(t, transaction1.Type.TransactionType, transaction2.Type.TransactionType)
	require.Equal(t, transaction1.Type.Valid, transaction2.Type.Valid)
//	require.False(t, transaction1.Status.Valid)
//	require.False(t,transaction2.Status.Valid)
	// require.Equal(t, transaction1.Status.TransactionStatus, transaction2.Status.TransactionStatus)
    require.Equal(t, transaction1.Status.Valid, transaction2.Status.Valid)
}


func TestListTransactions(t *testing.T) {
	senderWallet := createRandomWallet(t)
	receiverWallet := createRandomWallet(t)

	// Create multiple transactions for the same sender wallet for testing list transactions
	for i := 0; i < 5; i++ {
		createRandomTransaction(t, senderWallet, receiverWallet)
		createRandomTransaction(t, receiverWallet, senderWallet)
	}

	arg := ListTransactionsParams{
		SenderWalletID:   senderWallet.ID,
		ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
		Limit:            5,
		Offset:           0, // Set Offset to 0 instead of 5
	}

	transactions, err := testStore.ListTransactions(context.Background(), arg)
	require.NoError(t, err)

	// Adjust the expected length to 10 if you expect both sender and receiver transactions
	require.Len(t, transactions, 5)

	for _, transaction := range transactions {
		require.NotEmpty(t, transaction)
		require.True(t, transaction.SenderWalletID == senderWallet.ID || transaction.ReceiverWalletID.Int64 == senderWallet.ID)
		// Additional assertions based on your requirements
	}
}






func TestUpdateTransactionStatus(t *testing.T) {
	wallet1:=createRandomWallet(t)
	wallet2:=createRandomWallet(t)
	transaction := createRandomTransaction(t,wallet1,wallet2)

//	newStatus := "COMPLETED" // Update the status as needed
	updatedTransaction, err := testStore.UpdateTransactionStatus(context.Background(), UpdateTransactionStatusParams{
		ID:     transaction.ID,
		// Status: newStatus,
		Status: NullTransactionStatus{
			TransactionStatus: TransactionStatus(util.RandomTransactionStatus()),
			Valid: true,
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedTransaction)

	require.Equal(t, transaction.ID, updatedTransaction.ID)
	require.NotEqual(t, transaction.Status, updatedTransaction.Status)

	// Add more assertions based on your requirements
	require.Equal(t, transaction.ID, updatedTransaction.ID)
	require.Equal(t, transaction.SenderWalletID, updatedTransaction.SenderWalletID)
	require.Equal(t, transaction.ReceiverWalletID.Int64, updatedTransaction.ReceiverWalletID.Int64)
	require.Equal(t, transaction.Charge.Int64, updatedTransaction.Charge.Int64)
	require.Equal(t, transaction.Amount.Int64, updatedTransaction.Amount.Int64)
	require.Equal(t, transaction.Sendamount.Int64, updatedTransaction.Sendamount.Int64)
	require.Equal(t, transaction.Receiveamount.Int64, updatedTransaction.Receiveamount.Int64)
	require.Equal(t, transaction.Note.Valid, updatedTransaction.Note.Valid)
}







func CreateRandomTransactionForRedeem(t *testing.T, wallet1 Wallet) Transaction {
	//	senderWallet := createRandomWallet(t)
		// receiverWallet := createRandomWallet(t)
	
	  sendamount1:= util.RandomMoney()
	charge1:=50
		arg := CreateTransactionParams{
			SenderWalletID:   wallet1.ID,
			//ReceiverWalletID: pgtype.Int8{Int64: wallet2.ID, Valid: true},
			Amount:           pgtype.Int8{Int64: sendamount1, Valid: true}, 
			Charge:           pgtype.Int8{Int64: int64(charge1), Valid: true},   // Set the charge as needed
			// Type:             util.RandomTransactionType(),   
			Type: NullTransactionType{
				TransactionType: TransactionType(TransactionTypeREDEEM),
				Valid:           true,
			},
			Note: pgtype.Text{
				String: "not now",
		   Valid: false,
			},
			Status: NullTransactionStatus{
				TransactionStatus: TransactionStatus(TransactionStatusPENDING),
				Valid: true,
			},
											 // Set the type as needed
			Sendamount:       pgtype.Int8{Int64: sendamount1-int64(charge1), Valid: true},  // Set the send amount as needed
			Receiveamount:    pgtype.Int8{Int64: sendamount1-int64(charge1), Valid: true},  // Set the receive amount as needed
		}
	
		transaction, err := testStore.CreateTransaction(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, transaction)
	
		require.Equal(t, wallet1.ID, transaction.SenderWalletID)
	//	require.Equal(t, wallet2.ID, transaction.ReceiverWalletID.Int64)
		require.Equal(t, arg.Amount.Int64, transaction.Amount.Int64)
		require.Equal(t, arg.Charge.Int64, transaction.Charge.Int64)
		require.Equal(t, arg.Type, transaction.Type)
		require.Equal(t, arg.Sendamount.Int64, transaction.Sendamount.Int64)
		require.Equal(t, arg.Receiveamount.Int64, transaction.Receiveamount.Int64)
	//	require.Empty(t,transaction.Note)
	//require.Equal(t, arg.Note.String, transaction.Note.String)
	//require.Equal(t, arg.Note.Valid, transaction.Note.Valid)
		// Add more assertions based on your requirements
	
		return transaction
	}