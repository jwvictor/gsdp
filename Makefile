
all: codegen
	go build

codegen: proto/gsdp.proto
	mkdir -p ../gsdprotocol
	protoc -I proto/ proto/gsdp.proto --go_out=plugins=grpc:.
	mv gsdp.pb.go ../gsdprotocol

clean:
	rm -rf ../gsdprotocol
