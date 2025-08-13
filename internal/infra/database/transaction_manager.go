package database

import (
	"errors"
	"gorm.io/gorm"
)

// Transaction represents a database transaction
type Transaction interface{}

// TransactionManager defines the contract for managing database transactions
type TransactionManager interface {
	Begin() (Transaction, error)
	Commit(tx Transaction) error
	Rollback(tx Transaction) error
	Execute(fn func(tx Transaction) error) error
	ExecuteWithResult(fn func(tx Transaction) (interface{}, error)) (interface{}, error)
}

// GormTransactionManager implements TransactionManager for GORM
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager creates a new GORM transaction manager
func NewGormTransactionManager(db *gorm.DB) TransactionManager {
	return &GormTransactionManager{
		db: db,
	}
}

// Begin starts a new transaction
func (m *GormTransactionManager) Begin() (Transaction, error) {
	tx := m.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

// Commit commits the transaction
func (m *GormTransactionManager) Commit(tx Transaction) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return errors.New("invalid transaction type")
	}
	return gormTx.Commit().Error
}

// Rollback rolls back the transaction
func (m *GormTransactionManager) Rollback(tx Transaction) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return errors.New("invalid transaction type")
	}
	return gormTx.Rollback().Error
}

// Execute executes a function within a database transaction
// This is a utility method that handles the transaction lifecycle automatically
func (m *GormTransactionManager) Execute(fn func(tx Transaction) error) error {
	tx, err := m.Begin()
	if err != nil {
		return err
	}

	// Ensure rollback on panic or error
	defer func() {
		if r := recover(); r != nil {
			m.Rollback(tx)
			panic(r) // re-panic after rollback
		}
	}()

	err = fn(tx)
	if err != nil {
		if rollbackErr := m.Rollback(tx); rollbackErr != nil {
			// Log rollback error but return the original error
			return errors.New("transaction failed and rollback failed: " + err.Error() + "; rollback error: " + rollbackErr.Error())
		}
		return err
	}

	return m.Commit(tx)
}

// ExecuteWithResult executes a function within a transaction and returns a result
func (m *GormTransactionManager) ExecuteWithResult(fn func(tx Transaction) (interface{}, error)) (interface{}, error) {
	tx, err := m.Begin()
	if err != nil {
		return nil, err
	}

	// Ensure rollback on panic or error
	defer func() {
		if r := recover(); r != nil {
			m.Rollback(tx)
			panic(r) // re-panic after rollback
		}
	}()

	result, err := fn(tx)
	if err != nil {
		if rollbackErr := m.Rollback(tx); rollbackErr != nil {
			return nil, errors.New("transaction failed and rollback failed: " + err.Error() + "; rollback error: " + rollbackErr.Error())
		}
		return nil, err
	}

	if commitErr := m.Commit(tx); commitErr != nil {
		return nil, commitErr
	}

	return result, nil
}

// GetDB safely casts transaction to *gorm.DB
func GetDB(tx Transaction) *gorm.DB {
	if tx == nil {
		return nil
	}
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return nil
	}
	return gormTx
}
