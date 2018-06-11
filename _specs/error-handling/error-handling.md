---
layout: default
---
# Error handling for functions

## Types of Errors
### “User” / Input Errors

These errors are detected **by functions** themselves (early in the function body) or **input schema validation**. When such error is detected, it makes no sense to proceed: if a function expects a string, but the input is a map, there’s nothing meaningful the function can do with it. 

It is normal for a well implemented function to reject wrong input. Such errors mean that something is wrong with the **input data**. 


### Function Errors

These errors are thrown (usually **by the language runtime**) as a result of a logic error, such as NullPointerException or ArithmeticException (“divide by zero”). 

Such error means that something is wrong with the **function**. 

Examples:

- external services failures
- output validation errors
- timeout error
- other runtime exceptions


### System / Infrastructure Errors

These errors prevent function invocation from happening or communicating the results of the invocation back. 

Such error means something is wrong with Dispatch itself or its underlying infrastructure. 


## How errors are thrown

User errors (thrown by function code) and function errors are language dependent. To reduce dependency on image-specific conventions, language standard errors should be used whenever possible. 

It is the responsibility of language base-images to communicate those errors back in the output context.


## Communicating errors

Input and function errors should be recorded in the invocation output context and communicated back to the original caller. If a user API (via the API Gateway) call translated to a function invocation resulting in an input or function error, that error should be communicated back through that API using an appropriate HTTP status.

System errors are logged to Dispatch system logs. If a system error is detected invoking a function, it should also be recorded in the output context. 


### Error structure

If an invocation resulted in an error, the output context will have a non-null `error` field. 
For example (in YAML):

    context:
      error:
        type: FunctionError
        message: 'Something strange happened'
        stacktrace: ['...', '...']
      logs:
        stdout: ['...', '...']
        stderr: ['...', '...']
    payload: ~

Error fields:

- `**type**` — one of `InputError`, `FunctionError` or `SystemError`
- `**message**` — a string: the error’s message (every error comes with a message)
- `**stacktrace**` — a list of strings, can be empty: a stack trace or traceback of the error

