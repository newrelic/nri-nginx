echo "===> Running unit tests"

go test ./src/...
if (-not $?)
{
    echo "Failed running tests"
    exit -1
}