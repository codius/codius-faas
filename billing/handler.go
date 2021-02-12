package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/openfaas-cloud/sdk"
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

func Handle(w http.ResponseWriter, r *http.Request) {

	fnName, err := parseFunctionName(r)
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

	credit, err := getFunctionBalance(r.Context(), fnName)
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
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func parseFunctionName(r *http.Request) (functionName string, error error) {
	functionNameQuery := r.URL.Query().Get("function")

	if len(functionNameQuery) > 0 {
		return functionNameQuery, nil
	}

	return "", fmt.Errorf("there is no `function` query param")
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

func getFunctionBalance(ctx context.Context, function string) (uint64, error) {
	redis_uri := os.Getenv("redis_uri")
	redisPassword, _ := sdk.ReadSecret("redis-password")
	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_uri,
		Password: redisPassword,
		DB:       0, // use default DB
	})
	balances_key_prefix := os.Getenv("balances_key_prefix")
	val, err := rdb.Get(ctx, balances_key_prefix+"|"+function).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	} else {
		// fmt.Println(val)
		balance, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("Couldn't parse balance to uint64. Value: %s, Error: %t", val, err)
		}
		return balance, nil
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
