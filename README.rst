===============================
RIM - Remote Interfaces Monitor
===============================

Command line tool to get status of remote network interfaces on linux servers. It's like a ``vmstat`` for remote NICs.

On a multicore machine can concurrently handle hundreds of servers per time, fast.

It reads information exposed through ``/proc`` file system using ssh connections so no remote agents are needed on targets. Even *linux bridges* are included in report.

Find incoming and outgoing **DDoS** in your network in a snap, even before NetFlow probes!

.. contents::

Usage examples
==============

Put target hostnames in a file, one per line es.: ``~/data/target_hosts.txt``. It is possible to specify a different port than ``22`` using syntax::

        myhost.tld[:port]

Sorting
-------

``-k1`` & ``-k2`` set hierarchical sort keys. Supported sorting keys are::

        tx-Kbps, tx-pps, tx-eps, tx-dps, rx-Kbps, rx-pps, rx-eps, rx-dps

*Default sort settings* are ``1st: rx-dps`` & ``2nd: rx-Kbps`` because these have proven to be the most effective spotting anomalies in the network of cloud service provider where rim has born.

Find top-talkers
----------------

Interfaces most active receiving by Kb/s::

        rim -f ~/data/target_hosts.txt -k1 rx-Kbps

Interfaces most active transmitting by Packets/s, the first ten (useful to spot out going DDoS)::

        rim -f ~/data/target_hosts.txt -k1 tx-pps -l 10

It's also possible to use ``rim`` in a pipe::

        cat ~/data/target_hosts.txt | rim | less

Notes
~~~~~

In case of problems getting info from remote hosts, errors are printed to ``stderr`` so you must redirect it to stdout to propagate them throgh pipes::

        rim -f ~/data/target_hosts.txt -n 2>&1 | less

Spot problems
-------------

Many anomalies on network interfaces can be easily spotted via Drops/s and Errors/s.

Default sort key are for rx data, to show tx data::

        rim -f ~/data/target_hosts.txt -k1 tx-dps -k2 tx-Kbps

To print also Errors/s ``-e`` option must be used.

``-n`` do not show titles. Without ``-p`` ``rim`` will try no password authentication and ``ssh-agent`` as fallback. Default user is root, another one can be used with ``-u`` flag.

Configuration
-------------

A configuration file can be used to specify configuration parameters. File must be end with ``.cfg``. Use env var ``RIM_CONF_FILE`` to specify its path. You could put::

        export RIM_CONF_FILE=/path/to/conf.cfg

in your ``.bashrc``.

Available parameters can be showed with ``rim -h``, lowercase first letter when use them in file. For example to specify ``HostsFile``::

        hostsFile = /path/to/file

Build/Install
-------------

With a proper Go environment installed just run::

        godep go build

To install in ``$GOPATH/bin``::

        godep go install

Changelog
---------

- v2.2.0-beta: show a spinner.
- v2.1.0-beta: add connection timeout parameter.
- v2.0.0-beta: configuration file capabilities.
- v2.0.0-alpha: it adds sort capabilities, no more need to pipe the output to ``sort``. It breaks APIs (output changed).
- v1.0.0: initial relase, retrieve info from remote hosts via ssh.
