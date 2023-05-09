wget https://dl.google.com/go/go1.20.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.20.linux-amd64.tar.gz
sudo echo -e "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
source /etc/profile
go version

