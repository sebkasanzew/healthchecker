package health

import (
	"fmt"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/connectivity"
)

type MockGrpcReporter struct {
	state connectivity.State
}

func (m MockGrpcReporter) GetState() connectivity.State {
	return m.state
}

type MockVaultHealthReporter struct {
	health *vault.HealthResponse
	err    error
}

func (m MockVaultHealthReporter) Health() (*vault.HealthResponse, error) {
	return m.health, m.err
}

func TestGrpcProbe(t *testing.T) {
	reporter := &MockGrpcReporter{
		state: connectivity.Ready,
	}

	probe := GrpcProbe(reporter)

	assert.NoError(t, probe())
}

func TestGrpcProbe_err(t *testing.T) {
	reporter := &MockGrpcReporter{
		state: connectivity.Connecting,
	}

	probe := GrpcProbe(reporter)

	assert.Error(t, probe())
}

type MockNatsReporter struct {
	state nats.Status
}

func (m MockNatsReporter) Status() nats.Status {
	return m.state
}

func TestNatsProbe(t *testing.T) {
	reporter := &MockNatsReporter{
		state: nats.CONNECTED,
	}

	probe := NatsProbe(reporter)

	assert.NoError(t, probe())
}

func TestNatsProbe_err(t *testing.T) {
	reporter := &MockNatsReporter{
		state: nats.CONNECTING,
	}

	probe := NatsProbe(reporter)

	assert.Error(t, probe())
}

func TestVaultProbe(t *testing.T) {
	reporter := &MockVaultHealthReporter{
		health: &vault.HealthResponse{
			Initialized: true,
			Sealed:      false,
			Standby:     false,
		},
	}

	probe := VaultProbe(reporter)

	assert.NoError(t, probe())
}

func TestVaultProbe_failsForSealedVault(t *testing.T) {
	reporter := &MockVaultHealthReporter{
		health: &vault.HealthResponse{
			Initialized: true,
			Sealed:      true,
			Standby:     false,
		},
	}

	probe := VaultProbe(reporter)

	assert.Error(t, probe())
}

func TestVaultProbe_failsForNotInitializedVault(t *testing.T) {
	reporter := &MockVaultHealthReporter{
		health: &vault.HealthResponse{
			Initialized: false,
			Sealed:      false,
			Standby:     false,
		},
	}

	probe := VaultProbe(reporter)

	assert.Error(t, probe())
}

func TestVaultProbe_failsForVaultInStandby(t *testing.T) {
	reporter := &MockVaultHealthReporter{
		health: &vault.HealthResponse{
			Initialized: true,
			Sealed:      false,
			Standby:     true,
		},
	}

	probe := VaultProbe(reporter)

	assert.Error(t, probe())
}

func TestVaultProbe_failsForErrorDuringHealthCheck(t *testing.T) {
	reporter := &MockVaultHealthReporter{
		health: nil,
		err:    fmt.Errorf("could not get health"),
	}

	probe := VaultProbe(reporter)

	assert.Error(t, probe())
}
