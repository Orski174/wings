[![Logo Image](https://cdn.pterodactyl.io/logos/new/pterodactyl_logo.png)](https://pterodactyl.io)

![Discord](https://img.shields.io/discord/122900397965705216?label=Discord&logo=Discord&logoColor=white)
![GitHub Releases](https://img.shields.io/github/downloads/pterodactyl/wings/latest/total)
[![Go Report Card](https://goreportcard.com/badge/github.com/pterodactyl/wings)](https://goreportcard.com/report/github.com/pterodactyl/wings)

# Pterodactyl Wings

Wings is Pterodactyl's server control plane, built for the rapidly changing gaming industry and designed to be
highly performant and secure. Wings provides an HTTP API allowing you to interface directly with running server
instances, fetch server logs, generate backups, and control all aspects of the server lifecycle.

In addition, Wings ships with a built-in SFTP server allowing your system to remain free of Pterodactyl specific
dependencies, and allowing users to authenticate with the same credentials they would normally use to access the Panel.

## Sponsors

I would like to extend my sincere thanks to the following sponsors for helping fund Pterodactyl's development.
[Interested in becoming a sponsor?](https://github.com/sponsors/pterodactyl)

| Company                                                                           | About                                                                                                                                                                                                                                           |
|-----------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [**Aussie Server Hosts**](https://aussieserverhosts.com/)                         | No frills Australian Owned and operated High Performance Server hosting for some of the most demanding games serving Australia and New Zealand.                                                                                                 |
| [**BisectHosting**](https://www.bisecthosting.com/)                               | BisectHosting provides Minecraft, Valheim and other server hosting services with the highest reliability and lightning fast support since 2012.                                                                                                 |
| [**MineStrator**](https://minestrator.com/)                                       | Looking for the most highend French hosting company for your minecraft server? More than 24,000 members on our discord trust us. Give us a try!                                                                                                 |
| [**HostEZ**](https://hostez.io)                                                   | US & EU Rust & Minecraft Hosting. DDoS Protected bare metal, VPS and colocation with low latency, high uptime and maximum availability. EZ!                                                                                                     |
| [**Blueprint**](https://blueprint.zip/?utm_source=pterodactyl&utm_medium=sponsor) | Create and install Pterodactyl addons and themes with the growing Blueprint framework - the package-manager for Pterodactyl. Use multiple modifications at once without worrying about conflicts and make use of the large extension ecosystem. |
| [**indifferent broccoli**](https://indifferentbroccoli.com/)                      | indifferent broccoli is a game server hosting and rental company. With us, you get top-notch computer power for your gaming sessions. We destroy lag, latency, and complexity--letting you focus on the fun stuff.                              |

## Documentation

* [Panel Documentation](https://pterodactyl.io/panel/1.0/getting_started.html)
* [Wings Documentation](https://pterodactyl.io/wings/1.0/installing.html)
* [Kubernetes Provider Documentation](docs/k8s-provider.md) - **New!** Run Wings in Kubernetes clusters
* [Community Guides](https://pterodactyl.io/community/about.html)
* Or, get additional help [via Discord](https://discord.gg/pterodactyl)

## Runtime Providers

Wings now supports multiple runtime providers for managing game servers:

### Docker (Default)
The traditional Docker-based runtime provider for running game servers as containers.

**Requirements:**
- Docker Engine installed and running
- Access to Docker socket

### Kubernetes (Beta)
Run Wings inside a Kubernetes cluster and manage game servers as Pods instead of Docker containers. Ideal for Kubernetes-native environments like Talos Linux.

**Key Features:**
- Pod-based server management (1 Pod per server)
- Persistent storage via PersistentVolumeClaims
- Service-based port exposure (NodePort/LoadBalancer)
- Native Kubernetes resource management
- RBAC with least privilege

**Getting Started:**
```yaml
# Configure Wings to use Kubernetes
runtime:
  provider: kubernetes
  
  kubernetes:
    namespace: pterodactyl-servers
    service_type: NodePort
    storage_class: standard
```

See [docs/k8s-provider.md](docs/k8s-provider.md) for detailed setup instructions and [k8s/manifests/](k8s/manifests/) for deployment manifests.

## Reporting Issues

Please use the [pterodactyl/panel](https://github.com/pterodactyl/panel) repository to report any issues or make
feature requests for Wings. In addition, the [security policy](https://github.com/pterodactyl/panel/security/policy) listed
within that repository also applies to Wings.
