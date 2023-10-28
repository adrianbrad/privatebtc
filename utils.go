package privatebtc

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func generateRPCPasswordSalt(size uint8) (string, error) {
	salt := make([]byte, size)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	return hex.EncodeToString(salt), nil
}

func rpcPasswordToHMAC(salt, password string) string {
	h := hmac.New(sha256.New, []byte(salt))

	// sha256.Write() never returns an error
	h.Write([]byte(password))

	return hex.EncodeToString(h.Sum(nil))
}

func newRPCAuth(user, password string) (string, error) {
	const saltSize = 16

	salt, err := generateRPCPasswordSalt(saltSize)
	if err != nil {
		return "", fmt.Errorf("new salt: %w", err)
	}

	hashedPassword := rpcPasswordToHMAC(salt, password)

	return fmt.Sprintf("%s:%s$%s", user, salt, hashedPassword), nil
}

// ReplaceTransactionDrainToAddress performs a replace by fee(RBF) for the given transaction id.
// The new transaction will drain the inputs of the old transaction to the given address while
// increasing the fee by 100%.
func ReplaceTransactionDrainToAddress(
	ctx context.Context,
	client RPCClient,
	txID string,
	address string,
) (string, error) {
	tx, err := client.GetTransaction(ctx, txID)
	if err != nil {
		return "", fmt.Errorf("get transaction: %w", err)
	}

	totalInputs, err := tx.TotalInputsValue(ctx, client)
	if err != nil {
		return "", fmt.Errorf("total inputs value: %w", err)
	}

	fee := tx.GetTransactionFee(totalInputs)

	newFee := 2 * fee
	newAmount := totalInputs - newFee

	hash, err := client.SendCustomTransaction(ctx, tx.Vin, map[string]float64{address: newAmount})
	if err != nil {
		return "", fmt.Errorf("send custom transaction: %w", err)
	}

	return hash, nil
}
