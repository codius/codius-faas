package function

import (
	"fmt"
	"os"
	"testing"
)

func Test_parseUint64Value_whenNotSet(t *testing.T) {
	_, err := parseUint64Value("missing_value")

	if err == nil {
		t.Error("Expected error")
	}
}

func Test_parseUint64Value_whenSet(t *testing.T) {
	expected := uint64(1000)
	name := "test_value"

	if err := os.Setenv(name, fmt.Sprintf("%d", expected)); err != nil {
		t.Error(err)
	}

	got, err := parseUint64Value(name)

	if err != nil {
		t.Error(err)
	}

	if expected != got {
		t.Errorf("Expected: %d, got: %d", expected, got)
	}
}

func Test_parseUint64Value_whenSetInvalid(t *testing.T) {
	name := "test_value"

	if err := os.Setenv(name, "NaN"); err != nil {
		t.Error(err)
	}

	_, err := parseUint64Value(name)

	if err == nil {
		t.Error("Expected error")
	}
}

func Test_calculateInvocationsCost_whenExact(t *testing.T) {
	invocations := uint64(100)
	costPerUnitInvocations := uint64(10)
	unitInvocations := uint64(1000)

	expected := uint64(1)

	got := calculateInvocationsCost(invocations, costPerUnitInvocations, unitInvocations)

	if expected != got {
		t.Errorf("Expected: %d, got: %d", expected, got)
	}
}

func Test_calculateInvocationsCost_whenRounded(t *testing.T) {
	invocations := uint64(199)
	costPerUnitInvocations := uint64(10)
	unitInvocations := uint64(1000)

	expected := uint64(1)

	got := calculateInvocationsCost(invocations, costPerUnitInvocations, unitInvocations)

	if expected != got {
		t.Errorf("Expected: %d, got: %d", expected, got)
	}
}

func Test_NewFunctionBalance_whenBalanceZero(t *testing.T) {
	credit := uint64(0)
	invocations := uint64(5)
	costPerUnitInvocations := uint64(10)
	unitInvocations := uint64(100)
	bonusInvocations := uint64(50)

	expected := FunctionBalance{
		Balance:     0,
		Invocations: 45,
	}

	got := NewFunctionBalance(credit, invocations, costPerUnitInvocations, unitInvocations, bonusInvocations)

	if expected != got {
		t.Errorf("Expected: %d, got: %d", expected, got)
	}
}

func Test_NewFunctionBalance_whenBalancePositive(t *testing.T) {
	credit := uint64(100)
	invocations := uint64(100)
	costPerUnitInvocations := uint64(10)
	unitInvocations := uint64(1000)
	bonusInvocations := uint64(50)

	expected := FunctionBalance{
		Balance:     99,
		Invocations: 9850,
	}

	got := NewFunctionBalance(credit, invocations, costPerUnitInvocations, unitInvocations, bonusInvocations)

	if expected != got {
		t.Errorf("Expected: %d, got: %d", expected, got)
	}
}

func Test_NewFunctionBalance_whenBalanceNegative(t *testing.T) {
	credit := uint64(9)
	invocations := uint64(100)
	costPerUnitInvocations := uint64(10)
	unitInvocations := uint64(100)
	bonusInvocations := uint64(50)

	expected := FunctionBalance{
		Balance:     0,
		Invocations: 40,
	}

	got := NewFunctionBalance(credit, invocations, costPerUnitInvocations, unitInvocations, bonusInvocations)

	if expected != got {
		t.Errorf("Expected: %d, got: %d", expected, got)
	}
}

func Test_NewFunctionBalance_whenRemainingInvocationsNegative(t *testing.T) {
	credit := uint64(4)
	invocations := uint64(100)
	costPerUnitInvocations := uint64(10)
	unitInvocations := uint64(100)
	bonusInvocations := uint64(50)

	expected := FunctionBalance{
		Balance:     0,
		Invocations: 0,
	}

	got := NewFunctionBalance(credit, invocations, costPerUnitInvocations, unitInvocations, bonusInvocations)

	if expected != got {
		t.Errorf("Expected: %d, got: %d", expected, got)
	}
}
