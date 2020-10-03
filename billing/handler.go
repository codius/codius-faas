package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	metrics "github.com/openfaas/openfaas-cloud/metrics"
)

type FunctionBalance struct {
	Balance     uint64 `json:"balance"`
	Invocations uint64 `json:"remainingInvocations"`
}

func NewFunctionBalance(credit, invocations, costPerUnitInvocations, unitInvocations, bonusInvocations uint64) FunctionBalance {
	fnBalance := FunctionBalance{}

	totalCharge := calculateInvocationsCost(invocations, costPerUnitInvocations, unitInvocations)
	if credit < totalCharge {
		fnBalance.Balance = 0
		remainingInvocations := calculateRemainingInvocations(totalCharge-credit, costPerUnitInvocations, unitInvocations)
		if bonusInvocations < remainingInvocations {
			fnBalance.Invocations = 0
		} else {
			fnBalance.Invocations = bonusInvocations - remainingInvocations
		}
	} else {
		fnBalance.Balance = credit - totalCharge
		fnBalance.Invocations = bonusInvocations + calculateRemainingInvocations(fnBalance.Balance, costPerUnitInvocations, unitInvocations)
	}

	return fnBalance
}

// Handle exposes the OpenFaaS instance metrics
func Handle(req []byte) string {

	fnName, err := parseFunctionName()
	if err != nil {
		log.Fatalf("couldn't parse function name from query: %t", err)
	}

	costPerUnitInvocations, err := parseUint64Value("cost_per_unit_invocations")
	if err != nil {
		log.Fatal(err)
	}

	unitInvocations, err := parseUint64Value("unit_invocations")
	if err != nil {
		log.Fatal(err)
	}

	bonusInvocations, err := parseUint64Value("bonus_invocations")
	if err != nil {
		log.Fatal(err)
	}

	balancesUrl := os.Getenv("balances_url")
	credit, err := getFunctionBalance(fnName, balancesUrl)
	if err != nil {
		log.Fatalf("Couldn't get balance for function %s, %t", fnName, err)
	}

	gatewayURL := os.Getenv("gateway_url")
	invocations, err := getFunctionInvocations(fnName, gatewayURL)
	if err != nil {
		log.Fatalf("Couldn't get invocations for function %s, %t", fnName, err)
	}

	fnBalance := NewFunctionBalance(credit, invocations, costPerUnitInvocations, unitInvocations, bonusInvocations)
	res, err := json.Marshal(fnBalance)
	if err != nil {
		log.Fatalf("Couldn't marshal json %t", err)
	}
	return string(res)
}

func parseFunctionName() (functionName string, error error) {
	if query, exists := os.LookupEnv("Http_Query"); exists {
		vals, _ := url.ParseQuery(query)

		functionNameQuery := vals.Get("function")

		if len(functionNameQuery) > 0 {
			return functionNameQuery, nil
		}

		return "", fmt.Errorf("there is no `function` inside env var Http_Query")
	}

	return "", fmt.Errorf("unable to parse Http_Query")
}

func parseUint64Value(name string) (val uint64, err error) {
	if valStr, exists := os.LookupEnv(name); exists && len(valStr) > 0 {
		val, err = strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("Couldn't parse `%s` to uint64. Value: %s, Error: %t", name, valStr, err)
		}
	} else {
		return 0, fmt.Errorf("`%s` must be set", name)
	}
	return val, nil
}

func getFunctionBalance(function, balancesUrl string) (uint64, error) {
	if response, err := http.Get(balancesUrl + "/balances/" + function); err != nil {
		return 0, err
	} else if response.Body == nil {
		return 0, fmt.Errorf("no balance for function: %s", function)
	} else {
		defer response.Body.Close()

		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return 0, err
		}
		// fmt.Println(string(bodyBytes))
		balance, err := strconv.ParseUint(string(bodyBytes), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("Couldn't parse balance to uint64. Value: %s, Error: %t", string(bodyBytes), err)
		}
		return balance, nil
	}
	return 100, nil
}

func getFunctionInvocations(function, gatewayURL string) (uint64, error) {
	if response, err := http.Get(gatewayURL + "/function/metrics?metrics_window=10y&function=" + function); err != nil {
		return 0, err
	} else if response.Body == nil {
		return 0, fmt.Errorf("no invocations for function: %s", function)
	} else {
		defer response.Body.Close()

		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return 0, err
		}
		// fmt.Println(string(bodyBytes))
		invocations := metrics.Metrics{}
		if err := json.Unmarshal(bodyBytes, &invocations); err != nil {
			return 0, err
		}
		totalInvocations := uint64(invocations.Success) + uint64(invocations.Failure)
		return totalInvocations, nil
	}
}

func calculateInvocationsCost(invocations, costPerUnitInvocations, unitInvocations uint64) uint64 {
	return uint64(costPerUnitInvocations * invocations / unitInvocations)
}

func calculateRemainingInvocations(balance, costPerUnitInvocations, unitInvocations uint64) uint64 {
	return uint64(balance * unitInvocations / costPerUnitInvocations)
}
