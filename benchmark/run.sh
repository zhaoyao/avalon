#! /bin/sh

r=$1
c=$2

set -e

pid_of_port() {
    port=$1
    lsof -i tcp:4320 | tail -1 | awk '{print $2}'
}

kill_port() {
    port=$1
    pid=$(pid_of_port $port)
    if [[ -n $pid ]]; then
        echo "Kill $pid"
        kill -9 $pid
    fi
}

kill_port 4320
kill_port 4321

# start hello world server at port 4321
cd server
go run tcp.go &
cd ..

# start proxy
cd ..
go clean
go build
./avalon --backend 127.0.0.1:4321 --port 4320 & 
cd benchmark

# run benchmark agaist bare server
cd client
go run main.go -r $r -c $c tcp://127.0.0.1:4320

# run benchmark agaist proxy server
go run main.go -r $r -c $c tcp://127.0.0.1:4321
