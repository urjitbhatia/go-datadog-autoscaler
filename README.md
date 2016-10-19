# go-datadog-autoscaler
Autoscale AWS EC2 groups based on custom datadog metric queries.


#### Datadog config

Needs datadog API key and APP key.

#### AWS config

Needs aws credentials defined in the environment.

#### Scale UP/DOWN Intervals

You can now provide multiple intervals. The `count` value of the interval in which the metric query value falls will determine the change.

Imagine the following numberline:

```
-Inf --- 40 --- 50 --- 60<ScaleDOWN -- 'Stable Zone' -- ScaleUP>70 --- 80 --- 90 --- +Inf
```

If you want your metric to stay in the `Stable Zone` then you can provide iterative steps above or below it. The farther away it is from the `Stable Zone`, the more correction you can apply.

Assuming that the value is `95`, you probably want to scale up by 5 to immediately relieve pressure but if it is `73`, you might be ok scaling up by just 2. Same applies for the `ScaleDOWN` direction.

Using just a single element `ScaleUP`/`ScaleDOWN` interval will also work, with intervals defined as the example:
```
-Inf --- 60<ScaleDOWN -- 'Stable Zone' -- ScaleUP>90 --- +Inf
```


```yaml
    # The "ScaleUp" operation (Required section)
    scaleUp:
        # Number of instances to add
      - count: 5
        # Trigger ScaleUp if the "Transformed" value is strictly > `threshold` value and < next higher threshold.
        threshold: 90
        # Honour cooldown
        cooldown: true
      - count: 3
        # Trigger ScaleUp if the "Transformed" value is strictly > `threshold` value and < next higher threshold.
        threshold: 80
        # Honour cooldown
        cooldown: true
      - count: 2
        # Trigger ScaleUp if the "Transformed" value is strictly > `threshold` value and < next higher threshold.
        threshold: 70
        # Honour cooldown
        cooldown: true

    # The "ScaleDown" operation (Required section)
    scaleDown:
        # Number of instances to remove from the autoscaling group
      - count: 1
        # Trigger ScaleDown if the "Transformed" value is strictly < `threshold` value and > next lower threshold.
        threshold: 60
        # Honour cooldown
        cooldown: true
      - count: 2
        # Trigger ScaleDown if the "Transformed" value is strictly < `threshold` value and > next lower threshold.
        threshold: 50
        # Honour cooldown
        cooldown: true
      - count: 4
        # Trigger ScaleDown if the "Transformed" value is strictly < `threshold` value and > next lower threshold.
        threshold: 40
        # Honour cooldown
        cooldown: true

##############################################################
# ScaleUP/DOWN thresholds are open intervals
#
# In the example above, consider the following metric values:
#
# 95 : Will scale UP by 5 for threshold 90 but below +Inf
# 87 : Will scale UP by 3 for threshold 80 but below 90
# 74 : Will scale UP by 2 for threshold 70 but below 80
#
# 65 : Will not have any effect
#
# 57 : Will scale DOWN by 1 for threshold 60 but above 50
# 45 : Will scale DOWN by 2 for threshold 50 but above 40
# 36 : Will scale DOWN by 4 for threshold 40 but above -Inf
##############################################################
```


#### TODO:
- Wire in "Honour Cooldown" when triggering autoscale. True by default