Better Living Through Statistics: Monitoring Doesn't Have To Suck.
==================================================================

This is the source code for the demo that I used when I presented this talk to

* PuppetConf 2012
* again at OSDC 2013
* at PuppetCamp Sydney 2013
* Sysadmin Miniconf at linux.conf.au 2014
* using Prometheus as a demo at Monitorama 2015
* again at linux.conf.au 2016
* Velocity SC 2016
* SRECon Americas 2017
* incorporating SLO Burn alerting for Monitorama 2018, SRECon Asia 2018, Velocity SJ 2018, and Velocity NY 2018.

There are Git tags for each presentation so you can jump back to each demo if that pleases you.

The URL for the video from PuppetConf 2012 is http://youtu.be/eq4CnIzw-pE

The original slide deck is at https://docs.google.com/presentation/d/1uTLggLR5HICnSyhTJyQWNeYWZ6niHyCsup7wNFByex4/pub

I presented an abridged version to the Sysadmin Miniconf at Linux.conf.au 2014, the revised slide deck is here:

https://docs.google.com/presentation/d/1Dq4eRUlkONnVnnXg6M_ZSi6xBLEwe7kjwjx74vFL1N4/pub

Monitorama 2015 slides: https://docs.google.com/presentation/d/1X1rKozAUuF2MVc1YXElFWq9wkcWv3Axdldl8LOH9Vik/edit

linux.conf.au 2016 slides: https://docs.google.com/presentation/d/1NziwSTwuz91fqsFhXeOGwyhFUoT6ght1irA_0ABLPU0/edit

To use the demo code
--------------------

The code is likely to break without notice.  At best I can promise you it worked at the time I gave the most recent demo, with the dependencies available at the time.  YMMV.

The demo is of a mock service, a web application frontend, composed of a
cluster of application servers and a single loadbalancer.  The loadbalancer is
not very good, and the application servers fail often.  An antagonistic load
generator drives them past their capable limits.

[Prometheus](http://prometheus.io) is the metrics collector and alerting engine
used in this example.  [Zipkin](http://zipkin.io) is used for capturing traces
if you choose to add it.

The demo code uses Go but not in a Go friendly project layout.  But you
wouldn't try to import this into your code, would you?

`make` will build the things you need.

`./servers.sh` runs the servers.

`./load.sh` requires `ab` (ApacheBench) from the Apache webserver tools.  `./load-nice.sh` is the non-antagonistic version.

`./prom.sh` launches Prometheus with the included configuration.  You will need to change the path of the binary in this script.  It changes the on-disk storage path to a local path.

`./am.sh` runs the Prometheus Alertmanager with the included configuration.  You will need to chagne the path of the binary in the script.  It also changes on-disk storage to a local path.

`./grafana.sh` runs Grafana and shows the SLO Burn demo console.  It uses a system installed version, but attempts to run it isolated from the system by resetting the homepath and config locations.  `graf/grafana.ini` makes an unauthenticated service.

`./1000concurrent.sh` keeps 1000 concurrent HTTP sessions open to the loadbalancer.  Because the system latency is about 1s average, this means around 100 qps.  Thanks Little's Law!

`./replace.sh` takes the PID of one backend server and replaces it with a backend that fails more often.  Killing this script causes the entire backend to die.

`./zipkin.sh` runs a local openzipkin in a docker container (based on the [zipkin quickstart](https://zipkin.io/pages/quickstart)) if you want to try out the [opencensus](http://opencensus.io) tracing examples.

The subdirectory `prom` contains the main config and rules for the Prometheus tools.

The R code is from the version back in 2012 when Prometheus didn't exist, and I couldn't talk about Borgmon. ;-)  I like to keep it here as a reminder.

This code is available under the Apache v2 license.

# Demos

## Timeseries based alerting

The first demo series is based around the idea of alerting from timeseries, not from check scripts.

After running `make`, run the following scripts (at the same time, in different terminals):

* `./servers.sh`
* `./prom.sh`
* `./am.sh`

then start the demo with `./load-nice.sh`.  Go to the prometheus console at http://127.0.0.1:9090 and observe no alerts being fired.

Stop load-nice and start up `./load.sh`.  The system should quickly cascade to failure and you'll get some alerts firing.

The alerts are defined in [prom/tasks.rules](prom/tasks.rules), [prom/errors.rules](prom/errors.rules), and [prom/latency.rules](prom/latency.rules).

## SLO Burn alerting

The second demo shows how to avoid all those alerts and focus on the overall health of the system, i.e. by having defined  service level objective, let us know when that objective is in danger of being missed.

After running `make`, run the following scripts (at the same time, in different terminals):

* `./servers.sh`
* `./prom.sh`
* `./grafana.sh`
* `./1000concurrent.sh`

Look at the Grafana console at http://127.0.0.1:3000 and load the SLO Burn console.  The Burn rate vs Threshold chart shows you the current short term burn rate vs the estimated threshold.  The threshold estimate is based on a prediction of the total events over the SLO measurement period, but at a consumption rate faster the Burn Period, i.e. page if we are burning at a rate that would consume the entire error budget for one month in the next day.  The maths can be seen in [prom/slo.rules.yml](prom/slo.rules.yml)

Use the `./replace.sh` script to kill the pid of the process that has port 8009 open (ps ef | grep "port :8009"), and see a higher failure rate not yet page because the SLO burn rate is not breached yet.  Then ^C the replace script, killing that backend, and there should be a high enough failure rate to trigger the SLO burn alert.



** This is not an officially supported Google product **

There is a similar project at https://github.com/google/prometheus-slo-burn-example 
