==========
GoParallel
==========

|image0|_ |image1|_

.. |image0| image:: https://godoc.org/github.com/eraclitux/goparallel?status.png
.. _image0: https://godoc.org/github.com/eraclitux/goparallel

.. |image1| image:: https://drone.io/github.com/eraclitux/goparallel/status.png
.. _image1: https://drone.io/github.com/eraclitux/goparallel/latest

Package ``goparallel`` simplifies use of parallel (as not concurrent) workers that run on their own core.
Number of workers is adjusted at runtime in base of numbers of cores.
This paradigm is particulary uselfull in presence of heavy, indipended tasks.
