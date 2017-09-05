### Glossary of Terms

#### Event
All data passing through the Serverless Platform is in the form of events. 

An Event has the following properties:
- id
- namespace/name - fully qualified name of the event type (kind). There will probably be a few platform defined, like errors
- timestamp
- source, which can be
  - function invocation (ref) - if the event's payload is a function return value/object (if any)
  - direct - if the event was created directly using the API
- data - the event payload, can be empty

#### Function
Functions are the application code that is deployed onto the platform by developers (the platform primary users). 

A Function has the following properties:
- id (unique, even across versions of the same entity)
- namespace/name
- version
- base image (ref) 
- input schema (ref)
- output schema (ref)
- impl - (TBD) reference to the function implementation (source code, JAR, etc.)

#### Base Image
Base Image is a container image into which the function code gets injected at deploy time. 

A Base Image has the following properties: 
- id (unique, even across versions of the same entity)
- namespace/name
- version
- ... TBD

#### Invocation
An Invocation is a record that represents a single invocation of a Function on the serverless platform. 

An Invocation has the following properties:
- id
- function (ref)
- timestamp
- input event (ref) - the source event payload is the function argument
- state records - an ordered (by timestamp) list of the invocation states filled as the invocation progresses.
  A single state record contains:
  - timestamp
  - state - one of the following (incomplete) list:
    - `validation-input`
    - input-error:
      - `input-invalid`
    - `running`
    - `validation-output`
    - runtime-error:
      - `internal-error`
      - `output-invalid`
      - `timed-out`
    - `success`
- output event (ref) - on the function termination (successful or not), ref to the output (result/error) event

#### Schema
A schema is a formal definition of a type of data objects to be used as event payloads. 

A schema has the following properties:
- id (unique, even across versions of the same entity)
- namespace/name
- version
- data - the schema data (can be represented in JSON or YAML)
