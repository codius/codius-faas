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

	"github.com/openfaas/faas/gateway/metrics"
)

type FunctionBalance struct {
	Balance     uint64 `json:"balance,string"`
	Invocations uint64 `json:"remainingInvocations,string"`
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

func calculateInvocationsCost(invocations, costPerUnitInvocations, unitInvocations uint64) uint64 {
	return uint64(costPerUnitInvocations * invocations / unitInvocations)
}

func calculateRemainingInvocations(balance, costPerUnitInvocations, unitInvocations uint64) uint64 {
	return uint64(balance * unitInvocations / costPerUnitInvocations)
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

	invocations, err := getFunctionInvocations(fnName)
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

func getFunctionInvocations(function string) (uint64, error) {
	host := os.Getenv("prometheus_host")
	envPort := os.Getenv("prometheus_port")
	port, err := strconv.Atoi(envPort)
	if err != nil {
		log.Fatalf("Could not convert env-var prometheus_port to int. Env-var value: %s. Error: %t", envPort, err)
	}

	metricsQuery := metrics.NewPrometheusQuery(host, port, &http.Client{})

	queryValue := fmt.Sprintf(
		`sum(gateway_function_invocation_total{function_name="%s"})`,
		function,
	)
	expr := url.QueryEscape(queryValue)

	response, err := metricsQuery.Fetch(expr)
	if err != nil {
		return 0, fmt.Errorf("Failed to get query metrics for function %s, error: %t", function, err)
	}

	invocationCount := uint64(0)
	for _, v := range response.Data.Result {

		metricValue := v.Value[1]
		switch metricValue.(type) {
		case string:
			f, err := strconv.ParseUint(metricValue.(string), 10, 64)
			if err != nil {
				log.Printf("Unable to convert value for metric: %s\n", err)
				continue
			}
			invocationCount += f
			break
		}
	}

	return invocationCount, nil
}
