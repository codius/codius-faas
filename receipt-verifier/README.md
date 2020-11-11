## receipt-verifier

[OpenFaaS](https://www.openfaas.com/) function that wraps the [`receipt-verifier`](https://github.com/coilhq/receipt-verifier) `/verify` endpoint and credits hosted function balances in Redis.

### Environment Variables

All environment variables are required.

| Environment Variable   | Description |
| ---------------------- | -------------------------------------------------------------------------- |
| `balances_key_prefix`	 | Prefix for balances Redis keys.                                            |
| `payment_pointer`      | Host payment pointer. Function balance will be credited for amounts to it. |
| `receipt_verifier_uri` | The URI of the `receipt-verifier`.                                         |
| `redis_uri`            | The URI at which to connect to Redis.                                      |
