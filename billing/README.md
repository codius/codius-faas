## billing

[OpenFaaS](https://www.openfaas.com/) function that reports the current balance of an [OpenFaaS Cloud](https://docs.openfaas.com/openfaas-cloud/intro/) hosted function.

It takes function's name (i.e. `myFunction` \[required\]) from a query and is invoked by GET request to
http://gateway-url:8080/function/billing?function=myFunction

- queries [receipt-verifier](https://github.com/coilhq/receipt-verifier) for the function's total amount paid
- queries the [gateway](https://github.com/openfaas/faas/blob/master/gateway/README.md) for the total number of function invocations
- returns a JSON response of type:

```json
{
    "balance": "10",
    "remainingInvocations": "8"
}
```

### Environment Variables

All environment variables are required.

| Environment Variable        | Description |
| --------------------------- | ------------------------------------------------------------------------------------------------------ |
| `balances_key_prefix`       | Prefix for balances Redis keys. |
| `cost_per_unit_invocations` | Cost per `unit_invocations` for function invocations denominated in the host's asset (code and scale). |
| `unit_invocations`          | :point_up_2: |
| `bonus_invocations`         | The number of allowed invocations if the function balance is insufficient. |
| `prometheus_host`           | Host to connect to Prometheus. |
| `prometheus_port`           | Port to connect to Prometheus. |
| `redis_uri`                 | The URI at which to connect to Redis. |
