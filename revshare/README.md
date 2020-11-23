## revshare

[OpenFaaS](https://www.openfaas.com/) function returns the payment pointer to send SPSP queries to based on the current balance of an [OpenFaaS Cloud](https://docs.openfaas.com/openfaas-cloud/intro/) hosted function.

It takes function's name (i.e. `myFunction` \[required\]) from a query and is invoked by GET request to
http://gateway-url:8080/function/revshare?function=myFunction

- queries [balance](https://github.com/codius/codius-faas/tree/main/billing) function for the function's current balance
- queries the [gateway](https://github.com/openfaas/faas/blob/master/gateway/README.md) for the total number of function invocations
- returns a string response of the SPSP endpoint to proxy the SPSP query to

### Environment Variables

All environment variables are required.

| Environment Variable        | Description |
| --------------------------- | ------------------------------------------------------------------------------------------------------ |
| `gateway_url`               | Cluster-internal URL of the OpenFaaS gateway service. |
