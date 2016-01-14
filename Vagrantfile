Vagrant.configure(2) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.network "forwarded_port", guest: 443, host: 8443
  config.vm.network "private_network", ip: "192.168.33.10"
  config.vm.provision "shell", inline: <<-SHELL
    sed -i -e 's/ archive/ nl.archive/' /etc/apt/sources.list
    apt-get update
    apt-get upgrade -y
    if [[ ! -f /etc/apt/sources.list.d/docker.list ]] ; then
        apt-key adv --keyserver hkp://pgp.mit.edu:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
        echo "deb https://apt.dockerproject.org/repo ubuntu-trusty main" > /etc/apt/sources.list.d/docker.list
        apt-get update
        apt-get install -y docker-engine
    fi

    apt-get install -y git
    apt-get -y autoremove

    if [[ ! -d go1.5.3 ]] ; then
        echo "43afe0c5017e502630b1aea4d44b8a7f059bf60d7f29dfd58db454d4e4e0ae53  go1.5.3.linux-amd64.tar.gz" > go1.5.3.linux-amd64.tar.gz.sha256
        wget https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz
        shasum -c go1.5.3.linux-amd64.tar.gz.sha256
        tar xzvf go1.5.3.linux-amd64.tar.gz
        mv go go1.5.3
    fi

    if ! grep -q GOPATH ~/.bashrc ; then
cat >> .bashrc <<"EOF"

export EDITOR=/usr/bin/vim

# Go settings
export GOPATH=~/gopath
export GOROOT=~/go1.5.3
export PATH=$GOROOT/bin:$PATH

# git aliases
alias gs='git status'
alias gca='git commit --amend'
EOF
    fi

    if [[ ! -d ~/gopath/src/github.com/eBay/fabio/.git ]] ; then
        mkdir -p gopath/src/github.com/eBay/fabio
        git clone https://github.com/eBay/fabio.git gopath/src/github.com/eBay/fabio
        ( cd gopath/src/github.com/eBay/fabio && git remote set-url origin git@github.com:eBay/fabio )
    fi

    sudo -i -u vagrant git config --global push.default simple
    sudo -i -u vagrant git config --global user.email "frschroeder@ebay.com"
    sudo -i -u vagrant git config --global user.name "Frank Schroeder"

    chown -R vagrant:vagrant /home/vagrant
    echo "Done"
  SHELL
end
