# goerr
This package defines an error type that maintains a stack of nested errors and gives a human readable stack trace for logging.

# Problem
In typical go `error` type, you don't have the stack trace of complete call chain. The only possibility is to log the stack trace at every function in the call chain. This will have multiple log entries made from single call and also the log are very cumbersome. 
Also the default stack trace contains lot many details that becomes difficult to read.

# Principle
This package is to facilitate the well-known idiom

"throw error multiple times, but log once at the top most level"

For eg., if the call chain is `controller` -> `service` -> `repository`. 

Then with `goerr` you return error from repository to service to controller and log the same in controller.

# Output
If you return `goerr` in all methods and get the stack at top most level, it gives below nicely formatted, easily readable stack

```shell
controller failed [/Users/madhan.ganesh/src/github.com/angel-one/goerr/samplesrc/samples.go:11 (samplesrc.Controller)]
    service failed [/Users/madhan.ganesh/src/github.com/angel-one/goerr/samplesrc/samples.go:19 (samplesrc.Service)]
        error from database [/Users/madhan.ganesh/src/github.com/angel-one/goerr/samplesrc/samples.go:26 (samplesrc.Repository)]
```

# Usage
Whenver you wanted to return an error just use
```go
err := goerr.New(nil, "error in here")
```

if you have to nest the error, just pass it in New
```go
err1 := goerr.New(err, "error in here")
```

# Sample code that log in nested methods
```go
func Controller() error{
	err := Service()
	if err != nil {
		return goerr.New(err, "controller failed")
	}
	return nil
}

func Service() error {
	err := Repository()
	if err != nil {
		return goerr.New(err, "service failed")
	}
	return err
}

func Repository() error {
	err := errors.New("error from database")
	return goerr.New(nil, err.Error())
}
```
# Code to get the stack at the top most level
```go
        err := samplesrc.Controller()
	if err != nil {
		t.Logf("error in controller: %s", goerr.Stack(err))
	}
```

# Return goerr with an error code
`goerr` has ability to send an error code of int type. As part of the stack each `goerr` returned can optionally sent the error code. By default this code will be a `0`
```
func demo() error {
	return goerr.New(err, http.StatusNotAllowed, "key conflict")
}
```

In the `New` method if the second parameter is an `int` value that will be taken as error code.

If `goerr` has any error code that will be returned as part of the stack trace. See **(409)** in below sampel stack
```
controller error [goerr_test.go:171 (func3)]
    service error [goerr_test.go:164 (func2)]
        repository error (409) [goerr_test.go:159 (func1)]
```

## Retrieve code from goerr explicitly
You can also explicitly retrieve the error explicitky from `goerr.Code(err)` method.
```
err := demo()
code := goerr.Code(err)
```
The `Code` method iterates until it fonds the error code in teh stack. It stop in the first `goerr` that has an error code. This means you get the error code that is last given in the call chain.

# Compatibility
- `goerr` implements standard `error` interface, so can be assigned where ever error is used
- `goerr.Error()` will give the error text of the top most error object
- `goerr.Stack(err)` will give stack trace of call chain
- `goerr.Stack(err)` can be called for error type as well, in which case it will just return `Error()`
- `goerr` supports error checking and handling via the standard `errors.Is` and `errors.As` functions

# Installation
```shell
go get github.com/angel-one/goerr
```
