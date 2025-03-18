## Testdata

We have various Dockerfiles copied from various opensource projects stored in this directory.
We use these before and after Dockerfile for various internal tests related to conversion etc.

### Get all test names

```sh
TESTNAMES="$(find testdata/ | grep '\.before\.' | xargs -L 1 basename | sed 's/\.before\..*//' | sort)"
```

Run something for each each test name (print them all with `echo`):
```sh
for NAME in $TESTNAMES; do echo $NAME; done
```

### Running dfc conversion on a test Dockerfile

```sh
go run . testdata/$NAME.before.Dockerfile
```

### Regenerating expected conversion outputs (after files)

Single (gcds-hugo):

```sh
NAME=gcds-hugo go run . --org chainguard-private testdata/$NAME.before.Dockerfile > testdata/$NAME.after.Dockerfile
```

All:

```sh
for NAME in $TESTNAMES; do go run . --org chainguard testdata/$NAME.before.Dockerfile > testdata/$NAME.after.Dockerfile; done
```

### Build the before Dockerfile

For the original version of the Dockerfile (gcds-hugo):
```sh
NAME=gcds-hugo WORKDIR=$([ -d testdata/$NAME ] && echo testdata/$NAME || echo .) && ( \
  set -x; docker build -t dfc-$NAME-before:dev -f testdata/$NAME.before.Dockerfile $WORKDIR)
```

### Build the after Dockerfile

For the original version of the Dockerfile after dfc conversion applied (or expected):

```sh
NAME=django WORKDIR=$([ -d testdata/$NAME ] && echo testdata/$NAME || echo .) && ( \
  set -x; docker build -t dfc-$NAME-after:dev -f testdata/$NAME.after.Dockerfile $WORKDIR)
```

### Create a new case from a file on disk

```sh
NAME=my-image cp path/to/Dockerfile testdata/$NAME.before.Dockerfile && ( \
  set -x; go run . testdata/$NAME.before.Dockerfile > testdata/$NAME.after.Dockerfile)
```
