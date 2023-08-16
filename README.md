# Output directory containing the formatted manuscript

The [`gh-pages`](https://github.com/uiceds/syllabus/tree/gh-pages) branch hosts the contents of this directory at <https://uiceds.github.io/syllabus/>.
The permalink for this webpage version is <https://uiceds.github.io/syllabus/v/ee6b9a3b05909b8f18aec3a12c3bd9fab1b537c9/>.
To redirect to the permalink for the latest manuscript version at anytime, use the link <https://uiceds.github.io/syllabus/v/freeze/>.

## Files

This directory contains the following files, which are mostly ignored on the `main` branch:

+ [`index.html`](index.html) is an HTML manuscript.
+ [`manuscript.pdf`](manuscript.pdf) is a PDF manuscript.

The `v` directory contains directories for each manuscript version.
In general, a version is identified by the commit hash of the source content that created it.

### Timestamps

The `*.ots` files in version directories are OpenTimestamps which can be used to verify manuscript existence at or before a given time.
[OpenTimestamps](https://opentimestamps.org/) uses the Bitcoin blockchain to attest to file hash existence.
The `deploy.sh` script run during continuous deployment creates the `.ots` files through its `manubot webpage` call.
There is a delay before timestamps get confirmed by a Bitcoin block.
Therefore, `.ots` files are initially incomplete and should be upgraded at a later time, so that they no longer rely on the availability of a calendar server to verify.
The `manubot webpage` call during continuous deployment identifies files matched by `webpage/v/**/*.ots` and attempts to upgrade them.
You can also manually upgrade timestamps, by running the following in the `gh-pages` branch:

```shell
ots upgrade v/*/*.ots
rm v/*/*.ots.bak
git add v/*/*.ots
```

Verifying timestamps with the `ots verify` command requires running a local bitcoin node with JSON-RPC configured, at this time.

## Source

The manuscripts in this directory were built from
[`ee6b9a3b05909b8f18aec3a12c3bd9fab1b537c9`](https://github.com/uiceds/syllabus/commit/ee6b9a3b05909b8f18aec3a12c3bd9fab1b537c9).
