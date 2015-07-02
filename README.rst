===============================
RIM - Remote Interfaces Monitor
===============================

Command line tool to get status of remote network interfaces on linux servers. It's like a ``vmstat`` for remote NICs.

On a multicore machine can concurrently handle hundreds of servers per time, fast.

It reads information exposed through ``/proc`` file system using ssh connections so no remote agents are needed on targets. Even *linux bridges* are included in report.

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

Interfaces most active transmitting by Packets/s, the first ten::

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

Build/Install
-------------

With a proper Go environment installed just run::

        godep go build

or::
        godep go install

Changelog
---------

- v2.0.0: it adds sort capabilities, no more need to pipe the output to ``sort``. It breaks APIs (output changed).
- v1.0.0: initial relase, retrieve info from remote hosts via ssh.
