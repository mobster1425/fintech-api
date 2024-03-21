package db

import "context"


type CreateUserTxParams struct {

CreateUserParams
	AfterCreate func(user User) error
}


type CreateWalletTxParams struct {

	CreateWalletParams
	//	AfterCreate func(user User) error
	}


type CreateUserTxResult struct {
	User User
	Wallet Wallet
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams, arg2 CreateWalletTxParams) (CreateUserTxResult, error) {
	
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		result.Wallet, err = q.CreateWallet(ctx, arg2.CreateWalletParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.User)
	})

	return result, err
	
}