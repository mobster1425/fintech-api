package db

import (
	"context"
	"feyin/digital-fintech-api/util"
	"testing"

	//	"time"

	//	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateRandomRedeem generates random data for testing CreateRedeem function.
func CreateRandomRedeem(t *testing.T) Redeem {
	code := util.GenerateRandomCode()

	wallet1 := createRandomWallet(t)
	//wallet2:=createRandomWallet(t)
	transaction := CreateRandomTransactionForRedeem(t, wallet1)

	arg := CreateRedeemParams{
		Code:          code,
		Transactionid: transaction.ID,
	}
	// transactionID := int64(util.RandomInt(1, 100))
	createdRedeem, err := testStore.CreateRedeem(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, createdRedeem)
	// assertRedeemsEqual(t, arg, createdRedeem)

	require.NotEmpty(t, createdRedeem.ID)
	assert.Equal(t, arg.Code, createdRedeem.Code)
	assert.Equal(t, arg.Transactionid, createdRedeem.Transactionid)

	return createdRedeem

}

func assertRedeemsEqual(t *testing.T, expected, actual Redeem) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Code, actual.Code)
	assert.Equal(t, expected.Transactionid, actual.Transactionid)
	// Add additional assertions based on your requirements
}

func TestCreateRedeem(t *testing.T) {
	CreateRandomRedeem(t)

}

func TestGetRedeem(t *testing.T) {
	redeem := CreateRandomRedeem(t)

	fetchedRedeem, err := testStore.GetRedeem(context.Background(), redeem.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedRedeem)

	assertRedeemsEqual(t, redeem, fetchedRedeem)
}

func TestDeleteRedeem(t *testing.T) {
	redeem := CreateRandomRedeem(t)

	err := testStore.DeleteRedeem(context.Background(), redeem.ID)
	require.NoError(t, err)

	// Attempt to fetch the deleted redeem, should return an empty redeem and no error
	deletedRedeem, err := testStore.GetRedeem(context.Background(), redeem.ID)
	require.Error(t, err)
	require.Empty(t, deletedRedeem)
}


func TestGetRedeemWithCode(t *testing.T) {
	redeem := CreateRandomRedeem(t)

   	// Call the GetRedeemWithCode function to retrieve the redeem by its code
		retrievedRedeem, err := testStore.GetRedeemWithCode(context.Background(), redeem.Code)
		require.NoError(t, err)

		// Assert that the retrieved redeem matches the original redeem
		assertRedeemsEqual(t, redeem, retrievedRedeem)


}