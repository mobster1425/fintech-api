package db

import (
	"context"
	"feyin/digital-fintech-api/util"
	"fmt"
	// "sync"

	//"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)









func TestTransferTx(t *testing.T) {
	wallet1 := createRandomWallet(t)
	wallet2 := createRandomWallet(t)
	fmt.Println(">> before:", wallet1.Balance, wallet2.Balance)
	ChargeForSender := true
	amount := util.RandomMoney()
	fmt.Println(">> Amount:", amount)
	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), ChargeForSender)
	fmt.Println(">> SendAmount:", sendAmount)
	fmt.Println(">> ReceiveAmount:", receiveAmount)
	arg := TransactionParams{
		// SenderWalletID: senderWallet.ID,
		SenderWalletID:   wallet1.ID,
		ReceiverWalletID: pgtype.Int8{Int64: wallet2.ID, Valid: true},
		Amount:           pgtype.Int8{Int64: amount, Valid: true},
		Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
		Type: NullTransactionType{
			TransactionType: TransactionTypeTRANSFER,
			Valid:           true,
		},
		Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
		Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
		Note: pgtype.Text{
			String: "money to you",
			Valid:  true,
		},
		Status: NullTransactionStatus{
			TransactionStatus: TransactionStatusPROCESSING,
			Valid:             true,
		},
	}

	n := 5

	errs := make(chan error)
	results := make(chan TransferTxResult)

	// run n concurrent transfer transaction
	for i := 0; i < n; i++ {
		go func() {
			result, err := testStore.TransferTx(context.Background(), arg)
			errs <- err
			results <- result
		}()
	}

	// check results
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transaction := result.Transaction

		require.Equal(t, wallet1.ID, transaction.SenderWalletID)
		require.Equal(t, wallet2.ID, transaction.ReceiverWalletID.Int64)
		require.Equal(t, arg.Amount.Int64, transaction.Amount.Int64)
		require.Equal(t, arg.Charge.Int64, transaction.Charge.Int64)
		require.Equal(t, arg.Type, transaction.Type)
		require.Equal(t, arg.Sendamount.Int64, transaction.Sendamount.Int64)
		require.Equal(t, arg.Receiveamount.Int64, transaction.Receiveamount.Int64)
		require.Equal(t, arg.Note.String, transaction.Note.String)
		require.Equal(t, arg.Note.Valid, transaction.Note.Valid)
		require.NotEmpty(t, transaction.Status)

		_, err = testStore.GetTransaction(context.Background(), transaction.ID)
		require.NoError(t, err)

		// check fromWallet
		fromWallet := result.SenderWallet
		require.NotEmpty(t, fromWallet)

		require.Equal(t, wallet1.ID, fromWallet.ID)
		//require.Equal(t, wallet1.Balance-int64(amount), fromWallet.Balance) // check balance change
		require.NotZero(t, fromWallet.ID)

		// _, err = testStore.GetTransaction(context.Background(), fromEntry.ID)
		// require.NoError(t, err)

		// check toWallet
		toWallet := result.Receiverwallet
		require.NotEmpty(t, toWallet)

		require.Equal(t, wallet2.ID, toWallet.ID)
		require.NotZero(t, toWallet.ID)
		//require.Equal(t, wallet2.Balance+int64(receiveAmount), toWallet.Balance) // check balance change

		//	_, err = testStore.GetEntry(context.Background(), toEntry.ID)
		//	require.NoError(t, err)

		// check accounts
		/*
			fromAccount := result.FromAccount
			require.NotEmpty(t, fromAccount)
			require.Equal(t, account1.ID, fromAccount.ID)

			toAccount := result.ToAccount
			require.NotEmpty(t, toAccount)
			require.Equal(t, account2.ID, toAccount.ID)

		*/
		// check balances
		fmt.Println(">> tx:", fromWallet.Balance, toWallet.Balance)
		//fmt.Println(">> tx:", fromWallet.Balance, toWallet.Balance)

		diff1 := wallet1.Balance - fromWallet.Balance
		diff2 := toWallet.Balance - wallet2.Balance
		fmt.Println(">> diff:", diff1, diff2)

		require.NotEqual(t, diff1, diff2)
		require.True(t, diff1 > 0)
		// i am here
		require.True(t, diff1%amount != 0) // 1 * amount, 2 * amount, 3 * amount, ..., n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balance
	updatedWallet1, err := testStore.GetWallet(context.Background(), wallet1.ID)
	require.NoError(t, err)

	updatedWallet2, err := testStore.GetWallet(context.Background(), wallet2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedWallet1.Balance, updatedWallet2.Balance)


	
//	require.Equal(t, wallet1.Balance-int64(n)*amount, updatedWallet1.Balance)
//	require.Equal(t, wallet2.Balance+int64(n)*amount, updatedWallet2.Balance)
	

	
	// var chargeFeeValue = 0.02;
	// Calculate the total charge fee for n transfers
totalChargeFee := n * int(chargeFee)

// Adjust the expected balance by subtracting the total charge fee
expectedBalance1 := wallet1.Balance - int64(n)*amount - int64(totalChargeFee)
expectedBalance2 := wallet2.Balance + int64(n)*amount

// Check the final updated balance
require.Equal(t, expectedBalance1, updatedWallet1.Balance)
require.Equal(t, expectedBalance2, updatedWallet2.Balance)

}







/*
func TestTransferTxDeadlock(t *testing.T) {
	wallet1 := createRandomWallet(t)
	wallet2 := createRandomWallet(t)
	fmt.Println(">> before:", wallet1.Balance, wallet2.Balance)

	n := 10
	//amount := int64(10)
	errs := make(chan error)

	ChargeForSender := true
	amount := util.RandomMoney()
	fmt.Println(">> Amount:", amount)
	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), ChargeForSender)
	fmt.Println(">> SendAmount:", sendAmount)
	fmt.Println(">> ReceiveAmount:", receiveAmount)
	arg := TransactionParams{
		// SenderWalletID: senderWallet.ID,
		SenderWalletID:   wallet1.ID,
		ReceiverWalletID: pgtype.Int8{Int64: wallet2.ID, Valid: true},
		Amount:           pgtype.Int8{Int64: amount, Valid: true},
		Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
		Type: NullTransactionType{
			TransactionType: TransactionTypeTRANSFER,
			Valid:           true,
		},
		Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
		Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
		Note: pgtype.Text{
			String: "money to you",
			Valid:  true,
		},
		Status: NullTransactionStatus{
			TransactionStatus: TransactionStatusPROCESSING,
			Valid:             true,
		},
	}

	for i := 0; i < n; i++ {
		// fromWalletID := wallet1.ID
		arg.SenderWalletID = wallet1.ID
		arg.ReceiverWalletID = pgtype.Int8{Int64: wallet2.ID, Valid: true}


		// if u remove this it wouldnt cause a deadlock
		
		if i%2 == 1 {
			arg.SenderWalletID = wallet2.ID
			arg.ReceiverWalletID = pgtype.Int8{Int64: wallet1.ID, Valid: true}
		}
		

		go func() {
			_, err := testStore.TransferTx(context.Background(), arg)

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balance
	updatedWallet1, err := testStore.GetWallet(context.Background(), wallet1.ID)
	require.NoError(t, err)

	updatedWallet2, err := testStore.GetWallet(context.Background(), wallet2.ID)
	require.NoError(t, err)


	// Calculate the total charge fee for n transfers
	totalChargeFee := n * int(chargeFee)
	fmt.Println(">> charge fee:", totalChargeFee)
	// Adjust the expected balance by subtracting the total charge fee
	expectedBalance1 := wallet1.Balance - int64(n)*amount - int64(totalChargeFee)
	expectedBalance2 := wallet2.Balance + int64(n)*amount
	


	fmt.Println(">> after:", updatedWallet1.Balance, updatedWallet2.Balance)
//	require.Equal(t, wallet1.Balance, updatedWallet1.Balance)
//	require.Equal(t, wallet2.Balance, updatedWallet2.Balance)

// Check the final updated balance
require.Equal(t, expectedBalance1, updatedWallet1.Balance)
require.Equal(t, expectedBalance2, updatedWallet2.Balance)

}
*/

func TestRedeemTx(t *testing.T) {
	senderWallet := createRandomWallet(t)
	receiverWallet := createRandomWallet(t)
	fmt.Println(">> before:", senderWallet.Balance, receiverWallet.Balance)

	ChargeForSender := true
	amount := util.RandomMoney()
	fmt.Println(">> Amount:", amount)
	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), ChargeForSender)
	fmt.Println(">> SendAmount:", sendAmount)
	fmt.Println(">> ReceiveAmount:", receiveAmount)
	// Sender Redemption Transaction
	senderArg := TransactionParams{
		SenderWalletID:   senderWallet.ID,
	//	ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
		Amount:           pgtype.Int8{Int64: amount, Valid: true},
		Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
		Type: NullTransactionType{
			TransactionType: TransactionTypeREDEEM,
			Valid:           true,
		},
		Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
		Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
		Note: pgtype.Text{
			String: "redeem money",
			Valid:  true,
		},
		Status: NullTransactionStatus{
			TransactionStatus: TransactionStatusPROCESSING,
			Valid:             true,
		},
	}

	// Receiver Redemption Transaction
	
	receiverArg := TransactionParams{
		SenderWalletID:   senderWallet.ID,
		ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
		Amount:           pgtype.Int8{Int64: amount, Valid: true},
		Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
		Type: NullTransactionType{
			TransactionType: TransactionTypeREDEEM,
			Valid:           true,
		},
		Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
		Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
		Note: pgtype.Text{
			String: "redeem money",
			Valid:  true,
		},
		Status: NullTransactionStatus{
			TransactionStatus: TransactionStatusPROCESSING,
			Valid:             true,
		},
	}

	n := 5

	errs := make(chan error)
	results := make(chan RedeemTxResult)

	errs1 := make(chan error)
	results1 := make(chan RedeemTxResult)

	// run n concurrent sender redeem transactions
	for i := 0; i < n; i++ {
		go func() {
			result, err := testStore.RedeemTx(context.Background(), senderArg, true)
			errs <- err
			results <- result
		}()
	}

	// run n concurrent receiver redeem transactions
	
	for i := 0; i < n; i++ {
		go func() {
			result, err := testStore.RedeemTx(context.Background(), receiverArg, false)
			errs1 <- err
			results1 <- result
		}()
	}

	// check results for sender redemptions
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)
	//	fmt.Println(">> Result:", result)
		
		// check sender redeem
		redeem := result.Redeem
		require.NotEmpty(t, redeem)
		//require.Equal(t, senderArg.SenderWalletID, redeem.Transactionid)
		require.NotEmpty(t, redeem.Code)
		require.NotEmpty(t, redeem.Transactionid)

		_, err = testStore.GetRedeem(context.Background(), redeem.ID)
		require.NoError(t, err)

		// check sender wallet
		senderWalletResult := result.SenderWallet
		require.NotEmpty(t, senderWalletResult)
		require.Equal(t, senderWallet.ID, senderWalletResult.ID)

		// check sender wallet balance change
		_, err = testStore.GetWallet(context.Background(), senderWallet.ID)
		require.NoError(t, err)

	}
	
	// check results for receiver redemptions
	for i := 0; i < n; i++ {
		err := <-errs1
		require.NoError(t, err)

		result := <-results1
	//	fmt.Println(">> Receiver results:", result)
	
		require.Empty(t, result.Redeem)
		require.Empty(t, result.Transaction)
		require.Empty(t, result.SenderWallet)

	}
	

	// check the final updated balances
	updatedSenderWallet, err := testStore.GetWallet(context.Background(), senderWallet.ID)
	require.NoError(t, err)

	updatedReceiverWallet, err := testStore.GetWallet(context.Background(), receiverWallet.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedSenderWallet.Balance, updatedReceiverWallet.Balance)

	/*
	// check sender wallet balance
	require.Equal(t, senderWallet.Balance-int64(n)*amount- int64(chargeFee), updatedSenderWallet.Balance)

	// check receiver wallet balance
	require.Equal(t, receiverWallet.Balance+int64(n)*amount, updatedReceiverWallet.Balance)
*/


	// Calculate the total charge fee for n transfers
	totalChargeFee := n * int(chargeFee)

fmt.Println(">> charge fee:", totalChargeFee)
	// Adjust the expected balance by subtracting the total charge fee
	expectedBalance1 := senderWallet.Balance - int64(n)*amount - int64(totalChargeFee)
	expectedBalance2 := receiverWallet.Balance + int64(n)*amount
	
	// Check the final updated balance
	require.Equal(t, expectedBalance1, updatedSenderWallet.Balance)
	require.Equal(t, expectedBalance2, updatedReceiverWallet.Balance)

}

func TestTransferTxForVoucher(t *testing.T) {
	senderWallet := createRandomWallet(t)
	receiverWallet := createRandomWallet(t)
	voucher := CreateRandomVoucher(t)

	ChargeForSender := true
	amount := util.RandomMoney()

	sendAmount, receiveAmount, chargeFee := util.CalculateAmount(float64(amount), ChargeForSender)
	user1 := util.RandomOwner()
	user2 := util.RandomOwner()

	arg := TransactionParams{
		SenderWalletID:   senderWallet.ID,
		ReceiverWalletID: pgtype.Int8{Int64: receiverWallet.ID, Valid: true},
		Amount:           pgtype.Int8{Int64: amount, Valid: true},
		Charge:           pgtype.Int8{Int64: int64(chargeFee), Valid: true},
		Type: NullTransactionType{
			TransactionType: TransactionTypeTRANSFER,
			Valid:           true,
		},
		Sendamount:    pgtype.Int8{Int64: int64(sendAmount), Valid: true},
		Receiveamount: pgtype.Int8{Int64: int64(receiveAmount), Valid: true},
		Note: pgtype.Text{
			String: "money to you",
			Valid:  true,
		},
		Status: NullTransactionStatus{
			TransactionStatus: TransactionStatusPROCESSING,
			Valid:             true,
		},
		VoucherID:      voucher.ID,
		UsedByUsername: []string{user1, user2}, // Replace with valid usernames
	}

	result, err := testStore.TransferTxForVoucher(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Check transfer transaction
	transaction := result.Transaction
	require.Equal(t, senderWallet.ID, transaction.SenderWalletID)
	require.Equal(t, receiverWallet.ID, transaction.ReceiverWalletID.Int64)
	require.Equal(t, arg.Amount.Int64, transaction.Amount.Int64)
	require.Equal(t, arg.Charge.Int64, transaction.Charge.Int64)
	require.Equal(t, arg.Type, transaction.Type)
	require.Equal(t, arg.Sendamount.Int64, transaction.Sendamount.Int64)
	require.Equal(t, arg.Receiveamount.Int64, transaction.Receiveamount.Int64)
	require.Equal(t, arg.Note.String, transaction.Note.String)
	require.Equal(t, arg.Note.Valid, transaction.Note.Valid)
	require.NotEmpty(t, transaction.Status)

	_, err = testStore.GetTransaction(context.Background(), transaction.ID)
	require.NoError(t, err)

	resultSenderWallet := result.SenderWallet
	resultReiverWallet := result.Receiverwallet

	// Check sender wallet
	updatedSenderWallet, err := testStore.GetWallet(context.Background(), resultSenderWallet.ID)
	require.NoError(t, err)

	require.Equal(t, senderWallet.Balance-int64(amount)- int64(chargeFee), updatedSenderWallet.Balance)

	// Check receiver wallet
	updatedReceiverWallet, err := testStore.GetWallet(context.Background(), resultReiverWallet.ID)
	require.NoError(t, err)
	require.Equal(t, receiverWallet.Balance+int64(receiveAmount), updatedReceiverWallet.Balance)

	// Check voucher update
	updatedVoucher, err := testStore.GetVoucher(context.Background(), voucher.ID)
	require.NoError(t, err)
	require.Equal(t, voucher.ID, updatedVoucher.ID)
	// Add additional checks for voucher update if needed
}
