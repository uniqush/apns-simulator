# apns-simulator (archived)

**This repository is archived and unmaintained. It is kept for history only.**

It simulates [APNs][apns-binary] over the **binary provider protocol**, which
Apple switched off on **31 March 2021**. Nothing it accepts on the wire will
reach a device, and nothing that talks to real APNs today can be tested against
it. The last change here was in 2018; there is no `go.mod`, so it predates Go
modules and will not build with a modern toolchain without one.

## Use the simulator in uniqush-push instead

[`srv/apns/apnstest`][apnstest] in [uniqush-push][uniqush-push] is an APNs
simulator for the **HTTP/2 provider API**, the one Apple still serves. It is a
Go package rather than a separate process, so a test starts it in-process on a
random port:

```go
server := apnstest.NewServer()
defer server.Close()

// Point the provider at server.URL(), then assert on what arrived.
for _, r := range server.Requests() { ... }
for _, v := range server.Violations() { ... }
```

It answers what Apple documents, and it *rejects* what Apple rejects — a missing
`apns-push-type`, a background push at priority 10, a topic header sent twice
under two spellings — recording each as a conformance violation, so a failing
test says which rule was broken rather than only that the push failed. It also
validates `.p8` provider tokens, models the token refresh interval with an
injectable clock, and counts open connections. It does not deliver to a device
and never authenticates a client certificate.

This closes the two open issues that were filed here:

- [#2][issue2], scriptable library for integration testing — `SetResponse` makes
  the response depend on the device token, `Requests()` is the call log, and
  `ActiveConnections()` reports connection lifecycle.
- [#3][issue3], HTTP/2 simulation with a local certificate —
  `apnstest.NewServer` generates a self-signed certificate, and
  `WriteCACert` writes it out so the client under test can verify the chain
  rather than skip verification.

It lives in uniqush-push rather than here on purpose. The simulator's value is
that it disagrees with uniqush when uniqush is wrong, so it is developed in the
same commit as the behaviour it pins; splitting it across two repositories would
put a release boundary in the middle of that loop. Being in-tree does not make
it private — it is an ordinary importable package:

```
import "github.com/uniqush/uniqush-push/srv/apns/apnstest"
```

If you need a *standalone* simulator process — because the client you are
testing is not written in Go — open an issue on
[uniqush-push][uniqush-push-issues]. Wrapping `apnstest` in a command is small;
it has not been done because no one has asked.

[apns-binary]: https://developer.apple.com/documentation/usernotifications/setting-up-a-remote-notification-server
[apnstest]: https://github.com/uniqush/uniqush-push/tree/master/srv/apns/apnstest
[uniqush-push]: https://github.com/uniqush/uniqush-push
[uniqush-push-issues]: https://github.com/uniqush/uniqush-push/issues
[issue2]: https://github.com/uniqush/apns-simulator/issues/2
[issue3]: https://github.com/uniqush/apns-simulator/issues/3
