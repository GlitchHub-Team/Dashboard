<<<<<<< HEAD
#!/usr/bin/env bash
go test -coverpkg=./... -coverprofile=coverage.out ./tests/... -count=1 && go tool cover -html=coverage.out -o coverage.html
=======
if [ -z "$(find internal -maxdepth 1 -type d -ls)" ]; then
    echo ERRORE: Esegui questo script nella root del backend
    exit 1
fi

mkdir -p _coverage
echo "mode: set" > _coverage/coverage.out
echo "testedPackage,testPackage,coverage" > _coverage/coverage.csv
for Dir in $(find internal/* -maxdepth 10 -type d);
do
        if ls $Dir/*.go &> /dev/null;
        then
            TEST_DIR=./tests/${Dir//internal\//}

            if [ -d $TEST_DIR ]; then
                if [ "$(find $TEST_DIR -maxdepth 1 -type f -ls)" ]; then
                    testedPkg=backend/$Dir
                    echo "==== Testing package $testedPkg @ $TEST_DIR ===="
                    result=$(go test -coverpkg=$testedPkg -coverprofile=tmp.out $TEST_DIR -count=1)
                    testPkg=$(echo $result | grep -o -P '(?<=ok)\s*[\w/]+')
                    coverage=$(echo $result | grep -o -P '(?<=coverage:)\s*[\d]+\.[\d]+')
                    coverage=${coverage:-0.0}
                    echo $result
                    echo
                    echo "${testedPkg// /},${testPkg// /},${coverage// /}" >> _coverage/coverage.csv
                fi
            fi
            if [ -f tmp.out ]
            then
                cat tmp.out | grep -v "mode: set" >> _coverage/coverage.out
            fi
fi
done
rm tmp.out

go tool cover -html=_coverage/coverage.out -o _coverage/coverage.html

echo I file HTML e CSV si trovano nella cartella _coverage
>>>>>>> issue-95
