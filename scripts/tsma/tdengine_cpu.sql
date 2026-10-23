-- Run once after loading cpu-only data, with the benchmark database selected.
CREATE TSMA tsbs_cpu_1h ON cpu
FUNCTION(
    AVG(usage_user), AVG(usage_system), AVG(usage_idle), AVG(usage_nice),
    AVG(usage_iowait), AVG(usage_irq), AVG(usage_softirq), AVG(usage_steal),
    AVG(usage_guest), AVG(usage_guest_nice),
    MAX(usage_user), MAX(usage_system), MAX(usage_idle), MAX(usage_nice),
    MAX(usage_iowait), MAX(usage_irq), MAX(usage_softirq), MAX(usage_steal),
    MAX(usage_guest), MAX(usage_guest_nice)
) INTERVAL(1h);

CREATE TSMA tsbs_cpu_max_1m ON cpu
FUNCTION(MAX(usage_user), MAX(usage_system), MAX(usage_idle),
         MAX(usage_nice), MAX(usage_iowait)) INTERVAL(1m);
