====
cfgp
====

|image0|_ 
|image1|_

.. |image0| image:: https://godoc.org/github.com/eraclitux/cfgp?status.png
.. _image0: https://godoc.org/github.com/eraclitux/cfgp

.. |image1| image:: https://drone.io/github.com/eraclitux/cfgp/status.png
.. _image1: https://drone.io/github.com/eraclitux/cfgp/latest

.. contents::

Intro
=====
A go package for configuration parsing. Automagically populates a configuration ``struct`` using configuration files & command line arguments.

It aims to be modular and easily extendible to support other formats. Only INI format supported for now.

Usage and examples
==================
An example of utilization::

        type myConf struct {
                Address string
                Port    string
                // A command line flag "-users", which expects an int value,
                // will be created.
                // Same key name will be searched in configuration file.
                NumberOfUsers int `cfgp:"users,number of users,"`
                Daemon        bool
                Message       string
        }

        func Example() {
                // To create a dafault value for a flag
                // assign it when instantiate the conf struct.
                c := myConf{Message: "A default value"}
                cfgp.Path = "test_data/one.ini"
                err := cfgp.Parse(&c)
                if err != nil {
                        log.Fatal("Unable to parse configuration", err)
                }
                fmt.Println("address:", c.Address)
                fmt.Println("port:", c.Port)
                fmt.Println("number of users:", c.NumberOfUsers)
        }

See the flag arguments that are automagically created::

        go run main.go -h

See `godocs <http://godoc.org/github.com/eraclitux/cfgp>`_ for examples and documentation.

Pull requests that add new tests, features or fixes are welcome, encouraged, and credited.
