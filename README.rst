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

Find top-talkers
----------------

Interfaces most active receiving by Kb/s::

        rim -f ~/data/target_hosts.txt -n | sort -k3 -n -r | head -30

Interfaces most active transmitting by Packets/s::

        rim -f ~/data/target_hosts.txt -n | sort -k6 -n -r | head -30

Spot problems
-------------

Many anomalies on network interfaces can be easily spotted via Drops/s and Errors/s which are printed in columns 7-10::

        rim -f ~/data/target_hosts.txt -n | sort -k7 -k8 -n -r | head -30

``-n`` do not show titles. Without ``-p`` rim will try no password authentication and ``ssh-agent`` as fallback. Default user is root, another one can be used with ``-u`` flag.

Note
----

In case of problems getting info from remote hosts errors are printed to ``stderr`` so you must redirect it to stdout to propagate them throgh pipes::

        rim -f ~/data/target_hosts.txt -n 2>&1 | sort -k7 -k8 -n -r | less
