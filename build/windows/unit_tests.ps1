echo "===> Install unit tests dependencies"

go get github.com/axw/gocov/gocov github.com/AlekSi/gocov-xml
go get -v -d -t ./...

echo "===> Running unit tests"

gocov test ./src/...
if (-not $?)
{
    echo "Failed running tests"
    exit -1
}