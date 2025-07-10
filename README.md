# go-metrics-mqtt

The agent to collect `runtime/metrics` and publish them to MQTT broker.

The topics' names are generated atomatically and have syntax: `{client_id/Go_runntime_full_metric_name}`

For example: `my-test-app/cpu/classes/gc/mark/assist:cpu-seconds`

Please see `./examples` folder for usages (e.g., with AWS IoT core).