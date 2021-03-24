# seneca

Requirements:
* [Exiftool](https://exiftool.org/install.html#Unix) must be installed on the server for the datagatherer server to work
* There are many setup steps, like downloading go requirements and generating protobuf files, that are not yet documented here.  They will soon be automated in a Makefile.

### Setup from a fresh VM
1. Get the repository
    1. `sudo apt-get install git -y`
    1. `git clone https://github.com/Seneca-AI/seneca.git`
    1. `cd seneca`
1. Run the setup script
    1. `./setup.sh`
