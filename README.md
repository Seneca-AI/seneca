# seneca

## Dev Commands
* Run all tests (from root repo directory)
    * `$ go test ./...`

### Requirements
* [Exiftool](https://exiftool.org/install.html#Unix) must be installed on the server
* [Protoc](https://grpc.io/docs/protoc-installation/)

## Setup

### Setup from a fresh VM
1. Get the repository
    1. `$ sudo apt-get install git -y`
    1. `$ git clone https://github.com/Seneca-AI/seneca.git`
    1. `$ cd seneca`
1. Run the setup script
    1. `$ bash setup.sh setup`

### Open up VM port for external traffic
1. `$ bash setup.sh open_port`

### Start the datagatherer server
1. `$ bash setup.sh start_datagatherer`
