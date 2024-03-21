package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all functions to execute db queries and transactions
type Store interface {
	Querier
	//TransferTx(ctx context.Context, arg TransactionParams, IsSend bool) (TransferTxResult, error)
	TransferTx(ctx context.Context, arg TransactionParams) (TransferTxResult, error)
	RedeemTx(ctx context.Context, arg TransactionParams, IsSend bool) (RedeemTxResult, error) 
//	HandleTransaction(ctx context.Context, params TransferTxParams,  actions ...TransferAction) (TransferTxResult, error) 
	CreateUserTx(ctx context.Context, arg CreateUserTxParams, arg2 CreateWalletTxParams) (CreateUserTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
	TransferTxForVoucher(ctx context.Context, arg TransactionParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

// NewStore creates a new store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}
