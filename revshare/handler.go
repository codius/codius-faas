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
	"time"

	faasSDK "github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/openfaas-cloud/sdk"
)

type FaaSAuth struct{}

func (auth *FaaSAuth) Set(req *http.Request) error {
	return sdk.AddBasicAuth(req)
}

var (
	timeout   = 3 * time.Second
	namespace = ""
)

// TODO: import me
type FunctionBalance struct {
	Balance     uint64 `json:"balance,string"`
	Invocations uint64 `json:"remainingInvocations,string"`
}

type RevshareResponse struct {
	BalanceId    string `json:"balanceId,omitempty"`
	SpspEndpoint string `json:"spspEndpoint"`
}

func Handle(req []byte) string {

	fnName, err := parseFunctionName()
	if err != nil {
		log.Fatalf("couldn't parse function name from query: %t", err)
	}

	paymentPointer, exists := os.LookupEnv("payment_pointer")
	if !exists || len(paymentPointer) == 0 {
		log.Fatal("`payment_pointer` must be set")
	}

	var resp RevshareResponse
	if fnName == ".well-known/pay" {
		resp = RevshareResponse{
			SpspEndpoint: paymentPointer,
		}
	} else {
		gatewayURL := os.Getenv("gateway_url")
		balance, err := getFunctionBalance(fnName, gatewayURL)
		if err != nil {
			log.Fatalf("Couldn't get balance for function %s, %t", fnName, err)
		}
		if balance == 0 {
			resp = RevshareResponse{
				SpspEndpoint: paymentPointer,
				BalanceId:    fnName,
			}
		} else {
			fnPaymentPointer, err := getFunctionPaymentPointer(fnName, gatewayURL)
			if err != nil {
				log.Fatalf("Couldn't get payment pointer for function %s, %t", fnName, err)
			}
			resp = RevshareResponse{
				SpspEndpoint: fnPaymentPointer,
			}
		}
	}

	res, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Couldn't marshal json %t", err)
	}
	return string(res)
}

func parseFunctionName() (functionName string, error error) {
	if query, exists := os.LookupEnv("Http_Query"); exists {
		vals, _ := url.ParseQuery(query)

		functionNameQuery := vals.Get("id")

		if len(functionNameQuery) > 0 {
			return functionNameQuery, nil
		}

		return "", fmt.Errorf("there is no `id` inside env var Http_Query")
	}

	return "", fmt.Errorf("unable to parse Http_Query")
}

func getFunctionBalance(function, gatewayURL string) (uint64, error) {
	if response, err := http.Get(gatewayURL + "/function/billing?function=" + function); err != nil {
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
		balance := FunctionBalance{}
		if err := json.Unmarshal(bodyBytes, &balance); err != nil {
			return 0, err
		}
		return balance.Balance, nil
	}
}

func getFunctionPaymentPointer(function, gatewayURL string) (string, error) {
	client, err := faasSDK.NewClient(&FaaSAuth{}, gatewayURL, nil, &timeout)
	if err != nil {
		return "", err
	}

	functionStatus, err := client.GetFunctionInfo(context.Background(), function, namespace)
	if err != nil {
		return "", err
	}

	if functionStatus.Annotations != nil {
		paymentPointer, ok := (*functionStatus.Annotations)["interledger.org/payment-pointer"]
		if ok {
			return paymentPointer, nil
		}
	}
	return "", fmt.Errorf("no payment pointer annotation for function: %s", function)
}
