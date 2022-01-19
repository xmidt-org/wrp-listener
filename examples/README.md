# wrp-listener examples

# Testing Locally

Leverage Docker to stand up a local xmidt cluster for validating webhooks.

1. Clone xmidt deploy repo

```bash
git clone https://github.com/xmidt-org/xmidt.git
```

2. Start xmidt cluster

```bash
cd deploy/docker-compose/
./deploy.sh
```

3. Build and Run example inside docker

```bash
cd <example test dir>
docker-compose up
```

4. Validate webhook is registered by
   visiting http://localhost:9090/graph?g0.expr=xmidt_caduceus_webhook_list_size_value&g0.tab=1&g0.stacked=0&g0.show_exemplars=0&g0.range_input=1h

5. Restart the Simulator and watch logs and
   prometheus http://localhost:9090/graph?g0.expr=xmidt_caduceus_delivery_count&g0.tab=0&g0.stacked=0&g0.show_exemplars=0&g0.range_input=1h