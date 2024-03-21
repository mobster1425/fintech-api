package db

import (
	"context"
	"feyin/digital-fintech-api/util"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CreateRandomVoucher(t *testing.T) Voucher {
	// creator := createRandomUser(t)
	creator := CreateRandomUserWithRole(t, UserRoleMerchant)
	arg := CreateVoucherParams{
		Value:            50,
		// ApplyforUsername: pgtype.Text{String: util.RandomOwner(), Valid: true},
		ApplyforUsername:  pgtype.Text{String: creator.Username, Valid: true},
		Type:             VoucherType(util.RandomVoucherType()),

		Maxusage:          int32(util.RandomInt(1, 5)),
		Maxusagebyaccount: int32(util.RandomInt(1, 10)),
		Status:            VoucherStatus(VoucherStatusAVAILABLE),

		Expireat:        time.Now().Add(48 * time.Hour),
		Code:            util.GenerateRandomCode(),
		CreatorUsername: pgtype.Text{String: creator.Username, Valid: true},
	}

	voucher, err := testStore.CreateVoucher(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, voucher)

	require.Equal(t, int64(50), voucher.Value)
    require.NotEmpty(t, voucher.ID)
    require.NotZero(t, voucher.Createdat)
    require.NotZero(t, voucher.Updatedat)
    require.NotEmpty(t, voucher.CreatorUsername)
    require.NotEmpty(t, voucher.Type)
    require.NotEmpty(t, voucher.ApplyforUsername)
    require.True(t, voucher.Maxusage >= 1 && voucher.Maxusage <= 5)
    require.True(t, voucher.Maxusagebyaccount >= 1 && voucher.Maxusagebyaccount <= 10)
    require.NotEmpty(t, voucher.Status)
    require.NotZero(t, voucher.Expireat)
    require.Equal(t, arg.Code, voucher.Code)
	require.Equal(t,arg.CreatorUsername, voucher.CreatorUsername)
	require.Equal(t,arg.ApplyforUsername, voucher.ApplyforUsername)
	require.Equal(t,arg.Status, voucher.Status)

	return voucher
}

func TestCreateVoucher(t *testing.T) {
	CreateRandomVoucher(t)
}




// Test the scenario where a non-merchant user attempts to create a voucher
/*
func TestCreateVoucherNonMerchantUser(t *testing.T) {
    // Ensure the creator is not a merchant
    creator := CreateRandomUserWithRole(t, UserRoleCustomer)

    // Attempt to create a voucher with a non-merchant user
    arg := CreateVoucherParams{
        Value:             50,
        ApplyforUsername:  pgtype.Text{String: util.RandomOwner(), Valid: true},
        Type:              VoucherType(util.RandomVoucherType()),
        Maxusage:          int32(util.RandomInt(1, 5)),
        Maxusagebyaccount: int32(util.RandomInt(1, 10)),
        Status:            VoucherStatus(util.RandomVoucherStatus()),
        Expireat:          time.Now().Add(48 * time.Hour),
        Code:              util.GenerateRandomCode(),
        CreatorUsername:   pgtype.Text{String: creator.Username, Valid: true},
    }

    _, err := testStore.CreateVoucher(context.Background(), arg)
    require.Error(t, err) // Expect an error indicating that only merchants can create a voucher
}

*/


func TestGetVoucherWithCode(t *testing.T) {


	voucher1 := CreateRandomVoucher(t)
	arg := GetVoucherWithCodeParams{
		CreatorUsername: voucher1.CreatorUsername,
		Code:            voucher1.Code,
	}


	voucher2, err := testStore.GetVoucherWithCode(context.Background(), arg )
	require.NoError(t, err)
	assertVouchersEqual(t, voucher1, voucher2)
}






func TestGetVoucher(t *testing.T) {
	voucher1 := CreateRandomVoucher(t)
	voucher2, err := testStore.GetVoucher(context.Background(), voucher1.ID)
	require.NoError(t, err)
	assertVouchersEqual(t, voucher1, voucher2)
}


func TestListVouchers(t *testing.T) {
	var createdVouchers []Voucher
	var allVouchers []Voucher // Accumulate vouchers from all iterations

	// Create multiple vouchers for testing list vouchers
	for i := 0; i < 5; i++ {
		voucher := CreateRandomVoucher(t)
		createdVouchers = append(createdVouchers, voucher)
	}

	// List vouchers for each creator and check errors individually
	for i := 0; i < 5; i++ {
		// List vouchers for the creator of the current voucher
		vouchers, err := testStore.ListVouchers(context.Background(), createdVouchers[i].CreatorUsername)
		require.NoError(t, err)

		// Additional assertions based on your requirements
		require.NotEmpty(t, vouchers)
		require.Len(t, vouchers, 1) // Assuming you expect each creator to have only one voucher

		// Accumulate vouchers from this iteration
		allVouchers = append(allVouchers, vouchers...)
	}

	// Perform assertions on all accumulated vouchers
	require.Len(t, allVouchers, 5)

	// Add any additional assertions based on your requirements
	for _, voucher := range allVouchers {
		require.NotEmpty(t, voucher)
		
	}
}



func TestUpdateVoucherStatus(t *testing.T) {
    voucher := CreateRandomVoucher(t)

    updatedVoucher, err := testStore.UpdateVoucherStatus(context.Background(), UpdateVoucherStatusParams{
        ID:     voucher.ID,
        Status: VoucherStatusUNAVAILABLE,
    })
    require.NoError(t, err)

    // Update the expected status to VoucherStatusUNAVAILABLE
    expectedStatus := VoucherStatusUNAVAILABLE
    assert.Equal(t, expectedStatus, updatedVoucher.Status)

    // Or if you have a custom function assertVouchersEqual, make sure it checks the status correctly
    assertVouchersEqual(t, voucher, updatedVoucher)
}


func TestDeleteVoucher(t *testing.T) {
	voucher := CreateRandomVoucher(t)

	err := testStore.DeleteVoucher(context.Background(), voucher.ID)
	require.NoError(t, err)

	// Verify that the voucher is deleted
	_, err = testStore.GetVoucher(context.Background(), voucher.ID)
	assert.Error(t, err)
	assert.EqualError(t, err, "no rows in result set")
}








func TestUpdateVoucherUsedBy(t *testing.T) {
	voucher := CreateRandomVoucher(t)
	user := createRandomUser(t)
	
	

// Define parameters for updating the voucher used by
arg := UpdateVoucherUsedByParams{
	ID:     voucher.ID,
	Column2: []string{user.Username},
}

	 err := testStore.UpdateVoucherUsedBy(context.Background(),arg)
		require.NoError(t, err)
		// assert.Equal(t, VoucherStatusUNAVAILABLE, updatedVoucher.Status)
	
		//assertVouchersEqual(t,voucher,updatedVoucher)
	}





func assertVouchersEqual(t *testing.T, expected, actual Voucher) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Value, actual.Value)
	assert.Equal(t, expected.ApplyforUsername, actual.ApplyforUsername)
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Maxusage, actual.Maxusage)
	assert.Equal(t, expected.Maxusagebyaccount, actual.Maxusagebyaccount)
	//assert.NotEqual(t, expected.Status, actual.Status)
	assert.Equal(t, expected.Expireat, actual.Expireat)
	assert.Equal(t, expected.Code, actual.Code)
	assert.Equal(t, expected.CreatorUsername, actual.CreatorUsername)
}
