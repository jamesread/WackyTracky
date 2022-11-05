default:
	go build github.com/wacky-tracky/wacky-tracky-server/cmd/wt/

grpc:
	buf generate

clean:
	rm -rf ./wt
	rm -rf gen

