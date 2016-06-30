# go-datadog-autoscaler
Autoscale AWS EC2 groups based on custom datadog metric queries.


#### Datadog config

Needs datadog API key and APP key.

#### AWS config

Needs aws credentials defined in the environment.


#### TODO:
- Wire in "Honour Cooldown" when triggering autoscale. True by default