# Offline AddOn lifecycle tests

Generate a harness containing the current production method bodies:

```sh
python3 ninjascript/tests/build_lifecycle_harness.py /tmp/lifecycle-harness.cs
```

Compile that generated C# as a console executable with a C# compiler and .NET Framework references (System, System.Core and mscorlib suffice), then run it offline. It uses only inert recording fakes: no NT8 libraries, accounts, connections or filesystem actions. Every successful assertion prints PASS; an assertion failure throws and exits nonzero. The generator intentionally fails when an expected production member is missing.

The separate NT8-reference compile must include all five production AddOn files and the installed NT8 Core/Gui/WPF reference assemblies; do not execute or deploy the resulting library. The repair report records exact temporary response-file paths used for the 2026-09-13 verification.

These checks exercise account/holding resolution and lifecycle handlers, including reentrant callbacks and ambiguous broker calls. They do not reproduce NT8 broker scheduling or replace an owner-authorized integration verification.

The independent lifecycle follow-up also defers or rejects fake Change requests, checks actual per-leg quantity confirmation without retry storms, and delivers both terminal child events synchronously inside Submit before mutable order states update. The 2026-09-13 follow-up records 33 passing assertions; wire serialization and Go residual-position handling remain outside this harness.
