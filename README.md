# Golang Log Tracer

## Problem

We usually checking logs to understand application behaviour.
Sometimes application don't have enough logs.
For large apps it could take time to insert logs to all interesting places.

## Solution

Golang Log Tracer adds logs for each function call and return statements.
Addition could be done per-file for packages recursively.
It uses AST (Abstract Syntax Tree) to parse and modify code.

## Caution

Without `dry`-run, code will be modifyed in-place.
Make sure all your changes commited!

## Example

1. Initial code

    ```bash
    $ cat test/test.go
    package hello

    func hello(name string) string {
            helloStr := "Hello, " + name
            return helloStr
    }
    ```

2. Run golang-log-tracer

    ```bash
    $ ./golang-log-tracer -dry=false -paths=test/test.go
    INFO[0000] paths::: [test/test.go]
    INFO[0000] Checking path: test/test.go
    INFO[0000] Found 0: test/test.go
    INFO[0000] File: &{Doc:<nil> Package:1 Name:hello Decls:[0xc0000765a0] Scope:scope 0xc000074400 {
            func hello
    }
     Imports:[] Unresolved:[string string] Comments:[]}
    ```

3. Check result

    ```bash
    $ cat test/test.go
    package hello

    import (
            "github.com/sirupsen/logrus"
    )

    func hello(name string) string {
            logrus.Infof("CALL>>> test/test.go:hello(name:%+v, )",

                    name)
            defer logrus.Infof("RET>>> ",
            )
            helloStr := "Hello, " + name
            return helloStr
    }
    ```
