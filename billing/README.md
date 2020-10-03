## billing

[OpenFaaS](https://www.openfaas.com/) function that reports the current balance of an [OpenFaaS Cloud](https://docs.openfaas.com/openfaas-cloud/intro/) hosted function.

It takes function's name (i.e. `myFunction` \[required\]) from a query and is invoked by GET request to
http://gateway-url:8080/function/billing?function=myFunction

- queries [receipt-verifier](https://github.com/coilhq/receipt-verifier) for the function's total amount paid
- queries the [metrics](https://github.com/openfaas/openfaas-cloud/tree/master/metrics) function for the total number of funciton invocations
- returns a JSON response of type:

```json
{
    "balance": 10,
    "remainingInvocations": 8
}
```

### Environment Variables

All environment variables are required.

| Environment Variable        | Description |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `gateway_url`               | Cluster-internal URL of the OpenFaaS gateway service. |
| `balances_url`              | Root URI of the [receipt verifier](https://github.com/coilhq/receipt-verifier)'s `balances` API. |
| `cost_per_unit_invocations` | Cost per `unit_invocations` for function invocations denominated in the host's asset (code and scale). |
| `unit_invocations`          | :point_up_2: |
| `bonus_invocations`         | The number of allowed invocations if the function balance is insufficient. |
