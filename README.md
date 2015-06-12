Better Living Through Statistics: Monitoring Doesn't Have To Suck.
==================================================================


This is the source code for the demo, plus slide material, that I used when I
presented this talk to

    * PuppetConf 2012
    * again at OSDC 2013
    * at PuppetCamp Sydney 2013
    * using Prometheus as a demo at Monitorama 2015

The URL for the video from PuppetConf 2012 is http://youtu.be/eq4CnIzw-pE 

The original slide deck is at https://docs.google.com/presentation/d/1uTLggLR5HICnSyhTJyQWNeYWZ6niHyCsup7wNFByex4/pub

I presented an abridged version to the Sysadmin Miniconf at Linux.conf.au 2014, the revised slide deck is here:

https://docs.google.com/presentation/d/1Dq4eRUlkONnVnnXg6M_ZSi6xBLEwe7kjwjx74vFL1N4/pub

To use the demo code
--------------------

The demo is of a mock service, a web applicatoin frontend, composed of a cluster of application servers and a single loadbalancer.  The loadbalancer is not very good, and the application servers fail often.  An antagonistic load generator drives them past their capable limits.  A collector extracts metrics from all the members of the service.

The code uses Go but not in a Go friendly project layout.  But you wouldn't try to import this into your code, would you?

`make` will build the things you need.

`./servers.sh` runs the servers.

`./load.sh` requires `ab` (Apache Bench) from the apache webserver tools.

`./prom.sh` launches Prometheus with the included configuration.  You may need to change the path of the binary in this script.

This code is available under the Apache v2 license.
