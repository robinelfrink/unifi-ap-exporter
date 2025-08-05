# Prometheus metrics exporter for the Unifi AC Lite series access points

This program collects metrics from
[Unifi AC Lite](https://techspecs.ui.com/unifi/wifi/u7-lite) WiFi access
points, and presents them in Prometheus format. There is no need to run the
[UniFi Network Application](https://ui.com) as the metrics are collected
directly from the access points.

## Usage

The program expects a configuration file in [YAML](https://yaml.org/) format:

```yaml
global:
  port: 9130  # optional
accesspoints:
  - name: my-access-point
    username: admin
    password: secret  # optional
    keyfile: ssh-private-key-file # optional
  - name: my-other-access-point
    ...
```

Either a `password` or an SSH private `keyfile` is needed. The password is the
one that can be configured using the smartphone app. Configuring the access
point to use an SSH key is left to your Google skills.

The `name` of the accesspoint is used as a label on all metrics, for
identification and correlation purposes.

## Running with Docker

```shell
$ docker run \
      --detach \
      --rm \
      --name unif-ap-exporter \
      --publish 9130:9130 \
      --workdir /work \
      --volume $(pwd)/unifi-ap-exporter.yaml:/work/unifi-ap-exporter.yaml \
      --volume $(pwd)/sshkey:/work/sshkey \
      ghcr.io/robinelfrink/unifi-ap-exporter:latest
```

## Compatibility

*  This program has been tested with UAP-AC-Lite devices running firmware version
6.6.77.
*  Output metrics are compatible with those of [Unpoller](https://unpoller.com/)
as much as possible.
