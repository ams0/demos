# Kaito Testing with K6
# K6 is a modern load testing tool that provides a powerful scripting API and a rich set of features for testing web applications.


k6 run --vus 100 --duration 5m k6_llm_test.js

or ramping up the load over time:

```bash
# k6 run --vus 100 --duration 5m k6_llm_test.js --stage 30s:50 --stage 30s:100 --stage 4m:100 --stage 30s:0
```

This command will start with 50 virtual users for 30 seconds, ramp up to 100 virtual users for another 30 seconds, maintain 100 virtual users for 4 minutes, and then ramp down to 0 virtual users over the last 30 seconds.
