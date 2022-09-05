# Amplitude exporter

Amplitude charts to prometheus exporter PoC. Work in progress… 

Exposes amplitude chart data as prometheus metrics at http://localhost:8080/metrics

## Config

./config.yaml
```yaml
projects:
- name: Project1
  apiId: <amplitude Project 1 api username>
  apiKey: <amplitude Project 1 api password>
  charts:
    - id: <chartId1>
      labels: ["sre"]
      subsystem: socket
      name: error
    - id: <chartId2>
      labels: ["sre"]
      subsystem: zzz
      name: error
      type: gauge # default is counter
- name: Project2
  apiId: <amplitude Project 2 api username>
  apiKey: <amplitude Project 2 api password>
  charts:
    - …
```

## Example output

```
# HELP amplitude_exporter_scrapes_total Current total Amplitude scrapes.
# TYPE amplitude_exporter_scrapes_total counter
amplitude_exporter_scrapes_total 2
# HELP amplitude_socket2_error 
# TYPE amplitude_socket2_error gauge
amplitude_socket2_error{sre="sre"} 2
# HELP amplitude_socket_error 
# TYPE amplitude_socket_error gauge
amplitude_socket_error{sre="sre"} 2
# HELP amplitude_up Was the last scrape of Amplitude successful.
# TYPE amplitude_up gauge
amplitude_up 1
```