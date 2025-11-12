
# vmware-exporter

This is a simple prometheus exporter that collects various metrics from a vCenter. 


## How to use

Run the exporter in a docker container (or start as a process) with all the settings necessary. Scrape it..

Exporter scrapes single vCenter host when /metrics path is used. Multiple vCenter hosts can be scraped using /probe, however these vCenter hosts must share credentials.

### Settings 

The exporter can be configured via command line options, environment variables, a yaml config file or a combination of all three. The environment variables set will be overwritten by the contents of the config file, which then will be overwritten by any command line option set at startup. 
The options available are:

| key | description |
| --- | ----------- |
| -envflag.enable | Tells the exporter to use enviromnent flags in its configuration |
| -envflag.prefix | This allows to prefix the environment variables that will be used for configuration | 
| -file | Path to a yaml configuration file that follows the structure of command line options |
| -http.address | The address and port the exporter will bind to in host:port format (default: ":9169") |
| -log.format | Can be either json or logfmt (default: logfmt) |
| -log.level | One of debug,info,warn or error (default: debug) - Don't expect much..|
| -prom.maxRequests | Max concurrent scrape requests (default: 20) |
| -disable.exporter.metrics | Disables exporter process metrics |
| -disable.exporter.target | Disables exporter default target - /metrics will only return exporter data - use /probe |
| -disable.default.collectors | Disables all collectors enabled by default |
| -collector.datacenter | Enables or disables DataCenter metrics collection (default: enabled) |
| -collector.cluster | Enables or disables Cluster metrics collection (default: enabled) |
| -collector.datastore | Enables or disables Datastore metrics collection (default: enabled) |
| -collector.host | Enables or disables Host metrics collection (default: enabled) |
| -collector.vm | Enables or disables Virtual Machine metrics collection (default: enabled) |
| -collector.esxcli.host.nic | Collects ESXi NIC firmware information using esxcli invoked through the vCenter (default: disabled) |
| -collector.esxcli.storage | Collects ESXi storage firmware information using esxcli invoked through the vCenter (default: disabled) |
| -vmware.granularity | The frequency of the sampled data. Default is 20s (default 20) |
| -vmware.insecureTLS | Trust insecure vCenter TLS (true) or verify (default) |
| -vmware.interval | How often data will be collected. Default is every 20s. (default 20) |
| -vmware.password | Password for the user above |
| -vmware.schema | Use HTTP or HTTPS (default "https") |
| -vmware.username | Username to login to vCenter server |
| -vmware.vcenter | vCenter server address in host:port format. This is not the vCenter Management Console |


The esxcli collectors are a very specific use case that probably is not going to be needed by anyone. Left the code in here as an example on how custom information can be collected using esxcli command tool remotely via vCenter SOAP API 