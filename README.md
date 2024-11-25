build/src/test/test_bitcoin --log_level=all --run_test=getarg_tests -- -printtoconsole=1debug.log1003Nov 27 22:45409628$ build/src/test/test_bitcoin --run_test=getarg_tests/doubledash -- -testdatadir=/somewhere/mydatadir
Test directory (will not be deleted): "/somewhere/mydatadir/test_common bitcoin/getarg_tests/doubledash/datadir"
Running 1 test case...

*** No errors detected
$ ls -l '/somewhere/mydatadir/test_common bitcoin/getarg_tests/doubledash/datadir'
total 8
drwxrwxr-x 2 admin admin 4096 Nov 27 22:45 blocks
-rw-rw-r-- 1 admin admin 1003 Nov 27 22:45 debug.log1gdb build/src/test/test_bitcoingdb build/src/test/test_bitcoin core

(gdb) bt  # produce a backtrace for where a segfault occurredijunxyz123:patch-2 <h1 align="center">
  <br>
  <a href="https://corecheck.dev"><img src="https://github.com/bitcoin-coverage/core/raw/master/docs/assets/logo.png" alt="Bitcoin Coverage" width="200"></a>
  <br>
    Bitcoin Coverage's Infra as Code
  <br>
</h1>

<h4 align="center">Bitcoin Coverage's infrastructure as code</h4>

## 📖 Introduction
This repository contains the infrastructure as code for Bitcoin Coverage. It is used to deploy all the components of the project.

## 🚀 CI/CD
The CI/CD is handled by GitHub Actions and is located in the `.github/workflows` folder. It is used to deploy the infrastructure on AWS on every push to the `master` branch.

## 📝 License

MIT - [Aurèle Oulès](https://github.com/aureleoules)
