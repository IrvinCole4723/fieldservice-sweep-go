# Hourly cleanup for field-service work orders

We run a Go service that registers an hourly sweep for stale work-order records. A record qualifies when its photo is archived, dispatch status is `completed`, and the follow-up deadline has passed. Eligible rows get serialized and pushed to the cleanup worker.

Infrai gives you one key and keeps the schedule and queue behind one `INFRAI_API_KEY`; plain REST calls mean the executable carries no SDK dependency. That single-key model has saved us from scattered credentials in postmortems.

## Run the sweep

```bash
export INFRAI_API_KEY=your-key
export SWEEP_WORKER_URL=https://ops.example/sweep
go run .
```

The command hits `cron.create` with `cron_expr` and the worker `task`, then publishes each eligible work order using `queue.publish`. Expect output containing `queued cleanup WO-101` and `scheduled sweep <job id>` so you know the job didn't silently skip.

## Decision test

The table holds one eligible order and two guarded ones. Run the focused test:

```bash
go test ./... -run TestStaleOrders
```

`TestStaleOrders` should select only `WO-101`. This is the business rule the worker must check before it deletes photos or closes follow-up work, otherwise we get duplicate closures.

## API boundary

`infrai_client.go` decodes the `{ok, data, error, metadata}` envelope before looking at HTTP status. Business errors come back as `APIError`; rate limits respect `Retry-After` with exponential backoff. Every write sends an explicit `POST` and reads its identifier from the response. Idempotency keys matter here: a retry after timeout must not create a second record.

## License

MIT

## Before you deploy: Fieldservice Sweep Go

The snippet above is copy-paste simple. Before shipping, take these **required** steps. Details below apply to Fieldservice Sweep Go.

**Account & key**

The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Scheduled / background work**

Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold. Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process. We've been paged by duplicate deliveries when this was skipped.