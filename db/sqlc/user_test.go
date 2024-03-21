package db

import (
	"context"
	"feyin/digital-fintech-api/util"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	arg := CreateUserParams{
		Username:        util.RandomOwner(),
		Password:        hashedPassword,
		Email:           util.RandomEmail(),
		//IsEmailVerified: util.RandomBool(),
	//	Role:            util.RandomRole(),
	Role: NullUserRole{
		UserRole: UserRole(util.RandomRole()),
		Valid: true,
	},
		// Status:          util.RandomStatus(),
		/*
		Status: NullUserStatus{
			UserStatus: UserStatus(util.RandomStatus()),
			Valid: true,
		},*/
	}

	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Email, user.Email)
	require.Empty(t, user.IsEmailVerified)
	//require.Equal(t, arg.IsEmailVerified, user.IsEmailVerified)
	require.Equal(t, arg.Role, user.Role)
	require.NotEmpty(t, user.Status)
	//require.Equal(t, arg.Status, user.Status)
	// require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.Createdat)
	require.NotZero(t, user.Updatedat)

	return user

}

func TestCreateUser(t *testing.T) {
//	wallet := createRandomWallet(t)
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testStore.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Password, user2.Password)
	require.Equal(t, user1.Role, user2.Role)
	require.Equal(t, user1.Status, user2.Status)
	require.Equal(t, user1.IsEmailVerified, user2.IsEmailVerified)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.Updatedat, user2.Updatedat, time.Second)
	require.WithinDuration(t, user1.Createdat, user2.Createdat, time.Second)
}

/*
func TestUpdateUserOnlyUserName(t *testing.T) {
	oldUser := createRandomUser(t)

	newUserName := util.RandomOwner()
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{

		Username: pgtype.Text{
			String: newUserName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.Username, updatedUser.Username)
	//require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.Password, updatedUser.Password)
}
*/

func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)

	newEmail := util.RandomEmail()
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.Password, updatedUser.Password)
}

func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Password: pgtype.Text{
			String: newHashedPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.Password, updatedUser.Password)
	require.Equal(t, newHashedPassword, updatedUser.Password)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.Email, updatedUser.Email)
}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)

	//newUserName := util.RandomOwner()
	newEmail := util.RandomEmail()
	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{

		Username: oldUser.Username,
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
		Password: pgtype.Text{
			String: newHashedPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.Password, updatedUser.Password)
	require.Equal(t, newHashedPassword, updatedUser.Password)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)
	//require.NotEqual(t, oldUser.Username, updatedUser.Username)
	//require.Equal(t, newUserName, updatedUser.Username)
}




// Helper function to create a random user with a specific role

func CreateRandomUserWithRole(t *testing.T, role UserRole) User {
    hashedPassword, err := util.HashPassword(util.RandomString(6))
    require.NoError(t, err)

    arg := CreateUserParams{
        Username:        util.RandomOwner(),
        Password:        hashedPassword,
        Email:           util.RandomEmail(),
      //  IsEmailVerified: util.RandomBool(),
        Role:            NullUserRole{
            UserRole: role,
            Valid:    true,
        },
		
		/*
        Status: NullUserStatus{
            UserStatus: UserStatus(util.RandomStatus()),
            Valid:      true,
        },*/
		
    }

	
	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Email, user.Email)
	require.Empty(t, user.IsEmailVerified)
	//require.Equal(t, arg.IsEmailVerified, user.IsEmailVerified)
	require.Equal(t, arg.Role, user.Role)
	require.NotEmpty(t, user.Status)
	//require.Equal(t, arg.Status, user.Status)
	// require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.Createdat)
	require.NotZero(t, user.Updatedat)

	return user
}



func TestUpdateUserStatus(t *testing.T) {
	oldUser := createRandomUser(t)

	newStatus := NullUserStatus{
		UserStatus: UserStatusActive,
		Valid:      true,
	}

	updateParams := UpdateUserStatusParams{
		Username: oldUser.Username,
		Status:   newStatus,
	}

	updatedUser, err := testStore.UpdateUserStatus(context.Background(), updateParams)
		require.NoError(t, err)
		require.Equal(t, oldUser.Username, updatedUser.Username)
		require.Equal(t, newStatus.UserStatus, updatedUser.Status.UserStatus)
		require.True(t, updatedUser.Status.Valid)

}