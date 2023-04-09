## Project structure

The go code is placed in two directories:
- **cmd**: Where we define the `main` entrypoint. At this moment we have a single `main`, `cmd/kube-audit-rest/main.go`
(we are generating a single executable) but this structure allows us to add other `main` programs in the future
- **internal**: Where we have the components that build our application

Within `internal`, we have a folder for each component. The only exception is `common`,
which contains some functions, singletons and constants that are used across most of the components.

Each directory within `internal` (except `common`), has a file called `i<component-name>.go`,
which defines the interface of a component. Then, we have a directory for each implementation of this interface.
Let's see `audit_writer` as an example.

```
iaudit_writer.go  # defines the interface AuditWriter
disk_writer
  |- disk_writer.go  # implementation of AuditWriter. It writes stuff on a file on disk
  \_ disk_writer_test.go  # unittests for disk_writer.go
stderr_writer
  |- stderr_writer.go  # implementation of AuditWriter. It write stuff on stderr
  \_ stderr_writer_test.go  # unittests for stderr_writer.go
common_writer
  \_ common_writer.go  # common functions that can be used in all the AuditWriter implementations
```

## Application architecture

This project follows a modular architecture. Components can only talk between them using the component interface.
They should never use anything specific of the implementation of such interface. The advantages are:

- We can easily change the implementation of a component without affecting the other components.
- Cyclic dependencies are not a problem. With this layout, the implementation of interface A could use interface
B and the implementation of interface B could use interface A.
- We can generate mocks for each interface, so we can unittest all the components.

## Graph of dependencies

```
     ┌─────────────┐
     │http_listener│
     └──────┬──────┘
            │
    ┌───────▼───────┐
    │event_processor│
    └─┬───────────┬─┘
      │           │
┌─────▼─┐   ┌─────▼──────┐
│metrics│   │audit_writer│
└───────┘   └────────────┘
```

## Unittests

Each implementation of an interface has unittests. To help isolating the component that we are testing
we need to mock the other components that it interacts with. The mock generation is done using the
`mockgen` package. This generation process is automated using the following approach.

We write something similar to the following code snippet in each file that defines a component
interface, like `iaudit_writer.go`

```go
package <component>

//go:generate mockgen -package mymock -destination ../../mocks/<component>_mock.go github.com/RichardoC/kube-audit-rest/internal/<current-dir> <interface-a>,<interface-b>
```

Then, the following command regenerates all the mocks.

```
go generate ./...
```

As usual, the execution of the unittests is done by

```
go test ./...
```
