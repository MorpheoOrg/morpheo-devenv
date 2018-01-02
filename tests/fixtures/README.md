# Fixtures

This directory holds a very light set of 4 data, 2 algos and 2 problems to perform quick testing on the Morpheo Platform.

The **fastest** algo and problem are very light docker images (< 3MB) written in Golang. These scripts mainly perform copies of predefined fixture files.

The **pytest** algo and problem are ~260MB docker images written in Python. Images are heavier, because it performs real operations on hdf5 files (exactly like `hypnogram-wf`).

### Makefile Usage
```
make train         : Build algo and run task *train*
     pred          : Build algo and run task *predict*
     detarget      : Build problem and run task *detarget*
     perf          : Build problem and run task *perf*
     tar-gz        : Generate tar-gz archive of algo and problem
     clean         : Clean all previous command outputs
     gen-fixtures  : Generate fixtures for tests, and place them in morpheo-devenv/data
     register-algo : Register the test algo to the orchestrator, cleaning previous tests
```

To check that it's working, you can run `make clean detarget train perf pred` and see the files created in `/data`.

To use `make register-algo`, you need to:
* Set `kubectl` to interact with your cluster
* Set the orchestrator's authentication `user/pass` as environment variables `USER_AUTH`/`PWD_AUTH`. (ex: `export PWD_AUTH='pass/word'`. Note that quotes `'` wrapping the password can be necessary here, as sometimes characters could be cropped without them...)


### Improvements compared to previous testing datasets

| Data                    | Size    |
|-------------------------|---------|
| new train and test      | 16.6 KB |
| previous train and test | 1.6 GB  |

| Docker Images    | Size   |
|------------------|--------|
| fastest          | 2.5 MB |
| pytest           | 258 MB |
| example_sklearn  | 337 MB |
| example_keras    | 833 MB |
