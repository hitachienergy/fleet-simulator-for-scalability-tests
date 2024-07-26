#/bin/bash

git clone https://github.com/thingsboard/performance-tests.git

cd performance-tests
sh build.sh

rm -rf performance-tests
