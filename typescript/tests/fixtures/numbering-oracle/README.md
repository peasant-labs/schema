# Numbering oracle

The `src/` directory is a verbatim capture of the pre-change Fairtrade
transcript adapter and its transitive imports. It is the independent numbering
oracle used to verify that server-side cooking preserves the existing reader's
membership and labels. The source repository, commit and file digests are pinned
in the adjacent `numbering-oracle.yaml`; the replay checks every captured file
against that inventory before loading it.

The capture is excluded from the separately owned stringly-typed (`unstring`)
gate: editing it to satisfy that gate would change the oracle rather than
improve the production code. This repository does not yet configure that gate;
its separate adoption must preserve this exception. This README is
documentation, not part of the captured source.
