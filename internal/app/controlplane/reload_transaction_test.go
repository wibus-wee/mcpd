package controlplane

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"mcpv/internal/domain"
)

func TestReloadTransactionApply_RollbackSuccess(t *testing.T) {
	sequence := make([]string, 0, 4)
	applyErr := errors.New("apply failed")

	steps := []reloadStep{
		{
			name: "step1",
			apply: func(context.Context) error {
				sequence = append(sequence, "apply1")
				return nil
			},
			rollback: func(context.Context) error {
				sequence = append(sequence, "rollback1")
				return nil
			},
		},
		{
			name: "step2",
			apply: func(context.Context) error {
				sequence = append(sequence, "apply2")
				return applyErr
			},
			rollback: func(context.Context) error {
				sequence = append(sequence, "rollback2")
				return nil
			},
		},
	}

	transaction := newReloadTransaction(nil, zap.NewNop())
	err := transaction.apply(context.Background(), steps, domain.ReloadModeLenient)
	require.Error(t, err)

	var applyStageErr reloadApplyError
	require.ErrorAs(t, err, &applyStageErr)
	require.Equal(t, "step2", applyStageErr.stage)

	require.Equal(t, []string{"apply1", "apply2", "rollback1"}, sequence)
}

func TestReloadTransactionApply_RollbackFailure(t *testing.T) {
	applyErr := errors.New("apply failed")
	rollbackErr := errors.New("rollback failed")

	steps := []reloadStep{
		{
			name: "step1",
			apply: func(context.Context) error {
				return nil
			},
			rollback: func(context.Context) error {
				return rollbackErr
			},
		},
		{
			name: "step2",
			apply: func(context.Context) error {
				return applyErr
			},
			rollback: func(context.Context) error {
				return nil
			},
		},
	}

	transaction := newReloadTransaction(nil, zap.NewNop())
	err := transaction.apply(context.Background(), steps, domain.ReloadModeLenient)
	require.Error(t, err)
	require.ErrorIs(t, err, applyErr)
	require.ErrorIs(t, err, rollbackErr)
}
