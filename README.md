# Gollector Monitors: Third-Party monitoring services/commands.

This repository contains several specialized monitors that don't really make
sense for [gollector](https://github.com/erikh/gollector) itself. Usage varies
by program, but all return JSON so if you don't want to use gollector, you
don't have to -- just something that consumes JSON.

Many of the monitors are in the zygote phase and may not be very comprehensive.
You've been warned.

## Monitors:

* `ping-monitor`: monitors several hosts and reports metrics (such as median
  and 99.9th percentile) on their ICMP echo round trip time. Exposes itself on
  `localhost:9119` for querying with the
  [json\_poll](https://github.com/erikh/gollector/wiki/JSON-Poll) gollector
  plugin. This monitor must be run as root.
* `process-monitor`: monitors several process related statistics in
  relationship to the binary that was responsible for executing the process.
  Listens on localhost:9118 for using with [json\_poll](https://github.com/erikh/gollector/wiki/JSON-Poll).
  This monitor must be run as root.
* `redis-monitor`: jsonification of the `info` redis command. For use with the
  [command](https://github.com/erikh/gollector/wiki/Command) plugin.
* `postgresql-monitor`: reports several metrics for use with the [command](https://github.com/erikh/gollector/wiki/Command) plugin:
  * materialized views
  * open locks
  * open cursors
  * open prepared statements
  * open prepared transactions

## Building

If you want to build a specific monitor, type `make monitor-name`. If you want
to build all monitors, just type `make`. Note that you must have a working
Golang (1.2 preferred) environment to build the software.

## License

* (C) 2014 Erik Hollensbe -- MIT Licensed

## Author(s)

* Erik Hollensbe <erik+github@hollensbe.org>
