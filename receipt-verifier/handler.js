"use strict"

const fetch = require('node-fetch')
const Redis = require("ioredis")

let redisReady = false

const redis = new Redis(process.env.redis_uri)
redis.on('ready', () => {
  redisReady = true
})

module.exports = async (event, context) => {
  const res = await fetch(`${process.env.receipt_verifier_uri}/verify`, {
    method: 'post',
    body: event.body
  })

  if (!res.ok) {
    return context
      .status(res.status)
      .fail(res.statusText)
  }

  const receiptDetails = await res.json()
  if (receiptDetails.id &&
      receiptDetails.spspEndpoint === process.env.payment_pointer) {
    while (!redisReady) {
      await new Promise(resolve => setImmediate(resolve))
    }
    await redis.incrby(`${process.env.balances_key_prefix}|${receiptDetails.id}`, receiptDetails.amount)
  }

  context
    .status(res.status)
    .succeed(receiptDetails)
}
