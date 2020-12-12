# jsonobj

convert command line args to JSON

Example:

    $ ./jsonobj --pretty id: @uuid created: @now foo: bar isNew: true intValue: 123 floatVal: 23.88 doNull: null andNil: nil
    {
        "andNil": null,
        "created": "2020-12-12T20:52:57Z",
        "doNull": null,
        "floatVal": 23.88,
        "foo": "bar",
        "id": "56f2b89f-64f9-497f-ac0f-20ad6bac21ce",
        "intValue": 123,
        "isNew": true
    }

Handy for building JSON from shell:

    $ ./jsonobj -p files: * pwd: `pwd`
    {
        "files": [
            "Makefile",
            "README.md",
            "go.mod",
            "go.sum",
            "jsonobj",
            "jsonobj.go",
            "jsonobj.lnx"
        ],
        "pwd": "/Users/pkelly/zorkspace/fxr/jsonobj"
    }

Mac & Linux binaries checked in.
