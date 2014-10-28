
Obtaining the code
------------------

    go get github.com/niemeyer/flex

Running the tool
----------------

    cd $GOPATH/src/github.com/niemeyer/flex
    cd cmd/flex
    go build

    # FLEX_DIR defaults to /var/lib/flex and holds the unix socket.
    export FLEX_DIR=$PWD

    # On one terminal, run the daemon:
    ./flex daemon --debug

    # On another terminal, ping it:
    ./flex ping --debug
    ./flex create c1
    ./flex list
    ./flex start c1
    ./flex attach c1
    # play in the root@c1 shell, then exit
    ./flex stop c1

Running tests
-------------

    cd $GOPATH/src/github.com/niemeyer/flex
    go test
