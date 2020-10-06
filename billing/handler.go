package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/openfaas/openfaas-cloud/sdk"

	faasSDK "github.com/openfaas/faas-cli/proxy"
)

type FaaSAuth struct{}

func (auth *FaaSAuth) Set(req *http.Request) error {
	return sdk.AddBasicAuth(req)
}

var (
	timeout   = 3 * time.Second
	namespace = ""
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
	} else {
		switch response.StatusCode {
		case http.StatusNotFound:
			return 0, nil
		case http.StatusOK:
			if response.Body == nil {
				return 0, fmt.Errorf("no balance for function: %s", function)
			}
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
		default:
			return 0, fmt.Errorf("Unexpected response for function: %s, Code: %s", function, http.StatusText(response.StatusCode))
		}
	}
	return 100, nil
}

func getFunctionInvocations(function, gatewayURL string) (uint64, error) {
	client, err := faasSDK.NewClient(&FaaSAuth{}, gatewayURL, nil, &timeout)
	if err != nil {
		return 0, err
	}

	functionStatus, err := client.GetFunctionInfo(context.Background(), function, namespace)
	if err != nil {
		return 0, err
	}

	return uint64(functionStatus.InvocationCount), nil
}

func calculateInvocationsCost(invocations, costPerUnitInvocations, unitInvocations uint64) uint64 {
	return uint64(costPerUnitInvocations * invocations / unitInvocations)
}

func calculateRemainingInvocations(balance, costPerUnitInvocations, unitInvocations uint64) uint64 {
	return uint64(balance * unitInvocations / costPerUnitInvocations)
}
