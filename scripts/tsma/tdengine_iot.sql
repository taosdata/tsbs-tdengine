-- Run once after loading the IoT data, with the benchmark database selected.
CREATE TSMA tsbs_readings_10m ON readings
FUNCTION(AVG(velocity)) INTERVAL(10m);

CREATE TSMA tsbs_diagnostics_10m ON diagnostics
FUNCTION(AVG(status), AVG(current_load)) INTERVAL(10m);
