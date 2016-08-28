=======
Stracer
=======

|image0|_ 

.. |image0| image:: https://godoc.org/github.com/eraclitux/stracer?status.png
.. _image0: https://godoc.org/github.com/eraclitux/stracer

Package ``stracer`` is possibly the simplest tracing package. To enable its functions to print just build with ``debug`` tag otherwise they will do noop::

        go build -tags debug

``stderr`` is used to not perturb example functions.

Credits
=======

Original idea is by Dave Cheney http://dave.cheney.net.
