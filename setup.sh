#!/bin/bash

# Set path variables that are always necessary.
export GOPATH=$HOME/go
export PATH=$PATH:$HOME/go/bin
export GO_VERSION=1.16.2
export PROTOC_ZIP=protoc-3.14.0-linux-x86_64.zip

# Define a spinner.
spin() {
  spinner="/|\\-/|\\-"
  while :
  do
    for i in `seq 0 7`
    do
      echo -n "${spinner:$i:1}"
      echo -en "\010"
      sleep 1
    done
  done
}

# Download and install all necessary dependencies.
# TODO: download ffmpeg for cutting videos
setup() {
	echo "Output stored in setup.log"
	touch setup.log

	echo "Installing wget"
	sudo apt-get update > setup.log
	sudo apt-get upgrade -y > setup.log
	sudo apt-get install wget -y > setup.log

	echo "Installing golang"
	
	wget -c https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz > setup.log
	sudo tar -C $HOME -xzf go$GO_VERSION.linux-amd64.tar.gz
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

	echo "Done"
}

# Open a port to allow incoming traffic.
open_port() {
	read -p "Enter the port: " PORT
	read -p "Enter the VM instance name: " VM_INSTANCE
	gcloud compute firewall-rules create rule-allow-tcp-$PORT --source-ranges 0.0.0.0/0 --target-tags allow-tcp-$PORT --allow tcp:$PORT
	gcloud compute instances add-tags $VM_INSTANCE --tags allow-tcp-$PORT
	echo "Done"
}

# Start the datagatherer.
start_datagatherer() {
	echo "Starting datagatherer server."
	source env/ENV
	read -p "Enter GOOGLE_CLOUD_PROJECT: " GOOGLE_CLOUD_PROJECT
	read -p "Enter absolute path to GOOGLE_APPLICATION_CREDENTIALS json file: " GOOGLE_APPLICATION_CREDENTIALS
	cd cmd/datagatherer
	sudo env "PATH=$PATH" "GOPATH=$GOPATH" "GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT" "GOOGLE_APPLICATION_CREDENTIALS=$GOOGLE_APPLICATION_CREDENTIALS" go run .
}

if [ -z "$1" ]; then 
	echo "Must specify a command.  Options are [ setup, open_port, start_datagatherer ]."
	exit 1
fi

if [ $1 == "help" ]; then
	if [ -z "$2" ]; then
		echo "Options are [ setup, open_port, start_datagatherer ]. Type 'bash setup.sh help <command> to learn more."
	else
		if [ $2 == "setup" ]; then
			echo "Download and install all necessary dependencies."
		elif [ $2 == "open_port" ]; then 
			echo "Open a port to allow incoming traffic."
		elif [ $2 == "start_datagatherer" ]; then
			echo "Start the datagatherer."
		else 
			echo "Invalid argument."
		fi
	fi
	exit 1
fi

if [ $1 == "setup" ]; then
	# Start the spinner.
	spin &
	SPIN_PID=$!
	trap "kill -9 $SPIN_PID" `seq 0 15`
	# Run setup.
	setup
	kill -9 $SPIN_PID
elif [ $1 == "open_port" ]; then 
	open_port
elif [ $1 == "start_datagatherer" ]; then
	start_datagatherer
else
	echo "Invalid argument."
fi