default:
	go build 

grpc:
	buf generate

clean:
	rm -rf ./wt
	rm -rf gen

