# Raftdb
Raftdb is a database using raft algorithm.

## Install
```
go get github.com/baboonwu/raftdb
```

## Run 
- start cmd
```
./raftdb --id 1 --cluster http://127.0.0.1:9001,http://127.0.0.1:9002,http://127.0.0.1:9003 --port 10001
./raftdb --id 2 --cluster http://127.0.0.1:9001,http://127.0.0.1:9002,http://127.0.0.1:9003 --port 10002
./raftdb --id 3 --cluster http://127.0.0.1:9001,http://127.0.0.1:9002,http://127.0.0.1:9003 --port 10003
```

- test cmd
```
curl -i -X PUT -H 'Content-Type: application/json' \\n    -d '{"key":"1", "val":"abcd"}' http://127.0.0.1:10001/data
curl -i -X PUT -H 'Content-Type: application/json' \\n    -d '{"key":"1", "val":"abcd"}' http://127.0.0.1:10002/data
curl -i -X PUT -H 'Content-Type: application/json' \\n    -d '{"key":"3", "val":"abcd"}' http://127.0.0.1:10001/data

curl -L http://127.0.0.1:10003/data/:key
```
