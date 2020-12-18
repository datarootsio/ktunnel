<p align="center">
  <a href="" rel="noopener">
 <img width=200px height=100px src="./ktunnel-logo/cover.png" alt="Ktunnel logo"></a>
</p>

<h3 align="center">ktunnel</h3>

<div align="center">

  [![Status](https://img.shields.io/badge/status-active-success.svg)]() 
  [![GitHub Issues](https://img.shields.io/github/issues/omrikiei/ktunnel.svg)](https://github.com/omrikiei/ktunnel/issues)
  [![GitHub Pull Requests](https://img.shields.io/github/issues-pr/omrikiei/ktunnel.svg)](https://github.com/omrikiei/ktunnel/pulls)
  [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

</div>

---

<p align="center">Expose your local resources to kubernetes
    <br> 
</p>

## 📝 Table of Contents
- [About](#about)
- [Getting Started](#getting_started)
- [Usage](#usage)
- [Documentation](./docs/ktunnel.md)
- [Contributing](../CONTRIBUTING.md)
- [Authors](#authors)
- [Acknowledgments](#acknowledgement)

## 🧐 About <a name = "about"></a>
Ktunnel is a CLI tool that establishes a reverse tunnel between a kubernetes cluster and your local machine.
It lets you expose your machine as a service in the cluster or expose it to a specific deployment. 
You can also use the client and server without the orchestration part.
*Although ktunnel is identified with kubernetes, it can also be used as a reverse tunnel on any other remote system*

Ktunnel was born out of the need to access my development host when running applications on kubernetes. 
The aim of this project is to be a holistic solution to this specific problem (accessing the local machine from a kubernetes pod).
If you found this tool to be helpful on other scenarios, or have any suggesstions for new features - I would love to get in touch.

<p align="center">
<img src="./docs/request_sequence.png" alt="Ktunnel schema">
</p>

<p align="center">
<img src="./docs/ktunnel diagram.png" alt="Ktunnel schema">
</p>

## 🏁 Getting Started <a name = "getting_started"></a>
These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See [deployment](#deployment) for notes on how to deploy the project on a live system.

### Installation
#### Homebrew ####
```
brew tap omrikiei/ktunnel
brew install ktunnel
```
#### From the releases page
Download [here](https://github.com/omrikiei/ktunnel/releases/) and extract it to a local bin path
#### Building from source
Clone the project
```
git clone https://github.com/omrikiei/ktunnel; cd ktunnel
```
Build the binary
```
CGO_ENABLED=0 go build -ldflags="-s -w"
```
You can them move it to your bin path
```
sudo mv ./ktunnel /usr/local/bin/ktunnel
```
Test the commamd
```
ktunnel -h
```

## 🎈 Usage <a name="usage"></a>
### Expose your local machine as a service in the cluster
This will allow pods in the cluster to access your local web app (listening on port 8000) via 
http (i.e kubernetes applications can send requests to myapp:8000)
```bash
ktunnel expose myapp 80:8000
```

### Inject to an existing deployment
This will currently only work for deployments with 1 replica - it will expose a listening port on the pod through a tunnel to your local machine
```bash
ktunnel inject deployment mydeployment 3306
``` 

## ✍️ Authors <a name = "authors"></a>
- [@omrikiei](https://github.com/omrikiei)

See also the list of [contributors](https://github.com/omrikiei/ktunnel/contributors) who participated in this project.
