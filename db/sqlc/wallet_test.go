package db

import (
	"context"
	"feyin/digital-fintech-api/util"
	"testing"
	"time"

	// "github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomWallet(t *testing.T) Wallet {
	randomUser := createRandomUser(t)
	arg := CreateWalletParams{
		Owner:   randomUser.Username,
	//Owner: util.RandomOwner(),
		Balance: 0, // Set the initial balance as needed
	}

	wallet, err := testStore.CreateWallet(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wallet)

	require.Equal(t, arg.Owner, wallet.Owner)
	require.Equal(t, arg.Balance, wallet.Balance)
	require.NotZero(t, wallet.Createdat)
	require.NotZero(t, wallet.Updatedat)

	return wallet
}

func TestCreateWallet(t *testing.T) {
	createRandomWallet(t)
}

func TestGetWallet(t *testing.T) {
	wallet1 := createRandomWallet(t)
	wallet2, err := testStore.GetWallet(context.Background(), wallet1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, wallet2)

	require.Equal(t, wallet1.Owner, wallet2.Owner)
	require.Equal(t, wallet1.Balance, wallet2.Balance)
	require.WithinDuration(t, wallet1.Updatedat, wallet2.Updatedat, time.Second)
	require.WithinDuration(t, wallet1.Createdat, wallet2.Createdat, time.Second)
}

func TestUpdateWallet(t *testing.T) {
	wallet := createRandomWallet(t)

	newBalance := wallet.Balance + util.RandomMoney() // Update the balance as needed
	updatedWallet, err := testStore.UpdateWallet(context.Background(), UpdateWalletParams{
		ID:      wallet.ID,
		Balance: newBalance,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedWallet)

	require.Equal(t, wallet.ID, updatedWallet.ID)
	require.Equal(t, wallet.Owner, updatedWallet.Owner)
	require.Equal(t, newBalance, updatedWallet.Balance)
	require.WithinDuration(t, wallet.Updatedat, updatedWallet.Updatedat, time.Second)
	require.WithinDuration(t, wallet.Createdat, updatedWallet.Createdat, time.Second)
}

func TestAddWalletBalance(t *testing.T) {
	wallet := createRandomWallet(t)

	addedAmount := util.RandomMoney() // Add the desired amount
	updatedWallet, err := testStore.AddWalletBalance(context.Background(), AddWalletBalanceParams{
		ID:     wallet.ID,
		Amount: addedAmount,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedWallet)

	require.Equal(t, wallet.ID, updatedWallet.ID)
	require.Equal(t, wallet.Owner, updatedWallet.Owner)
	require.Equal(t, wallet.Balance+addedAmount, updatedWallet.Balance)
	require.WithinDuration(t, wallet.Updatedat, updatedWallet.Updatedat, time.Second)
	require.WithinDuration(t, wallet.Createdat, updatedWallet.Createdat, time.Second)
}

// write test for get walletbyowner

func TestGetWalletByOwner(t *testing.T) {
	wallet1 := createRandomWallet(t)
	wallet2, err := testStore.GetWalletbyOwner(context.Background(), wallet1.Owner)
	require.NoError(t, err)
	require.NotEmpty(t, wallet2)

	require.Equal(t, wallet1.Owner, wallet2.Owner)
	require.Equal(t, wallet1.Balance, wallet2.Balance)
	require.WithinDuration(t, wallet1.Updatedat, wallet2.Updatedat, time.Second)
	require.WithinDuration(t, wallet1.Createdat, wallet2.Createdat, time.Second)
}
