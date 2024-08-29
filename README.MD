# Testing running Server CMD
go run main.go &

# Call server CMD
## Get Home
curl localhost:9000/api/sim

## Start runner
curl localhost:9000/api/sim/test/cf62d31e25ba3743e357d93ad3669ff2b8f944bd/commit

## Get result
curl localhost:9000/api/sim/test/{testResultId}/result

# Always running server
## Ubuntu
Add linked copy of sim_test_server.sh to boot apps

## Windows
Add linked copy of sim_test_server.bat to boot apps