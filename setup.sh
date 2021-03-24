echo "Output stored in setup.log"
touch setup.log

echo "Installing wget"
sudo apt-get update > setup.log
sudo apt-get upgrade -y > setup.log
sudo apt-get install wget -y > setup.log

echo "Installing golang"
export GO_VERSION=1.16.2
wget -c https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz > setup.log
sudo tar -C $HOME -xzf go$GO_VERSION.linux-amd64.tar.gz
export GOPATH=$HOME/go
export PATH=$PATH:$HOME/go/bin
rm go$GO_VERSION*

echo "Installing Exiftool"
sudo apt-get install build-essential -y > setup.log
wget -c https://exiftool.org/Image-ExifTool-12.22.tar.gz > setup.log
gzip -dc Image-ExifTool-12.22.tar.gz | tar -xf - > setup.log
cd Image-ExifTool-12.22 > setup.log
perl Makefile.PL > setup.log
make test > setup.log
sudo make install > setup.log
cd .. > setup.log
sudo rm -r Image-ExifTool-12.22 
rm Image-ExifTool-12.22.tar.gz 

echo "Installing unzip"
sudo apt-get install -y unzip > setup.log

echo "Installing protobuf compiler"
export PROTOC_ZIP=protoc-3.14.0-linux-x86_64.zip
curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/$PROTOC_ZIP > setup.log
sudo unzip -o $PROTOC_ZIP -d /usr/local bin/protoc > setup.log
sudo unzip -o $PROTOC_ZIP -d /usr/local 'include/*' > setup.log
rm -f $PROTOC_ZIP

echo "Generating protobuf golang code"
cd api/types
sudo env "PATH=$PATH" "GOPATH=$GOPATH" go get -u github.com/golang/protobuf/protoc-gen-go > ../../setup.log
sudo env "PATH=$PATH" "GOPATH=$GOPATH" protoc raw.proto --go_out=../../..
cd ../..

echo "Installing golang libraries"
sudo env "PATH=$PATH" "GOPATH=$GOPATH" go mod tidy

