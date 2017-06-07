#!/usr/bin/env bats


pushd .. > /dev/null
go build 1>&2
mv fin test/fin
popd > /dev/null

TESTCONFIG="--config /home/ben/.fin-test.toml"

setup() {
    ./fin $TESTCONFIG clear Transactions
    ./fin $TESTCONFIG ingest
}

# Note: to actually force bats to print out status and output for debugging
# tests insert
#   echo "status = ${status}"
#   echo "output = ${output}"
# into the test. Bats sets run output into the variable $output and also
# traps non-run output and will display it as a TAP comment.
# see https://github.com/sstephenson/bats/issues/191#issuecomment-256947900

@test "Ingest adds Uncategorized" {
    run ./fin $TESTCONFIG query --name GOOG

    [ "$status" -eq 0 ]
    [ "${lines[0]}" == "2017-4-22,GOOGLE *GOOGLE MUSIC 222 NS,-9.99,Entertainment" ]
    [ "${lines[1]}" == "2017-4-24,GOOGLE *GOOGLE MUSIC 101 NS,-9.99,Uncategorized" ]
}

@test "Apply bad category to data " {
    run ./fin $TESTCONFIG apply Entertain --expr ".*LATINA.*"

    [ "$status" -eq 1 ]
    [ "$output" == "fin: error: Category Entertain not found. Did you mean 'Entertainment'" ]
}

@test "Apply case-insensitive category to data" {
    ./fin $TESTCONFIG apply entertainment --expr ".*LATINA.*"
    run ./fin $TESTCONFIG query --expr ".*LATINA.*"

    [ "$status" -eq 0 ]
    [ "$output" == "2017-4-23,EPICERIE LATINA INC QC,-10.10,Entertainment" ]
}

@test "Setup actually resetting table. We should not see 'Entertainment'." {
    run ./fin $TESTCONFIG query --expr ".*LATINA.*"
    [ "$status" -eq 0 ]
    [ "$output" == "2017-4-23,EPICERIE LATINA INC QC,-10.10,Groceries" ]
}

@test "Query categories." {
    run ./fin $TESTCONFIG query --cat Uncategorized

    [ "$status" -eq 0 ]
    [ "${lines[0]}" == "2017-4-21,STATIONNEMENT DE MONTR QC,-5.00,Uncategorized" ]
    [ "${lines[1]}" == "2017-4-24,GOOGLE *GOOGLE MUSIC 101 NS,-9.99,Uncategorized" ]
}


@test "Ingesting does not overwrite existing categories" {
    ./fin $TESTCONFIG apply entertainment --expr ".*LATINA.*"
    ./fin $TESTCONFIG ingest
    run ./fin $TESTCONFIG query --expr ".*LATINA.*"

    [ "$status" -eq 0 ]
    [ "$output" == "2017-4-23,EPICERIE LATINA INC QC,-10.10,Entertainment" ]
}

@test "Internal and Google Place recommendations" {
    run ./fin $TESTCONFIG query --search

    [ "$status" -eq 0 ]
    [ "${lines[0]}" == "2017-4-24,GOOGLE *GOOGLE MUSIC 101 NS,-9.99,Entertainment" ]
    [ "${lines[1]}" == "2017-4-21,STATIONNEMENT DE MONTR QC,-5.00,Travel" ]
}

@test "Printing unmatched google place hits" {
    ./fin $TESTCONFIG apply Uncategorized --name EPICERIE
    run ./fin $TESTCONFIG query --places

    [ "$status" -eq 0 ]
    [ "$output" == "Name: EPICERIE LATINA INC QC Place: grocery_or_supermarket Record: 2017-4-23,EPICERIE LATINA INC QC,-10.10,Uncategorized" ]
}
